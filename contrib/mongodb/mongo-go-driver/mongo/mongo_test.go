package mongo

import (
	"context"
	"fmt"
	"github.com/signalfx/signalfx-go-tracing/contrib/internal/testutil"
	"io"
	"net"
	"testing"
	"time"

	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/mocktracer"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"github.com/signalfx/signalfx-go-tracing/internal/globalconfig"
	"github.com/signalfx/signalfx-go-tracing/tracing"
	"github.com/signalfx/signalfx-go-tracing/zipkinserver"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/network/result"
	"go.mongodb.org/mongo-driver/x/network/wiremessage"
)

func Test(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	li, err := mockMongo()
	if err != nil {
		t.Fatal(err)
	}

	hostname, port, _ := net.SplitHostPort(li.Addr().String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	span, ctx := tracer.StartSpanFromContext(ctx, "mongodb-test")

	addr := fmt.Sprintf("mongodb://%s", li.Addr().String())
	opts := options.Client()
	opts.SetMonitor(NewMonitor())
	client, err := mongo.Connect(ctx, opts.ApplyURI(addr))
	if err != nil {
		t.Fatal(err)
	}

	client.
		Database("test-database").
		Collection("test-collection").
		InsertOne(ctx, bson.D{{Key: "test-item", Value: "test-value"}})

	span.Finish()

	spans := mt.FinishedSpans()
	assert.Len(t, spans, 2)
	assert.Equal(t, spans[0].TraceID(), spans[1].TraceID())

	s := spans[0]
	assert.Equal(t, "mongo", s.Tag(ext.ServiceName))
	assert.Equal(t, "mongo.insert", s.Tag(ext.ResourceName))
	assert.Equal(t, hostname, s.Tag(ext.PeerHostname))
	assert.Equal(t, port, s.Tag(ext.PeerPort))
	assert.Contains(t, s.Tag(ext.DBStatement), `"test-item":"test-value"`)
	assert.Equal(t, "test-database", s.Tag(ext.DBInstance))
	assert.Equal(t, "mongo", s.Tag(ext.DBType))
}

func TestWithZipkin(t *testing.T) {
	assert := assert.New(t)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	tracing.Start(tracing.WithEndpointURL(zipkin.URL()), tracing.WithServiceName("test-mongo-service"))
	defer tracing.Stop()

	li, err := mockMongo()
	if err != nil {
		t.Fatal(err)
	}

	hostname, port, _ := net.SplitHostPort(li.Addr().String())

	addr := fmt.Sprintf("mongodb://%s", li.Addr().String())
	opts := options.Client()
	opts.SetMonitor(NewMonitor())
	client, err := mongo.Connect(context.Background(), opts.ApplyURI(addr))

	assert.NotNil(client)
	assert.Nil(err)

	t.Run("test insert", func(t *testing.T) {
		zipkin.Reset()

		client.
			Database("test-database").
			Collection("test-collection").
			InsertOne(context.Background(), bson.D{{Key: "test-item", Value: "test-value"}})

		tracer.ForceFlush()

		spans := zipkin.WaitForSpans(t, 1)

		span := spans[0]
		if assert.NotNil(span.LocalEndpoint.ServiceName) {
			assert.Equal("test-mongo-service", *span.LocalEndpoint.ServiceName)
		}

		assert.Equal("mongo.insert", *span.Name)
		assert.Equal(ext.SpanKindClient, *span.Kind)

		assert.Equal("mongodb", span.Tags["component"])
		assert.Equal(hostname, span.Tags[ext.PeerHostname])
		assert.Equal(port, span.Tags[ext.PeerPort])
		assert.Equal("test-database", span.Tags[ext.DBInstance])
		assert.Equal("mongo", span.Tags[ext.DBType])
		assert.Contains(span.Tags[ext.DBStatement], `"test-item":"test-value"`)

		testutil.AssertSpanWithNoError(t, span)
	})
}

// mockMongo implements a crude mongodb server that responds with
// expected replies so that we can confirm tracing works properly
func mockMongo() (net.Listener, error) {
	li, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return li, err
	}

	go func() {
		var requestID int32
		nextRequestID := func() int32 {
			requestID++
			return requestID
		}
		defer li.Close()
		for {
			conn, err := li.Accept()
			if err != nil {
				break
			}
			go func() {
				defer conn.Close()

				for {
					var hdrbuf [16]byte
					_, err := io.ReadFull(conn, hdrbuf[:])
					if err != nil {
						panic(err)
					}

					hdr, err := wiremessage.ReadHeader(hdrbuf[:], 0)
					if err != nil {
						panic(err)
					}

					msgbuf := make([]byte, hdr.MessageLength)
					copy(msgbuf, hdrbuf[:])
					_, err = io.ReadFull(conn, msgbuf[16:])
					if err != nil {
						panic(err)
					}

					switch hdr.OpCode {
					case wiremessage.OpQuery:
						var msg wiremessage.Query
						err = msg.UnmarshalWireMessage(msgbuf)
						if err != nil {
							panic(err)
						}

						bs, _ := bson.Marshal(result.IsMaster{
							IsMaster:                     true,
							OK:                           1,
							MaxBSONObjectSize:            16777216,
							MaxMessageSizeBytes:          48000000,
							MaxWriteBatchSize:            100000,
							LogicalSessionTimeoutMinutes: 30,
							ReadOnly:                     false,
							MinWireVersion:               0,
							MaxWireVersion:               7,
						})

						reply := wiremessage.Reply{
							MsgHeader: wiremessage.Header{
								RequestID:  nextRequestID(),
								ResponseTo: hdr.RequestID,
							},
							ResponseFlags:  wiremessage.AwaitCapable,
							NumberReturned: 1,
							Documents:      []bson.Raw{bs},
						}
						bs, err = reply.MarshalWireMessage()
						if err != nil {
							panic(err)
						}

						_, err = conn.Write(bs)
						if err != nil {
							panic(err)
						}

					case wiremessage.OpMsg:
						var msg wiremessage.Msg
						err = msg.UnmarshalWireMessage(msgbuf)
						if err != nil {
							panic(err)
						}
						d := bson.D{{Key: "n", Value: 1}, {Key: "ok", Value: 1}}
						bs, _ := bson.Marshal(d)

						bs, _ = wiremessage.Msg{
							MsgHeader: wiremessage.Header{
								RequestID:  nextRequestID(),
								ResponseTo: hdr.RequestID,
							},
							Sections: []wiremessage.Section{
								&wiremessage.SectionBody{
									Document: bs,
								},
							},
						}.MarshalWireMessage()

						_, err = conn.Write(bs)
						if err != nil {
							panic(err)
						}

					default:
						panic("unknown op code: " + hdr.OpCode.String())
					}

				}
			}()
		}
	}()

	return li, nil
}

func TestAnalyticsSettings(t *testing.T) {
	assertRate := func(t *testing.T, mt mocktracer.Tracer, rate interface{}, opts ...Option) {
		li, err := mockMongo()
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		addr := fmt.Sprintf("mongodb://%s", li.Addr().String())
		mongopts := options.Client()
		mongopts.SetMonitor(NewMonitor(opts...))
		client, err := mongo.Connect(ctx, mongopts.ApplyURI(addr))
		if err != nil {
			t.Fatal(err)
		}
		client.
			Database("test-database").
			Collection("test-collection").
			InsertOne(ctx, bson.D{{Key: "test-item", Value: "test-value"}})

		spans := mt.FinishedSpans()
		assert.Len(t, spans, 1)
		s := spans[0]
		assert.Equal(t, rate, s.Tag(ext.EventSampleRate))
	}

	t.Run("defaults", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		assertRate(t, mt, nil)
	})

	t.Run("global", func(t *testing.T) {
		t.Skip("global flag disabled")
		mt := mocktracer.Start()
		defer mt.Stop()

		rate := globalconfig.AnalyticsRate()
		defer globalconfig.SetAnalyticsRate(rate)
		globalconfig.SetAnalyticsRate(0.4)

		assertRate(t, mt, 0.4)
	})

	t.Run("enabled", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		assertRate(t, mt, 1.0, WithAnalytics(true))
	})

	t.Run("disabled", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		assertRate(t, mt, nil, WithAnalytics(false))
	})

	t.Run("override", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		rate := globalconfig.AnalyticsRate()
		defer globalconfig.SetAnalyticsRate(rate)
		globalconfig.SetAnalyticsRate(0.4)

		assertRate(t, mt, 0.23, WithAnalyticsRate(0.23))
	})
}

package http

import (
	"github.com/signalfx/signalfx-go-tracing/contrib/internal/testutil"
	"github.com/signalfx/signalfx-go-tracing/ddtrace"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/mocktracer"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"github.com/signalfx/signalfx-go-tracing/internal/globalconfig"
	"github.com/signalfx/signalfx-go-tracing/tracing"
	"github.com/signalfx/signalfx-go-tracing/zipkinserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoundTripperZipkin(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	tracing.Start(tracing.WithEndpointURL(zipkin.URL()), tracing.WithServiceName("test-http-service"))
	defer tracing.Stop()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spanctx, err := tracer.Extract(tracer.HTTPHeadersCarrier(r.Header))
		if err != nil {
			// Can't call test functions from different goroutine.
			panic("inject failed: " + err.Error())
		}

		span := tracer.StartSpan("test-span",
			tracer.ChildOf(spanctx))
		defer span.Finish()

		switch r.URL.Query().Get("error") {
		case "400":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Bad Request"))
			return
		case "500":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal Server Error"))
			return
		}

		_, _ = w.Write([]byte("Hello World"))
	}))
	defer s.Close()

	rt := WrapRoundTripper(http.DefaultTransport)

	client := &http.Client{
		Transport: rt,
	}

	t.Run("successful request", func(t *testing.T) {
		zipkin.Reset()
		_, err := client.Get(s.URL + "/query")
		require.NoError(err)

		tracer.ForceFlush()

		spans := zipkin.Spans
		require.Len(spans, 2)

		s1 := spans[1]
		if assert.NotNil(s1.LocalEndpoint.ServiceName) {
			assert.Equal("test-http-service", *s1.LocalEndpoint.ServiceName)
		}

		assert.Equal(map[string]string{
			"component":        "http",
			"http.method":      "GET",
			"http.url":         s.URL + "/query",
			"span.kind":        "client",
			"http.status_code": "200",
		}, s1.Tags)
		assert.Len(s1.Annotations, 0)
	})

	t.Run("500 error", func(t *testing.T) {
		zipkin.Reset()

		_, err := client.Post(s.URL+"/?error=500", "text", strings.NewReader(""))
		require.NoError(err)

		tracer.ForceFlush()

		spans := zipkin.Spans
		require.Len(spans, 2)

		s1 := spans[1]
		if assert.NotNil(s1.LocalEndpoint.ServiceName) {
			assert.Equal("test-http-service", *s1.LocalEndpoint.ServiceName)
		}
		tags := s1.Tags

		assert.Equal("http.request", *s1.Name)
		assert.Equal(map[string]string{
			"component":        "http",
			"error":            "true",
			"http.method":      "POST",
			"http.url":         s.URL + "/?error=500",
			"span.kind":        "client",
			"http.status_code": "500",
		}, tags)

		ann := testutil.GetAnnotations(t, s1)

		assert.Equal(tags["error"], "true")
		assert.Len(ann, 0)
	})

	t.Run("host connect error", func(t *testing.T) {
		zipkin.Reset()

		_, err := client.Get("http://localhost:1/query")
		require.Error(err)

		tracer.ForceFlush()

		spans := zipkin.Spans
		require.Len(spans, 1)

		s0 := spans[0]
		if assert.NotNil(s0.LocalEndpoint.ServiceName) {
			assert.Equal("test-http-service", *s0.LocalEndpoint.ServiceName)
		}
		tags := s0.Tags

		assert.Equal("http.request", *s0.Name)
		assert.Equal(map[string]string{
			"component":   "http",
			"error":       "true",
			"http.method": "GET",
			"http.url":    "http://localhost:1/query",
			"span.kind":   "client",
		}, tags)

		ann := testutil.GetAnnotations(t, s0)

		assert.Equal(tags["error"], "true")
		assert.Len(ann, 5)
		assert.Equal(ann["event"], "error")
		assert.Contains(ann["message"], "connection refused")
		assert.Greater(len(ann["stack"]), 50)
		assert.Contains(ann["stack"], "goroutine")
		assert.Equal(ann["error.kind"], "*net.OpError")
		assert.Contains(ann["error.object"], "&net.OpError{")
	})
}

func TestRoundTripper(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spanctx, err := tracer.Extract(tracer.HTTPHeadersCarrier(r.Header))
		assert.NoError(t, err)

		span := tracer.StartSpan("test",
			tracer.ChildOf(spanctx))
		defer span.Finish()

		w.Write([]byte("Hello World"))
	}))
	defer s.Close()

	rt := WrapRoundTripper(http.DefaultTransport,
		WithBefore(func(req *http.Request, span ddtrace.Span) {
			span.SetTag("CalledBefore", true)
		}),
		WithAfter(func(res *http.Response, span ddtrace.Span) {
			span.SetTag("CalledAfter", true)
		}))

	client := &http.Client{
		Transport: rt,
	}

	client.Get(s.URL + "/hello/world")

	spans := mt.FinishedSpans()
	assert.Len(t, spans, 2)
	assert.Equal(t, spans[0].TraceID(), spans[1].TraceID())

	s0 := spans[0]
	assert.Equal(t, "test", s0.OperationName())
	assert.Equal(t, "test", s0.Tag(ext.ResourceName))

	s1 := spans[1]
	assert.Equal(t, "http.request", s1.OperationName())
	assert.Equal(t, "http.request", s1.Tag(ext.ResourceName))
	assert.Equal(t, "200", s1.Tag(ext.HTTPCode))
	assert.Equal(t, "GET", s1.Tag(ext.HTTPMethod))
	assert.Equal(t, s.URL + "/hello/world", s1.Tag(ext.HTTPURL))
	assert.Equal(t, true, s1.Tag("CalledBefore"))
	assert.Equal(t, true, s1.Tag("CalledAfter"))
}

func TestWrapClient(t *testing.T) {
	c := WrapClient(http.DefaultClient)
	assert.Equal(t, c, http.DefaultClient)
	_, ok := c.Transport.(*roundTripper)
	assert.True(t, ok)
}

func TestRoundTripperAnalyticsSettings(t *testing.T) {
	assertRate := func(t *testing.T, mt mocktracer.Tracer, rate interface{}, opts ...RoundTripperOption) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		rt := WrapRoundTripper(http.DefaultTransport, opts...)

		client := &http.Client{Transport: rt}
		client.Get(srv.URL + "/hello/world")
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

		assertRate(t, mt, 1.0, RTWithAnalytics(true))
	})

	t.Run("disabled", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		assertRate(t, mt, nil, RTWithAnalytics(false))
	})

	t.Run("override", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		rate := globalconfig.AnalyticsRate()
		defer globalconfig.SetAnalyticsRate(rate)
		globalconfig.SetAnalyticsRate(0.4)

		assertRate(t, mt, 0.23, RTWithAnalyticsRate(0.23))
	})
}

package tracing

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"github.com/signalfx/signalfx-go-tracing/zipkinserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_span_LogFields(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	Start(WithEndpointURL(zipkin.URL()))

	span := opentracing.GlobalTracer().StartSpan("test-span")
	span.LogFields(log.Int("int", 5), log.String("str", "value"))
	span.LogFields(log.Bool("bool", true))
	span.Finish()

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 1)

	annotations := spans[0].Annotations
	require.Len(annotations, 2)

	assert.Equal(`{"int":5,"str":"value"}`, *annotations[0].Value)
	assert.Equal(`{"bool":true}`, *annotations[1].Value)
}


func TestOpenTracingChildSpan(t *testing.T) {
	assert := assert.New(t)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	Start(WithEndpointURL(zipkin.URL()))

	span0, _ := tracer.StartSpanFromContext(context.Background(), "span-1")
	span1 := opentracing.StartSpan("span-2", opentracing.ChildOf(span0.Context()))

	span1.Finish()
	span0.Finish()

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 2)

	assert.Equal(spans[0].TraceID, spans[1].TraceID)
	assert.Equal(spans[0].ID, *spans[1].ParentID)
}

func TestOpenTracingParentSpan(t *testing.T) {
	assert := assert.New(t)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	Start(WithEndpointURL(zipkin.URL()))

	span0, ctxt := opentracing.StartSpanFromContext(context.Background(), "span-0")
	span1, _ := tracer.StartSpanFromContext(ctxt, "span-1")

	span1.Finish()
	span0.Finish()

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 2)

	assert.Equal(spans[0].TraceID, spans[1].TraceID)
	assert.Equal(spans[0].ID, *spans[1].ParentID)
}

func TestEmptyContext(t *testing.T) {
	assert := assert.New(t)
	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	Start(WithEndpointURL(zipkin.URL()))

	sessionName := "session"
	span, _ := opentracing.StartSpanFromContext(context.Background(), sessionName)
	span.Finish()

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 1)
	assert.Equal(sessionName, *spans[0].Name)
}

func TestWithGlobalTags(t *testing.T) {
	require := require.New(t)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	Start(WithEndpointURL(zipkin.URL()),
		WithGlobalTag("abc-test", "1234"),
		WithGlobalTag("test", "value"))

	span := tracer.StartSpan("test")
	span.Finish()

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 1)

	tags := spans[0].Tags
	require.Equal(4, len(tags))
	assert.Equal(t, "value",tags["test"], )
	assert.Equal(t, "1234", tags["abc-test"])
	assert.Equal(t, signalfxVersionValue, tags[signalfxVersionKey])
	assert.Equal(t, signalfxLibraryValue,tags[signalfxLibraryKey])
}

func TestEnvironmentVariables(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// There will be no "  " key with value "silentjay" as Zipkin states empty keys are not valid.
	err := os.Setenv(signalfxSpanTags, "a:b, c :d  , bob:,  : silentjay")
	require.Nil(err)
	defer os.Unsetenv(signalfxSpanTags)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	Start(WithServiceName("MyService"),
		WithEndpointURL(zipkin.URL()),
		WithGlobalTag("abc-test", "1234"),
		WithGlobalTag("test", "value"))

	span := tracer.StartSpan("test")
	span.Finish()

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 1)

	assert.Equal("MyService", *spans[0].LocalEndpoint.ServiceName)

	tags := spans[0].Tags
	require.Equal(7, len(tags))
	assert.Equal("value",tags["test"])
	assert.Equal("1234", tags["abc-test"])
	assert.Equal("b", tags["a"])
	assert.Equal("d", tags["c"])
	assert.Equal("", tags["bob"])
	assert.Equal(signalfxVersionValue, tags[signalfxVersionKey])
	assert.Equal(signalfxLibraryValue,tags[signalfxLibraryKey])
}

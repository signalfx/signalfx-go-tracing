package opentracer

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"github.com/signalfx/signalfx-go-tracing/tracing"
	"github.com/signalfx/signalfx-go-tracing/zipkinserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_span_LogFields(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	ot := New(tracing.WithEndpointURL(zipkin.URL()))
	opentracing.SetGlobalTracer(ot)

	span := opentracing.GlobalTracer().StartSpan("test-span")
	span.LogFields(log.Int("int", 5), log.String("str", "value"))
	span.LogFields(log.Bool("bool", true))
	span.Finish()

	tracer.ForceFlush()

	require.Len(zipkin.Spans, 1)
	annotations := zipkin.Spans[0].Annotations
	require.Len(annotations, 2)

	assert.Equal(`{"int":5,"str":"value"}`, *annotations[0].Value)
	assert.Equal(`{"bool":true}`, *annotations[1].Value)
}

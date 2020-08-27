package tracing

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	traceformat "github.com/signalfx/golib/trace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sfxtracing "github.com/signalfx/signalfx-go-tracing"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"github.com/signalfx/signalfx-go-tracing/zipkinserver"
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

func TestLibraryMetadata(t *testing.T) {
	require := require.New(t)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	Start(WithEndpointURL(zipkin.URL()))

	span := tracer.StartSpan("root")
	{
		span2 := tracer.StartSpan("inner", tracer.ChildOf(span.Context()))
		{
			span3 := tracer.StartSpan("inner-most", tracer.ChildOf(span2.Context()))
			span3.Finish()
		}
		span2.Finish()
	}
	span.Finish()

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 3)

	rootSpan := spans[0]
	require.Equal("root", *rootSpan.Name)
	require.Equal(2, len(rootSpan.Tags))
	require.Equal(sfxtracing.LibraryName, rootSpan.Tags[ext.SFXTracingLibrary])
	require.Equal(sfxtracing.Version, rootSpan.Tags[ext.SFXTracingVersion])

	for _, span := range spans[1:] {
		require.Equal(0, len(span.Tags))
	}
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
	require.Equal("value", tags["test"])
	require.Equal("1234", tags["abc-test"])
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
	assert.Equal("value", tags["test"])
	assert.Equal("1234", tags["abc-test"])
	assert.Equal("b", tags["a"])
	assert.Equal("d", tags["c"])
	assert.Equal("", tags["bob"])
}

func TestWithRecordedValueMaxLength(t *testing.T) {
	require := require.New(t)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	maxLength := 5
	Start(WithEndpointURL(zipkin.URL()),
		WithRecordedValueMaxLength(maxLength),
		WithoutLibraryTags())

	span := tracer.StartSpan("test")
	tagVal := "vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv"
	require.Len(tagVal, 31)

	span.SetTag("k1", tagVal)
	span.SetTag("k2", "v2")
	span.LogFields(log.String("l1", tagVal), log.String("l2", "v2"))
	span.Finish()

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 1)

	tags := spans[0].Tags
	require.Equal(2, len(tags))
	require.Equal("v2", tags["k2"])
	require.Equal(tagVal[:maxLength], tags["k1"])

	annotations := spans[0].Annotations
	require.Equal(1, len(annotations))
	logs := annotationToMap(t, annotations[0])
	require.Equal("v2", logs["l2"])
	require.Equal(tagVal[:maxLength], logs["l1"])
}

func TestWithRecordedValueMaxLengthSpanOverride(t *testing.T) {
	require := require.New(t)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	Start(WithEndpointURL(zipkin.URL()),
		WithRecordedValueMaxLength(5),
		WithoutLibraryTags())

	span := tracer.StartSpan("test", tracer.WithRecordedValueMaxLength(10))
	tagVal := "vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv"
	require.Len(tagVal, 31)

	span.SetTag("k1", tagVal)
	span.SetTag("k2", "v2")
	span.LogFields(log.String("l1", tagVal), log.String("l2", "v2"))
	span.Finish()

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 1)

	tags := spans[0].Tags
	require.Equal(2, len(tags))
	require.Equal("v2", tags["k2"])
	require.Equal(tagVal[:10], tags["k1"])

	annotations := spans[0].Annotations
	require.Equal(1, len(annotations))
	logs := annotationToMap(t, annotations[0])
	require.Equal("v2", logs["l2"])
	require.Equal(tagVal[:10], logs["l1"])
}

func TestWithRecordedValueMaxLengthDefault(t *testing.T) {
	require := require.New(t)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	Start(WithEndpointURL(zipkin.URL()), WithoutLibraryTags())

	span := tracer.StartSpan("test")
	v1 := strings.Repeat("v", 1300)
	span.SetTag("k1", v1)

	v2 := strings.Repeat("v", 1199)
	span.SetTag("k2", v2)
	span.LogFields(log.String("l1", v1), log.String("l2", v2))
	span.Finish()

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 1)

	tags := spans[0].Tags
	require.Equal(2, len(tags))
	require.Equal(v1[:defaultRecordedValueMaxLength], tags["k1"])
	require.Equal(v2, tags["k2"])

	annotations := spans[0].Annotations
	require.Equal(1, len(annotations))
	logs := annotationToMap(t, annotations[0])
	require.Equal(v1[:defaultRecordedValueMaxLength], logs["l1"])
	require.Equal(v2, logs["l2"])
}

func annotationToMap(t *testing.T, annotation *traceformat.Annotation) map[string]string {
	var m map[string]string

	if err := json.Unmarshal([]byte(*annotation.Value), &m); err != nil {
		t.Error(err)
	}
	return m
}

func TestDropTagsFromEnv(t *testing.T) {
	require := require.New(t)

	err := os.Setenv(signalfxDropTags, "ttt, T2,")
	require.Nil(err)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	Start(WithEndpointURL(zipkin.URL()),
		WithRecordedValueMaxLength(5),
		WithoutLibraryTags(),
	)

	span := tracer.StartSpan("test", tracer.WithRecordedValueMaxLength(10))

	span.SetTag("t1", "v1")
	span.SetTag("t2", "v2")
	span.SetTag("tt", "v3")
	span.SetTag("ttt", "v4")
	span.Finish()

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 1)

	tags := spans[0].Tags
	require.Equal(2, len(tags))
	require.Equal("v1", tags["t1"])
	require.Equal("v3", tags["tt"])
}

func TestDropTagsFromOption(t *testing.T) {
	require := require.New(t)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	Start(WithEndpointURL(zipkin.URL()),
		WithRecordedValueMaxLength(5),
		WithoutLibraryTags(),
		WithDropTags("t1", "t2", "TT"),
	)

	span := tracer.StartSpan("test", tracer.WithRecordedValueMaxLength(10))

	span.SetTag("t1", "v1")
	span.SetTag("t2", "v2")
	span.SetTag("tt", "v3")
	span.SetTag("ttt", "v4")
	span.Finish()

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 1)

	tags := spans[0].Tags
	require.Equal(1, len(tags))
	require.Equal("v4", tags["ttt"])
}

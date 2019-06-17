package http

import (
	"github.com/signalfx/signalfx-go-tracing/contrib/internal/testutil"
	"github.com/signalfx/signalfx-go-tracing/tracing"
	"github.com/signalfx/signalfx-go-tracing/zipkinserver"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/mocktracer"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"github.com/signalfx/signalfx-go-tracing/internal/globalconfig"
	"github.com/stretchr/testify/assert"
)

func TestHttpTracer200Zipkin(t *testing.T) {
	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	tracing.Start(tracing.WithEndpointURL(zipkin.URL()), tracing.WithServiceName("test-http-service"))
	defer tracing.Stop()

	url := "/200"
	r := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	router().ServeHTTP(w, r)

	assert := assert.New(t)
	require := require.New(t)
	assert.Equal(200, w.Code)
	assert.Equal("OK\n", w.Body.String())

	tracer.ForceFlush()
	spans := zipkin.Spans()
	require.Len(spans, 1)
	span := spans[0]

	assert.Equal("SERVER", *span.Kind)
	assert.Equal("/200", *span.Name)
	assert.Equal(map[string]string{
		"component":        "web",
		"foo":              "bar",
		"http.url":         "/200",
		"http.method":      "GET",
		"http.status_code": "200",
		"span.kind":        "server",
	}, span.Tags)
}

func TestHttpTracer500Zipkin(t *testing.T) {
	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	tracing.Start(tracing.WithEndpointURL(zipkin.URL()), tracing.WithServiceName("test-http-service"))
	defer tracing.Stop()

	// Send and verify a 500 request
	url := "/500"
	r := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	router().ServeHTTP(w, r)

	assert := assert.New(t)
	require := require.New(t)
	assert.Equal(500, w.Code)
	assert.Equal("500!\n", w.Body.String())

	tracer.ForceFlush()
	spans := zipkin.Spans()
	require.Len(spans, 1)
	span := spans[0]

	assert.Equal("SERVER", *span.Kind)
	assert.Equal("/500", *span.Name)
	assert.Equal(map[string]string{
		"component":        "web",
		"foo":              "bar",
		"http.url":         "/500",
		"http.method":      "GET",
		"http.status_code": "500",
		"span.kind":        "server",
		"error":            "true",
	}, span.Tags)

	require.Len(span.Annotations, 1)

	ann := testutil.GetAnnotation(t, span, 0)
	assert.Equal(ann["event"], "error")
	assert.Contains(ann["message"], "500: Internal Server Error")
	assert.Greater(len(ann["stack"]), 50)
	assert.Contains(ann["stack"], "goroutine")
	assert.Equal(ann["error.kind"], "*errors.errorString")
	assert.Contains(ann["error.object"], "&errors.errorString")
}

func TestHttpTracer200(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	url := "/200"
	r := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	router().ServeHTTP(w, r)

	assert := assert.New(t)
	assert.Equal(200, w.Code)
	assert.Equal("OK\n", w.Body.String())

	spans := mt.FinishedSpans()
	assert.Equal(1, len(spans))

	s := spans[0]
	assert.Equal("http.request", s.OperationName())
	assert.Equal("my-service", s.Tag(ext.ServiceName))
	assert.Equal(url, s.Tag(ext.ResourceName))
	assert.Equal("200", s.Tag(ext.HTTPCode))
	assert.Equal("GET", s.Tag(ext.HTTPMethod))
	assert.Equal(url, s.Tag(ext.HTTPURL))
	assert.Equal(nil, s.Tag(ext.Error))
	assert.Equal("bar", s.Tag("foo"))
}

func TestHttpTracer500(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	// Send and verify a 500 request
	url := "/500"
	r := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	router().ServeHTTP(w, r)

	assert := assert.New(t)
	assert.Equal(500, w.Code)
	assert.Equal("500!\n", w.Body.String())

	spans := mt.FinishedSpans()
	assert.Equal(1, len(spans))

	s := spans[0]
	assert.Equal("http.request", s.OperationName())
	assert.Equal("my-service", s.Tag(ext.ServiceName))
	assert.Equal(url, s.Tag(ext.ResourceName))
	assert.Equal("500", s.Tag(ext.HTTPCode))
	assert.Equal("GET", s.Tag(ext.HTTPMethod))
	assert.Equal(url, s.Tag(ext.HTTPURL))
	assert.Equal("500: Internal Server Error", s.Tag(ext.Error).(error).Error())
	assert.Equal("bar", s.Tag("foo"))
}

func TestWrapHandler200(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()
	assert := assert.New(t)

	handler := WrapHandler(http.HandlerFunc(handler200), "my-service", "my-resource",
		WithSpanOptions(tracer.Tag("foo", "bar")))

	url := "/"
	r := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(200, w.Code)
	assert.Equal("OK\n", w.Body.String())

	spans := mt.FinishedSpans()
	assert.Equal(1, len(spans))

	s := spans[0]
	assert.Equal("http.request", s.OperationName())
	assert.Equal("my-service", s.Tag(ext.ServiceName))
	assert.Equal("my-resource", s.Tag(ext.ResourceName))
	assert.Equal("200", s.Tag(ext.HTTPCode))
	assert.Equal("GET", s.Tag(ext.HTTPMethod))
	assert.Equal(url, s.Tag(ext.HTTPURL))
	assert.Equal(nil, s.Tag(ext.Error))
	assert.Equal("bar", s.Tag("foo"))
}

func TestAnalyticsSettings(t *testing.T) {
	assertRate := func(t *testing.T, mt mocktracer.Tracer, rate interface{}, opts ...Option) {
		mux := NewServeMux(opts...)
		mux.HandleFunc("/200", handler200)
		r := httptest.NewRequest("GET", "/200", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

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

func router() http.Handler {
	mux := NewServeMux(WithServiceName("my-service"), WithSpanOptions(tracer.Tag("foo", "bar")))
	mux.HandleFunc("/200", handler200)
	mux.HandleFunc("/500", handler500)
	return mux
}

func handler200(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK\n"))
}

func handler500(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "500!", http.StatusInternalServerError)
}

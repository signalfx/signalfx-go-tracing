package gin

import (
	"errors"
	"fmt"
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/signalfx/signalfx-go-tracing/contrib/internal/testutil"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/mocktracer"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"github.com/signalfx/signalfx-go-tracing/internal/globalconfig"
	"github.com/signalfx/signalfx-go-tracing/tracing"
	"github.com/signalfx/signalfx-go-tracing/zipkinserver"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.ReleaseMode) // silence annoying log msgs
}

func TestChildSpan(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()

	router := gin.New()
	router.Use(Middleware("foobar"))
	router.GET("/user/:id", func(c *gin.Context) {
		_, ok := tracer.SpanFromContext(c.Request.Context())
		assert.True(ok)
	})

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)
}

func TestTrace200(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()

	router := gin.New()
	router.Use(Middleware("foobar"))
	router.GET("/user/:id", func(c *gin.Context) {
		span, ok := tracer.SpanFromContext(c.Request.Context())
		assert.True(ok)
		assert.Equal(span.(mocktracer.Span).Tag(ext.ServiceName), "foobar")
		id := c.Param("id")
		c.Writer.Write([]byte(id))
	})

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()

	// do and verify the request
	router.ServeHTTP(w, r)
	response := w.Result()
	assert.Equal(response.StatusCode, 200)

	// verify traces look good
	spans := mt.FinishedSpans()
	assert.Len(spans, 1)
	if len(spans) < 1 {
		t.Fatalf("no spans")
	}
	span := spans[0]
	assert.Equal("/user/:id", span.OperationName())
	assert.Equal(ext.SpanTypeGin, span.Tag(ext.SpanType))
	assert.Equal("foobar", span.Tag(ext.ServiceName))
	assert.Contains(span.Tag(ext.ResourceName), "/user/:id")
	assert.Equal("200", span.Tag(ext.HTTPCode))
	assert.Equal("GET", span.Tag(ext.HTTPMethod))
	// TODO(x) would be much nicer to have "/user/:id" here
	assert.Equal("http://example.com/user/123", span.Tag(ext.HTTPURL))
}

func TestError(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()

	// setup
	router := gin.New()
	router.Use(Middleware("foobar"))
	wantErr := errors.New("oh no")

	// a handler with an error and make the requests
	router.GET("/err", func(c *gin.Context) {
		c.AbortWithError(500, wantErr)
	})
	r := httptest.NewRequest("GET", "/err", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	response := w.Result()
	assert.Equal(response.StatusCode, 500)

	// verify the errors and status are correct
	spans := mt.FinishedSpans()
	assert.Len(spans, 1)
	if len(spans) < 1 {
		t.Fatalf("no spans")
	}
	span := spans[0]
	assert.Equal("/err", span.OperationName())
	assert.Equal("foobar", span.Tag(ext.ServiceName))
	assert.Equal("500", span.Tag(ext.HTTPCode))
	assert.Equal(wantErr.Error(), span.Tag(ext.Error).(error).Error())
}

func TestHTML(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()

	// setup
	router := gin.New()
	router.Use(Middleware("foobar"))

	// add a template
	tmpl := template.Must(template.New("hello").Parse("hello {{.}}"))
	router.SetHTMLTemplate(tmpl)

	// a handler with an error and make the requests
	router.GET("/hello", func(c *gin.Context) {
		HTML(c, 200, "hello", "world")
	})
	r := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	response := w.Result()
	assert.Equal(response.StatusCode, 200)
	assert.Equal("hello world", w.Body.String())

	// verify the errors and status are correct
	spans := mt.FinishedSpans()
	assert.Len(spans, 2)
	for _, s := range spans {
		assert.Equal("foobar", s.Tag(ext.ServiceName), s.String())
	}

	var tspan mocktracer.Span
	for _, s := range spans {
		// we need to pick up the span we're searching for, as the
		// order is not garanteed within the buffer
		if s.OperationName() == "gin.render.html" {
			tspan = s
		}
	}
	assert.NotNil(tspan)
	assert.Equal("hello", tspan.Tag("go.template"))
}

func TestGetSpanNotInstrumented(t *testing.T) {
	assert := assert.New(t)
	router := gin.New()
	router.GET("/ping", func(c *gin.Context) {
		// Assert we don't have a span on the context.
		_, ok := tracer.SpanFromContext(c.Request.Context())
		assert.False(ok)
		c.Writer.Write([]byte("ok"))
	})
	r := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	response := w.Result()
	assert.Equal(response.StatusCode, 200)
}

func TestPropagation(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()

	pspan := tracer.StartSpan("test")
	tracer.Inject(pspan.Context(), tracer.HTTPHeadersCarrier(r.Header))

	router := gin.New()
	router.Use(Middleware("foobar"))
	router.GET("/user/:id", func(c *gin.Context) {
		span, ok := tracer.SpanFromContext(c.Request.Context())
		assert.True(ok)
		assert.Equal(span.(mocktracer.Span).ParentID(), pspan.(mocktracer.Span).SpanID())
	})

	router.ServeHTTP(w, r)
}

func TestAnalyticsSettings(t *testing.T) {
	assertRate := func(t *testing.T, mt mocktracer.Tracer, rate interface{}, opts ...Option) {
		router := gin.New()
		router.Use(Middleware("foobar", opts...))
		router.GET("/user/:id", func(_ *gin.Context) {})

		r := httptest.NewRequest("GET", "/user/123", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, r)

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

func TestWithZipkin(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	tracing.Start(tracing.WithEndpointURL(zipkin.URL()), tracing.WithServiceName("test-gin-service"))
	defer tracing.Stop()

	router := gin.New()
	router.Use(Middleware("foobar"))
	router.GET("/successful", func(c *gin.Context) {
		_, ok := tracer.SpanFromContext(c.Request.Context())
		assert.True(ok)
		c.Status(200)
	})
	router.POST("/unsuccessful", func(c *gin.Context) {
		_, ok := tracer.SpanFromContext(c.Request.Context())
		assert.True(ok)
		c.AbortWithError(400, fmt.Errorf("Gin Error"))
	})

	t.Run("successful request", func(t *testing.T) {
		zipkin.Reset()

		r := httptest.NewRequest("GET", "/successful", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		response := w.Result()
		assert.Equal(response.StatusCode, 200)

		tracer.ForceFlush()

		spans := zipkin.WaitForSpans(t, 1)
		require.Len(spans, 1)

		span := spans[0]

		assert.Equal(*span.Name, "/successful")
		assert.Equal(*span.Kind, "SERVER")
		assert.Equal(*span.LocalEndpoint.ServiceName, "test-gin-service")
		testutil.AssertSpanWithTags(t, span, map[string]string{
			"component":        "gin",
			"http.method":      "GET",
			"http.status_code": "200",
			"http.url":         "http://example.com/successful",
			"span.kind":        strings.ToLower(ext.SpanKindServer),
		})
		testutil.AssertSpanWithNoError(t, span)
	})

	t.Run("unsuccessful request", func(t *testing.T) {
		zipkin.Reset()

		r := httptest.NewRequest("POST", "/unsuccessful", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		response := w.Result()
		assert.Equal(400, response.StatusCode)

		tracer.ForceFlush()

		spans := zipkin.WaitForSpans(t, 1)

		span := spans[0]

		assert.Equal(*span.LocalEndpoint.ServiceName, "test-gin-service")
		assert.Equal(*span.Kind, "SERVER")
		assert.Equal(*span.Name, "/unsuccessful")
		testutil.AssertSpanWithTags(t, span, map[string]string{
			"component":        "gin",
			"http.method":      "POST",
			"http.status_code": "400",
			"http.url":         "http://example.com/unsuccessful",
			"span.kind":        strings.ToLower(ext.SpanKindServer),
		})
		testutil.AssertSpanWithError(t, span, testutil.ErrorAssertion{
			KindEquals:      "*gin.Error",
			MessageContains: "Gin Error",
			ObjectContains:  "&gin.Error{",
			StackContains:   []string{"goroutine"},
			StackMinLength:  50,
		})
	})
}

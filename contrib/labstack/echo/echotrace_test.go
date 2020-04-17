package echo

import (
	"errors"
	"github.com/signalfx/signalfx-go-tracing/contrib/internal/testutil"
	"github.com/signalfx/signalfx-go-tracing/tracing"
	"github.com/signalfx/signalfx-go-tracing/zipkinserver"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/mocktracer"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestChildSpan(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()
	var called, traced bool

	router := echo.New()
	router.Use(Middleware(WithServiceName("foobar")))
	router.GET("/user/:id", func(c echo.Context) error {
		called = true
		_, traced = tracer.SpanFromContext(c.Request().Context())
		return c.NoContent(200)
	})

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	// verify traces look good
	assert.True(called)
	assert.True(traced)
}

func TestTrace200(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()
	var called, traced bool

	router := echo.New()
	router.Use(Middleware(WithServiceName("foobar")))
	router.GET("/user/:id", func(c echo.Context) error {
		called = true
		var span tracer.Span
		span, traced = tracer.SpanFromContext(c.Request().Context())

		// we patch the span on the request context.
		span.SetTag("test.echo", "echony")
		assert.Equal(span.(mocktracer.Span).Tag(ext.ServiceName), "foobar")
		return c.NoContent(200)
	})

	root := tracer.StartSpan("root")
	r := httptest.NewRequest("GET", "/user/123", nil)
	err := tracer.Inject(root.Context(), tracer.HTTPHeadersCarrier(r.Header))
	assert.Nil(err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	// verify traces look good
	assert.True(called)
	assert.True(traced)

	spans := mt.FinishedSpans()
	assert.Len(spans, 1)

	span := spans[0]
	assert.Equal("/user/:id", span.OperationName())
	assert.Equal(ext.SpanTypeEcho, span.Tag(ext.SpanType))
	assert.Equal("foobar", span.Tag(ext.ServiceName))
	assert.Equal("echony", span.Tag("test.echo"))
	assert.Contains(span.Tag(ext.ResourceName), "/user/:id")
	assert.Equal("200", span.Tag(ext.HTTPCode))
	assert.Equal("GET", span.Tag(ext.HTTPMethod))
	//assert.Equal(root.Context().SpanID(), span.ParentID())

	assert.Equal("http://example.com/user/123", span.Tag(ext.HTTPURL))
}

func TestError(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()
	var called, traced bool

	// setup
	router := echo.New()
	router.Use(Middleware(WithServiceName("foobar")))
	wantErr := errors.New("oh no")

	// a handler with an error and make the requests
	router.GET("/err", func(c echo.Context) error {
		_, traced = tracer.SpanFromContext(c.Request().Context())
		called = true

		err := wantErr
		c.Error(err)
		return err
	})
	r := httptest.NewRequest("GET", "/err", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	// verify the errors and status are correct
	assert.True(called)
	assert.True(traced)

	spans := mt.FinishedSpans()
	assert.Len(spans, 1)

	span := spans[0]
	assert.Equal("/err", span.OperationName())
	assert.Equal("foobar", span.Tag(ext.ServiceName))
	assert.Equal("500", span.Tag(ext.HTTPCode))
	assert.Equal(wantErr.Error(), span.Tag(ext.Error).(error).Error())
}

func TestGetSpanNotInstrumented(t *testing.T) {
	assert := assert.New(t)
	router := echo.New()
	var called, traced bool

	router.GET("/ping", func(c echo.Context) error {
		// Assert we don't have a span on the context.
		called = true
		_, traced = tracer.SpanFromContext(c.Request().Context())
		return c.NoContent(200)
	})

	r := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)
	assert.True(called)
	assert.False(traced)
}

func TestEchoTracer200Zipkin(t *testing.T) {
	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	tracing.Start(tracing.WithEndpointURL(zipkin.URL()), tracing.WithServiceName("test-echo"))
	defer tracing.Stop()

	e := echo.New()
	e.Use(Middleware(WithServiceName("test-echo")))
	e.GET("/mock200", mock200)

	r := httptest.NewRequest("GET", "/mock200", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 1)
	span := spans[0]

	assert.Equal(t, *span.Name, "/mock200")
	assert.Equal(t, *span.Kind, "SERVER")
	testutil.AssertSpanWithTags(t,span, map[string]string{
		"component":        "echo",
		"http.url":         "http://example.com/mock200",
		"http.method":      "GET",
		"http.status_code": "200",
		"span.kind":        "server",
	})
	testutil.AssertSpanWithNoError(t, span)
}

func TestEchoTracer401Zipkin(t *testing.T) {
	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	tracing.Start(tracing.WithEndpointURL(zipkin.URL()), tracing.WithServiceName("test-echo"))
	defer tracing.Stop()

	e := echo.New()
	e.Use(Middleware(WithServiceName("test-echo")))
	e.POST("/mock401", mock401)

	r := httptest.NewRequest("POST", "/mock401", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 1)
	span := spans[0]

	assert.Equal(t, *span.Name, "/mock401")
	assert.Equal(t, *span.Kind, "SERVER")
	testutil.AssertSpanWithTags(t, span, map[string]string{
		"component":        "echo",
		"http.url":         "http://example.com/mock401",
		"http.method":      "POST",
		"http.status_code": "401",
		"span.kind":        "server",
	})
	testutil.AssertSpanWithNoError(t, span)
}

func TestEchoTracer500Zipkin(t *testing.T) {
	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	tracing.Start(tracing.WithEndpointURL(zipkin.URL()), tracing.WithServiceName("test-echo"))
	defer tracing.Stop()

	e := echo.New()
	e.Use(Middleware(WithServiceName("test-echo")))
	e.POST("/mock500", mock500)

	r := httptest.NewRequest("POST", "/mock500", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 1)
	span := spans[0]

	assert.Equal(t, *span.Kind, "SERVER")
	assert.Equal(t, *span.Name, "/mock500")

	testutil.AssertSpanWithTags(t, span, map[string]string{
		"component":        "echo",
		"http.url":         "http://example.com/mock500",
		"http.method":      "POST",
		"http.status_code": "500",
		"span.kind":        "server",
		ext.Error:            "true",
		ext.ErrorKind: "*echo.HTTPError",
	})

	testutil.AssertSpanWithError(t, span, testutil.ErrorAssertion{
		KindEquals:      "*echo.HTTPError",
		MessageContains: "Internal Server Error",
		ObjectContains:  "&echo.HTTPError",
		StackContains:   []string{"goroutine"},
		StackMinLength:  50,
	})
}

func mock200(c echo.Context) error {
	return c.JSON(http.StatusOK, "")
}

func mock401(c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, "")
}

func mock500(c echo.Context) error {
	err := &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Internal Server Error"}
	c.Error(err)
	return err
}

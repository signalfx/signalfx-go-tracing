package mux

import (
	"fmt"
	"github.com/signalfx/signalfx-go-tracing/contrib/internal/testutil"
	"github.com/signalfx/signalfx-go-tracing/tracing"
	"github.com/signalfx/signalfx-go-tracing/zipkinserver"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/mocktracer"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"github.com/signalfx/signalfx-go-tracing/internal/globalconfig"

	"github.com/stretchr/testify/assert"
)

func TestTracedGorillaMux(t *testing.T) {
	for _, ht := range []struct {
		code     int
		method   string
		url      string
		route    string
		errorStr string
	}{
		{
			code:   http.StatusOK,
			method: "GET",
			url:    "http://localhost/200",
			route:  "/200",
		},
		{
			code:   http.StatusNotFound,
			method: "GET",
			url:    "http://localhost/not_a_real_route",
			route:  "unknown",
		},
		{
			code:   http.StatusMethodNotAllowed,
			method: "POST",
			url:    "http://localhost/405",
			route:  "unknown",
		},
		{
			code:     http.StatusInternalServerError,
			method:   "GET",
			url:      "http://localhost/500",
			route:    "/500",
			errorStr: "500: Internal Server Error",
		},
	} {
		t.Run(http.StatusText(ht.code), func(t *testing.T) {
			assert := assert.New(t)
			mt := mocktracer.Start()
			defer mt.Stop()
			codeStr := strconv.Itoa(ht.code)

			// Send and verify a request
			r := httptest.NewRequest(ht.method, ht.url, nil)
			w := httptest.NewRecorder()
			router().ServeHTTP(w, r)
			assert.Equal(ht.code, w.Code)
			assert.Equal(codeStr+"!\n", w.Body.String())

			spans := mt.FinishedSpans()
			assert.Equal(1, len(spans))

			s := spans[0]
			assert.Equal("http.request", s.OperationName())
			assert.Equal("my-service", s.Tag(ext.ServiceName))
			assert.Equal(codeStr, s.Tag(ext.HTTPCode))
			assert.Equal(ht.method, s.Tag(ext.HTTPMethod))
			assert.Equal(ht.url, s.Tag(ext.HTTPURL))
			assert.Equal(ht.route, s.Tag(ext.ResourceName))
			if ht.errorStr != "" {
				assert.Equal(ht.errorStr, s.Tag(ext.Error).(error).Error())
			}
		})
	}
}

func TestDomain(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()
	mux := NewRouter(WithServiceName("my-service"))
	mux.Handle("/200", okHandler()).Host("localhost")
	r := httptest.NewRequest("GET", "http://localhost/200", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	spans := mt.FinishedSpans()
	assert.Equal(1, len(spans))
	assert.Equal("localhost", spans[0].Tag("mux.host"))
}

func TestSpanOptions(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()
	mux := NewRouter(WithSpanOptions(tracer.Tag(ext.SamplingPriority, 2)))
	mux.Handle("/200", okHandler()).Host("localhost")
	r := httptest.NewRequest("GET", "http://localhost/200", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	spans := mt.FinishedSpans()
	assert.Equal(1, len(spans))
	assert.Equal(2, spans[0].Tag(ext.SamplingPriority))
}

// TestImplementingMethods is a regression tests asserting that all the mux.Router methods
// returning the router will return the modified traced version of it and not the original
// router.
func TestImplementingMethods(t *testing.T) {
	r := NewRouter()
	_ = (*Router)(r.StrictSlash(false))
	_ = (*Router)(r.SkipClean(false))
	_ = (*Router)(r.UseEncodedPath())
}

func TestAnalyticsSettings(t *testing.T) {
	assertRate := func(t *testing.T, mt mocktracer.Tracer, rate interface{}, opts ...RouterOption) {
		mux := NewRouter(opts...)
		mux.Handle("/200", okHandler()).Host("localhost")
		r := httptest.NewRequest("GET", "http://localhost/200", nil)
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

func TestTracedGorillaMuxToZipKinServerHTTP200(t *testing.T) {
	testTracedGorillaMuxHelper(t, "GET", "200", 200, "/200")
}

func TestTracedGorillaMuxToZipKinServerHTTP404(t *testing.T) {
	testTracedGorillaMuxHelper(t, "GET", "404", 404, "unknown")
}

func TestTracedGorillaMuxToZipKinServerHTTP405(t *testing.T) {
	testTracedGorillaMuxHelper(t, "POST", "405", 405, "unknown")
}

func TestTracedGorillaMuxToZipKinServerHTTP500(t *testing.T) {
	testTracedGorillaMuxHelper(t, "GET", "500", 500, "/500")
}

func testTracedGorillaMuxHelper(t *testing.T, httpMethod string, path string, wantHTTPCode int, wantSpanName string) {
	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	tracing.Start(tracing.WithEndpointURL(zipkin.URL()), tracing.WithServiceName("test-mux-service"))
	defer tracing.Stop()

	url := "/" + path
	r := httptest.NewRequest(httpMethod, url, nil)
	w := httptest.NewRecorder()
	router().ServeHTTP(w, r)

	assert := assert.New(t)
	gotHTTPCode := w.Code
	assert.Equal(wantHTTPCode, gotHTTPCode)
	wantHTTPCodeStr := strconv.Itoa(wantHTTPCode)
	assert.Equal(wantHTTPCodeStr+ "!\n", w.Body.String())

	tracer.ForceFlush()
	gotSpans := zipkin.WaitForSpans(t, 1)
	gotSpan := gotSpans[0]

	wantSpan := map[string]interface{}{
		"kind": "SERVER",
		"name": wantSpanName,
		"tags": map[string]string{
			"component":        "web",
			"http.url":         "http://example.com/" + path,
			"http.method":      httpMethod,
			"http.status_code": wantHTTPCodeStr,
			"span.kind":        "server",
		},
	}

	if wantHTTPCode == 500 {
		wantSpan["tags"].(map[string]string)["error"] = "true"
		testutil.AssertSpanWithErrorEvent(t, wantSpan, gotSpan)
	} else {
		testutil.AssertSpanWithNoErrorEvent(t, wantSpan, gotSpan)
	}
}

func router() http.Handler {
	mux := NewRouter(WithServiceName("my-service"))
	mux.Handle("/200", okHandler())
	mux.Handle("/500", errorHandler(http.StatusInternalServerError))
	mux.Handle("/405", okHandler()).Methods("GET")
	mux.NotFoundHandler = errorHandler(http.StatusNotFound)
	mux.MethodNotAllowedHandler = errorHandler(http.StatusMethodNotAllowed)
	return mux
}

func errorHandler(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, fmt.Sprintf("%d!", code), code)
	})
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("200!\n"))
	})
}

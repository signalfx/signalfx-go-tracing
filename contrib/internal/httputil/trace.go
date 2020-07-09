package httputil // import "github.com/signalfx/signalfx-go-tracing/contrib/internal/httputil"

//go:generate sh -c "go run make_responsewriter.go | gofmt > trace_gen.go"

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/signalfx/signalfx-go-tracing/ddtrace"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
)

// TraceAndServe will apply tracing to the given http.Handler using the passed tracer under the given service and resource.
func TraceAndServe(h http.Handler, w http.ResponseWriter, r *http.Request, service, resource string, spanopts ...ddtrace.StartSpanOption) {
	originalURL := url.URL{
		Scheme: "http",
		Host: r.Host,
		RawPath: r.URL.RawPath,
		Path: r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}
	if r.TLS != nil {
		originalURL.Scheme = "https"
	}

	opts := append([]ddtrace.StartSpanOption{
		tracer.SpanType(ext.SpanTypeWeb),
		tracer.ServiceName(service),
		tracer.ResourceName(resource),
		tracer.Tag(ext.HTTPMethod, r.Method),
		tracer.Tag(ext.HTTPURL, originalURL.String()),
	}, spanopts...)
	if spanctx, err := tracer.Extract(tracer.HTTPHeadersCarrier(r.Header)); err == nil {
		opts = append(opts, tracer.ChildOf(spanctx))
	}
	span, ctx := tracer.StartSpanFromContext(r.Context(), "http.request", opts...)
	defer span.Finish()

	w = wrapResponseWriter(w, span)

	if strings.EqualFold("true", os.Getenv("SIGNALFX_SERVER_TIMING_CONTEXT")) {
		if traceParent, ok := tracer.FormatAsTraceParent(span.Context()); ok {
			w.Header().Add("Access-Control-Expose-Headers", "Server-Timing")
			w.Header().Add("Server-Timing", traceParent)
		}

	h.ServeHTTP(w, r.WithContext(ctx))
}

// responseWriter is a small wrapper around an http response writer that will
// intercept and store the status of a request.
type responseWriter struct {
	http.ResponseWriter
	span   ddtrace.Span
	status int
}

func newResponseWriter(w http.ResponseWriter, span ddtrace.Span) *responseWriter {
	return &responseWriter{w, span, 0}
}

// Write writes the data to the connection as part of an HTTP reply.
// We explicitely call WriteHeader with the 200 status code
// in order to get it reported into the span.
func (w *responseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// WriteHeader sends an HTTP response header with status code.
// It also sets the status code to the span.
func (w *responseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.status = status
	w.span.SetTag(ext.HTTPCode, strconv.Itoa(status))
	if status >= 500 && status < 600 {
		w.span.SetTag(ext.Error, fmt.Errorf("%d: %s", status, http.StatusText(status)))
	}
}

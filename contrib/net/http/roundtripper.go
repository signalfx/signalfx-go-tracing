package http

import (
	"fmt"
	"github.com/signalfx/signalfx-go-tracing/contrib/internal/log"
	"github.com/signalfx/signalfx-go-tracing/ddtrace"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"net/http"
	"os"
	"strconv"
)

const defaultResourceName = "http.request"

type roundTripper struct {
	base http.RoundTripper
	cfg  *roundTripperConfig
}

func (rt *roundTripper) RoundTrip(req *http.Request) (res *http.Response, err error) {
	opts := []ddtrace.StartSpanOption{
		tracer.SpanType(ext.SpanTypeHTTP),
		tracer.ResourceName(defaultResourceName),
		tracer.Tag(ext.HTTPMethod, req.Method),
		tracer.Tag(ext.HTTPURL, req.URL.String()),
	}
	if rate := rt.cfg.analyticsRate; rate > 0 {
		opts = append(opts, tracer.Tag(ext.EventSampleRate, rate))
	}
	span, ctx := tracer.StartSpanFromContext(req.Context(), defaultResourceName, opts...)
	defer func() {
		if rt.cfg.after != nil {
			rt.cfg.after(res, span)
		}
		span.Finish(tracer.WithError(err))
	}()
	if rt.cfg.before != nil {
		rt.cfg.before(req, span)
	}
	// inject the span context into the http request
	err = tracer.Inject(span.Context(), tracer.HTTPHeadersCarrier(req.Header))
	if err != nil {
		// this should never happen
		fmt.Fprintf(os.Stderr, "contrib/net/http.Roundtrip: failed to inject http headers: %v\n", err)
	}
	res, err = rt.base.RoundTrip(req.WithContext(ctx))
	if err != nil {
		log.LogError(span, err)
	} else {
		span.SetTag(ext.HTTPCode, strconv.Itoa(res.StatusCode))
		// treat 5XX as errors but there's no err object to log.
		if res.StatusCode/100 == 5 {
			span.SetTag(ext.Error, "true")
		}
	}
	return res, err
}

// WrapRoundTripper returns a new RoundTripper which traces all requests sent
// over the transport.
func WrapRoundTripper(rt http.RoundTripper, opts ...RoundTripperOption) http.RoundTripper {
	cfg := newRoundTripperConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	if wrapped, ok := rt.(*roundTripper); ok {
		rt = wrapped.base
	}
	return &roundTripper{
		base: rt,
		cfg:  cfg,
	}
}

// WrapClient modifies the given client's transport to augment it with tracing and returns it.
func WrapClient(c *http.Client, opts ...RoundTripperOption) *http.Client {
	if c.Transport == nil {
		c.Transport = http.DefaultTransport
	}
	c.Transport = WrapRoundTripper(c.Transport, opts...)
	return c
}

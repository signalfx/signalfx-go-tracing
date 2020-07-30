// Package opentracer provides a wrapper on top of the Datadog tracer that can be used with Opentracing.
// It also provides a set of opentracing.StartSpanOption that are specific to Datadog's APM product.
// To use it, simply call "New".
//
// Note that there are currently some small incompatibilities between the Opentracing spec and the Datadog
// APM product, which we are in the process of addressing on the long term. When using Datadog, the
// Opentracing operation name is what is called resource in Datadog's terms and the Opentracing "component"
// tag is Datadog's operation name. Meaning that in order to define (in Opentracing terms) a span that
// has the operation name "/user/profile" and the component "http.request", one would do:
//  opentracing.StartSpan("http.request", opentracer.ResourceName("/user/profile"))
//
// Some libraries and frameworks are supported out-of-the-box by using our integrations. You can see a list
// of supported integrations here: https://godoc.org/github.com/signalfx/signalfx-go-tracing/contrib. They are fully
// compatible with the Opentracing implementation.
package opentracer

import (
	"github.com/opentracing/opentracing-go"
	"github.com/signalfx/signalfx-go-tracing/ddtrace"
)

// New creates, instantiates and returns an Opentracing compatible version of the
// tracer using the provided set of options.
func New() opentracing.Tracer {
	return &opentracer{ddtrace.GetGlobalTracer()}
}

var _ opentracing.Tracer = (*opentracer)(nil)

// opentracer implements opentracing.Tracer on top of ddtrace.Tracer.
type opentracer struct{ ddtrace.Tracer }

// StartSpan implements opentracing.Tracer.
func (t *opentracer) StartSpan(operationName string, options ...opentracing.StartSpanOption) opentracing.Span {
	var sso opentracing.StartSpanOptions
	for _, o := range options {
		o.Apply(&sso)
	}
	opts := []ddtrace.StartSpanOption{ddtrace.WithStartTime(sso.StartTime)}
	for _, ref := range sso.References {
		if v, ok := ref.ReferencedContext.(ddtrace.SpanContext); ok && ref.Type == opentracing.ChildOfRef {
			opts = append(opts, ddtrace.WithChildOf(v))
			break // can only have one parent
		}
	}
	for k, v := range sso.Tags {
		opts = append(opts, ddtrace.WithTag(k, v))
	}
	return t.Tracer.StartSpan(operationName, opts...)
}

// Inject implements opentracing.Tracer.
func (t *opentracer) Inject(ctx opentracing.SpanContext, format interface{}, carrier interface{}) error {
	sctx, ok := ctx.(ddtrace.SpanContext)
	if !ok {
		return opentracing.ErrUnsupportedFormat
	}
	switch format {
	case opentracing.TextMap, opentracing.HTTPHeaders:
		return t.Tracer.Inject(sctx, carrier)
	default:
		return opentracing.ErrUnsupportedFormat
	}
}

// Extract implements opentracing.Tracer.
func (t *opentracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	switch format {
	case opentracing.TextMap, opentracing.HTTPHeaders:
		return t.Tracer.Extract(carrier)
	default:
		return nil, opentracing.ErrUnsupportedFormat
	}
}

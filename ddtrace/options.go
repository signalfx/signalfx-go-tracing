// Package ddtrace contains the interfaces that specify the implementations of Datadog's
// tracing library, as well as a set of sub-packages containing various implementations:
// our native implementation ("tracer"), a wrapper that can be used with Opentracing
// ("opentracer") and a mock tracer to be used for testing ("mocktracer"). Additionally,
// package "ext" provides a set of tag names and values specific to Datadog's APM product.
//
// To get started, visit the documentation for any of the packages you'd like to begin
// with by accessing the subdirectories of this package: https://godoc.org/github.com/signalfx/signalfx-go-tracing/ddtrace#pkg-subdirectories.
package ddtrace // import "github.com/signalfx/signalfx-go-tracing/ddtrace"

import "time"

// WithTag sets the given key/value pair as a tag on the started Span.
func WithTag(k string, v interface{}) StartSpanOption {
	return func(cfg *StartSpanConfig) {
		if cfg.Tags == nil {
			cfg.Tags = map[string]interface{}{}
		}
		cfg.Tags[k] = v
	}
}

// WithChildOf tells StartSpan to use the given span context as a parent for the
// created span.
func WithChildOf(ctx SpanContext) StartSpanOption {
	return func(cfg *StartSpanConfig) {
		cfg.Parent = ctx
	}
}

// WithStartTime sets a custom time as the start time for the created span. By
// default a span is started using the creation time.
func WithStartTime(t time.Time) StartSpanOption {
	return func(cfg *StartSpanConfig) {
		cfg.StartTime = t
	}
}

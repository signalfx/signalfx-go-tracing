// Package ddtrace contains the interfaces that specify the implementations of Datadog's
// tracing library, as well as a set of sub-packages containing various implementations:
// our native implementation ("tracer"), a wrapper that can be used with Opentracing
// ("opentracer") and a mock tracer to be used for testing ("mocktracer"). Additionally,
// package "ext" provides a set of tag names and values specific to Datadog's APM product.
//
// To get started, visit the documentation for any of the packages you'd like to begin
// with by accessing the subdirectories of this package: https://godoc.org/github.com/signalfx/signalfx-go-tracing/ddtrace#pkg-subdirectories.
package ddtrace // import "github.com/signalfx/signalfx-go-tracing/ddtrace"

import (
	"time"

	"github.com/opentracing/opentracing-go"
)

// Tracer specifies an implementation of the Datadog tracer which allows starting
// and propagating spans. The official implementation if exposed as functions
// within the "tracer" package.
type Tracer interface {
	// StartSpan starts a span with the given operation name and options.
	StartSpan(operationName string, opts ...StartSpanOption) Span

	// Extract extracts a span context from a given carrier. Note that baggage item
	// keys will always be lower-cased to maintain consistency. It is impossible to
	// maintain the original casing due to MIME header canonicalization standards.
	Extract(carrier interface{}) (SpanContext, error)

	// Inject injects a span context into the given carrier.
	Inject(context SpanContext, carrier interface{}) error

	// Stop stops the active tracer and sets the global tracer to a no-op. Calls to
	// Stop should be idempotent.
	Stop()

	// ForceFlush pending traces
	ForceFlush()
}

// Span represents a chunk of computation time. Spans have names, durations,
// timestamps and other metadata. A Tracer is used to create hierarchies of
// spans in a request, buffer and submit them to the server.
type Span interface {
	opentracing.Span
	FinishWithOptionsExt(opts ...FinishOption)
}

// SpanContext represents a span state that can propagate to descendant spans
// and across process boundaries. It contains all the information needed to
// spawn a direct descendant of the span that it belongs to. It can be used
// to create distributed tracing by propagating it using the provided interfaces.
type SpanContext = opentracing.SpanContext

// StartSpanOption is a configuration option that can be used with a Tracer's StartSpan method.
type StartSpanOption func(cfg *StartSpanConfig)

// FinishOption is a configuration option that can be used with a Span's Finish method.
type FinishOption func(cfg *FinishConfig)

// FinishConfig holds the configuration for finishing a span. It is usually passed around by
// reference to one or more FinishOption functions which shape it into its final form.
type FinishConfig struct {
	opentracing.FinishOptions

	// Error holds an optional error that should be set on the span before
	// finishing.
	Error error

	// NoDebugStack will prevent any set errors from generating an attached stack trace tag.
	NoDebugStack bool

	// StackFrames specifies the number of stack frames to be attached in spans that finish with errors.
	StackFrames uint

	// SkipStackFrames specifies the offset at which to start reporting stack frames from the stack.
	SkipStackFrames uint
}

// StartSpanConfig holds the configuration for starting a new span. It is usually passed
// around by reference to one or more StartSpanOption functions which shape it into its
// final form.
type StartSpanConfig struct {
	// Parent holds the SpanContext that should be used as a parent for the
	// new span. If nil, implementations should return a root span.
	Parent SpanContext

	// StartTime holds the time that should be used as the start time of the span.
	// Implementations should use the current time when StartTime.IsZero().
	StartTime time.Time

	// Tags holds a set of key/value pairs that should be set as metadata on the
	// new span.
	Tags map[string]interface{}

	// Force-set the SpanID, rather than use a random number. If no Parent SpanContext is present,
	// then this will also set the TraceID to the same value.
	SpanID uint64

	// RecordedValueMaxLength determines the maximum allowed length a tag/log can have.
	RecordedValueMaxLength *int
}

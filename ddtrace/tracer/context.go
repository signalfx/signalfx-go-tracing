package tracer

import (
	"context"
	"log"

	"github.com/opentracing/opentracing-go"
	"github.com/signalfx/signalfx-go-tracing/ddtrace"
)

// ContextWithSpan returns a copy of the given context which includes the span s.
func ContextWithSpan(ctx context.Context, s Span) context.Context {
	return opentracing.ContextWithSpan(ctx, s)
}

// SpanFromContext returns the span contained in the given context. A second return
// value indicates if a span was found in the context. If no span is found, a no-op
// span is returned.
func SpanFromContext(ctx context.Context) (Span, bool) {
	if ctx == nil {
		return &ddtrace.NoopSpan{}, false
	}
	s := opentracing.SpanFromContext(ctx)
	span, ok := s.(ddtrace.Span)
	if !ok {
		// span on context is not nil but it's not ddtrace.Span
		// explicitly log unsupported behavior
		if s != nil {
			log.Println(warningPrefix + "found invalid span. Only spans generated by the in-built tracer are valid. Third-party traces are not supported")
		}
		return &ddtrace.NoopSpan{}, false
	}
	return span, true
}

// StartSpanFromContext returns a new span with the given operation name and options. If a span
// is found in the context, it will be used as the parent of the resulting span. If the ChildOf
// option is passed, the span from context will take precedence over it as the parent span.
func StartSpanFromContext(ctx context.Context, operationName string, opts ...StartSpanOption) (Span, context.Context) {
	if s, ok := SpanFromContext(ctx); ok {
		opts = append(opts, ChildOf(s.Context()))
	}
	s := StartSpan(operationName, opts...)
	return s, ContextWithSpan(ctx, s)
}

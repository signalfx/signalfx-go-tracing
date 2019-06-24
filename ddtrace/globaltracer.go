package ddtrace // import "github.com/signalfx/signalfx-go-tracing/ddtrace"

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"sync"
)

var (
	mu           sync.RWMutex // guards globalTracer
	globalTracer Tracer = &NoopTracer{}
)

// SetGlobalTracer sets the global tracer to t.
func SetGlobalTracer(t Tracer) {
	mu.Lock()
	defer mu.Unlock()
	if !Testing {
		// avoid infinite loop when calling (*mocktracer.Tracer).Stop
		globalTracer.Stop()
	}
	globalTracer = t
}

// GetGlobalTracer returns the currently active tracer.
func GetGlobalTracer() Tracer {
	mu.RLock()
	defer mu.RUnlock()
	return globalTracer
}

// Testing is set to true when the mock tracer is active. It usually signifies that we are in a test
// environment. This value is used by tracer.Start to prevent overriding the GlobalTracer in tests.
var Testing = false

var _ Tracer = (*NoopTracer)(nil)

// NoopTracer is an implementation of ddtrace.Tracer that is a no-op.
type NoopTracer struct{}

// ForceFlush traces
func (NoopTracer) ForceFlush() {
}

// StartSpan implements ddtrace.Tracer.
func (NoopTracer) StartSpan(operationName string, opts ...StartSpanOption) Span {
	return NoopSpan{}
}

// SetServiceInfo implements ddtrace.Tracer.
func (NoopTracer) SetServiceInfo(name, app, appType string) {}

// Extract implements ddtrace.Tracer.
func (NoopTracer) Extract(carrier interface{}) (SpanContext, error) {
	return NoopSpanContext{}, nil
}

// Inject implements ddtrace.Tracer.
func (NoopTracer) Inject(context SpanContext, carrier interface{}) error { return nil }

// Stop implements ddtrace.Tracer.
func (NoopTracer) Stop() {}

var _ Span = (*NoopSpan)(nil)

// NoopSpan is an implementation of ddtrace.Span that is a no-op.
type NoopSpan struct{}

// FinishWithOptionsExt implements ddtrace.Span.
func (NoopSpan) FinishWithOptionsExt(opts ...FinishOption) {
}

// Finish implements ddtrace.Span.
func (NoopSpan) Finish() {
}

// FinishWithOptions implements ddtrace.Span.
func (NoopSpan) FinishWithOptions(opts opentracing.FinishOptions) {
}

// Context implements ddtrace.Span.
func (NoopSpan) Context() opentracing.SpanContext {
	return NoopSpanContext{}
}

// SetOperationName implements ddtrace.Span.
func (NoopSpan) SetOperationName(operationName string) opentracing.Span {
	return NoopSpan{}
}

// SetTag implements ddtrace.Span.
func (NoopSpan) SetTag(key string, value interface{}) opentracing.Span {
	return NoopSpan{}
}

// LogFields impelments ddtrace.Span
func (NoopSpan) LogFields(fields ...log.Field) {
}

// LogKV implements ddtrace.Span.
func (NoopSpan) LogKV(alternatingKeyValues ...interface{}) {
}

// SetBaggageItem implements ddtrace.Span.
func (NoopSpan) SetBaggageItem(restrictedKey, value string) opentracing.Span {
	return NoopSpan{}
}

// Tracer implements ddtrace.Span.
func (NoopSpan) Tracer() opentracing.Tracer {
	panic("not implemented")
}

// LogEvent implements ddtrace.Span.
func (NoopSpan) LogEvent(event string) {
}

// LogEventWithPayload implements ddtrace.Span.
func (NoopSpan) LogEventWithPayload(event string, payload interface{}) {
}

// Log implements ddtrace.Span.
func (NoopSpan) Log(data opentracing.LogData) {
}

// BaggageItem implements ddtrace.Span.
func (NoopSpan) BaggageItem(key string) string { return "" }

var _ SpanContext = (*NoopSpanContext)(nil)

// NoopSpanContext is an implementation of ddtrace.SpanContext that is a no-op.
type NoopSpanContext struct{}

// ForeachBaggageItem implements ddtrace.SpanContext.
func (NoopSpanContext) ForeachBaggageItem(handler func(k, v string) bool) {}

//go:generate msgp -unexported -marshal=false -o=span_msgp.go -tests=false

package tracer

import (
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/tinylib/msgp/msgp"

	"github.com/signalfx/signalfx-go-tracing/ddtrace"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/opentracer"
)

type (
	// spanList implements msgp.Encodable on top of a slice of spans.
	spanList []*span

	// spanLists implements msgp.Decodable on top of a slice of spanList.
	// This type is only used in tests.
	spanLists []spanList
)

var (
	_ ddtrace.Span   = (*span)(nil)
	_ msgp.Encodable = (*spanList)(nil)
	_ msgp.Decodable = (*spanLists)(nil)
)

// errorConfig holds customization options for setting error tags.
type errorConfig struct {
	noDebugStack bool
	stackFrames  uint
	stackSkip    uint
}

// logFields holds the results of one invocation of LogFields
type logFields struct {
	fields map[string]interface{}
	time   time.Time
}

// span represents a computation. Callers must call Finish when a span is
// complete to ensure it's submitted.
type span struct {
	sync.RWMutex `msg:"-"`

	Name     string             `msg:"name"`              // operation name
	Service  string             `msg:"service"`           // service name (i.e. "grpc.server", "http.request")
	Resource string             `msg:"resource"`          // resource name (i.e. "/user?id=123", "SELECT * FROM users")
	Type     string             `msg:"type"`              // protocol associated with the span (i.e. "web", "db", "cache")
	Start    int64              `msg:"start"`             // span start time expressed in nanoseconds since epoch
	Duration int64              `msg:"duration"`          // duration of the span expressed in nanoseconds
	Meta     map[string]string  `msg:"meta,omitempty"`    // arbitrary map of metadata
	Metrics  map[string]float64 `msg:"metrics,omitempty"` // arbitrary map of numeric metrics
	SpanID   uint64             `msg:"span_id"`           // identifier of this span
	TraceID  uint64             `msg:"trace_id"`          // identifier of the root span
	ParentID uint64             `msg:"parent_id"`         // identifier of the span's direct parent
	Error    int32              `msg:"error"`             // error status of the span; 0 means no errors
	Logs     []*logFields

	recordedValueMaxLength *int
	finished               bool         `msg:"-"` // true if the span has been submitted to a tracer.
	context                *spanContext `msg:"-"` // span propagation context
}

// LogKV pairs
func (s *span) LogKV(alternatingKeyValues ...interface{}) {
	fields, err := log.InterleavedKVToFields(alternatingKeyValues...)
	if err != nil {
		return
	}
	s.LogFields(fields...)
}

// Tracer returns the global tracer
func (s *span) Tracer() opentracing.Tracer {
	// return ddtrace.GetGlobalTracer() wrapped to make it
	// a compatible opentracing Tracer
	return opentracer.New()
}

func (s *span) truncate(val string) string {
	if s.recordedValueMaxLength == nil {
		return val
	}
	max := *s.recordedValueMaxLength
	if max < 0 {
		return val
	}
	if len(val) > max {
		val = val[:max]
	}
	return val
}

// LogFields field to span
func (s *span) LogFields(fields ...log.Field) {
	m := map[string]interface{}{}
	for _, field := range fields {
		val := field.Value()
		if vStr, ok := val.(string); ok {
			val = s.truncate(vStr)
		}
		m[field.Key()] = val
	}
	s.Logs = append(s.Logs, &logFields{m, time.Now()})
}

// LogEvent is deprecated
func (s *span) LogEvent(event string) { /* deprecated */ }

// LogEventWithPayload is deprecated
func (s *span) LogEventWithPayload(event string, payload interface{}) { /* deprecated */ }

// Log is deprecated
func (s *span) Log(data opentracing.LogData) { /* deprecated */ }

// Context yields the SpanContext for this Span. Note that the return
// value of Context() is still valid after a call to Finish(). This is
// called the span context and it is different from Go's context.
func (s *span) Context() ddtrace.SpanContext { return s.context }

// SetBaggageItem sets a key/value pair as baggage on the span. Baggage items
// are propagated down to descendant spans and injected cross-process. Use with
// care as it adds extra load onto your tracing layer.
func (s *span) SetBaggageItem(key, val string) opentracing.Span {
	s.context.setBaggageItem(key, val)
	return s
}

// BaggageItem gets the value for a baggage item given its key. Returns the
// empty string if the value isn't found in this Span.
func (s *span) BaggageItem(key string) string {
	return s.context.baggageItem(key)
}

// SetTag adds a set of key/value metadata to the span.
func (s *span) SetTag(key string, value interface{}) opentracing.Span {
	s.Lock()
	defer s.Unlock()
	// We don't lock spans when flushing, so we could have a data race when
	// modifying a span as it's being flushed. This protects us against that
	// race, since spans are marked `finished` before we flush them.
	if s.finished {
		return s
	}

	if key == ext.Error {
		s.setTagError(value, &errorConfig{})
		return s
	}

	if v, ok := value.(bool); ok {
		s.setTagBool(key, v)
		return s
	}
	if v, ok := value.(string); ok {
		s.setTagString(key, v)
		return s
	}
	if v, ok := toFloat64(value); ok {
		s.setTagNumeric(key, v)
		return s
	}
	// not numeric, not a string and not an error, the likelihood of this
	// happening is close to zero, but we should nevertheless account for it.
	s.Meta[key] = fmt.Sprint(value)
	return s
}

// setTagError sets the error tag. It accounts for various valid scenarios.
// This method is not safe for concurrent use.
func (s *span) setTagError(value interface{}, cfg *errorConfig) {
	if s.finished {
		return
	}
	switch v := value.(type) {
	case bool:
		// bool value as per Opentracing spec.
		if !v {
			s.Error = 0
		} else {
			s.Error = 1
		}
	case error:
		// if anyone sets an error value as the tag, be nice here
		// and provide all the benefits.
		s.Error = 1

		s.setTagString(ext.ErrorKind, reflect.TypeOf(v).String())
		s.setTagString(ext.ErrorObject, fmt.Sprintf("%#v", v))
		s.setTagString(ext.ErrorMsg, v.Error())

		if !cfg.noDebugStack {
			if cfg.stackFrames == 0 {
				// TODO: maybe support xerrors and/or https://godoc.org/github.com/pkg/errors?
				s.setTagString(ext.ErrorStack, string(debug.Stack()))
			} else {
				s.setTagString(ext.ErrorStack, takeStacktrace(cfg.stackFrames, cfg.stackSkip))
			}
		}
	case nil:
		// no error
		s.Error = 0
	default:
		// in all other cases, let's assume that setting this tag
		// is the result of an error.
		s.Error = 1
	}

	if s.Error == 1 {
		s.setTagString(ext.Error, "true")
	}
}

// takeStacktrace takes stacktrace
func takeStacktrace(n, skip uint) string {
	var builder strings.Builder
	pcs := make([]uintptr, n)

	// +2 to exclude runtime.Callers and takeStacktrace
	numFrames := runtime.Callers(2+int(skip), pcs)
	if numFrames == 0 {
		return ""
	}
	frames := runtime.CallersFrames(pcs[:numFrames])
	for i := 0; ; i++ {
		frame, more := frames.Next()
		if i != 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(frame.Function)
		builder.WriteByte('\n')
		builder.WriteByte('\t')
		builder.WriteString(frame.File)
		builder.WriteByte(':')
		builder.WriteString(strconv.Itoa(frame.Line))
		if !more {
			break
		}
	}
	return builder.String()
}

// setTagString sets a string tag. This method is not safe for concurrent use.
func (s *span) setTagString(key, v string) {
	switch key {
	case ext.SpanName:
		s.Name = v
	case ext.ServiceName:
		s.Service = v
	case ext.ResourceName:
		s.Resource = v
	case ext.SpanType:
		s.Type = v
	default:
		s.Meta[key] = s.truncate(v)
	}
}

// setTagBool sets a boolean tag on the span.
func (s *span) setTagBool(key string, v bool) {
	switch key {
	case ext.AnalyticsEvent:
		if v {
			s.setTagNumeric(ext.EventSampleRate, 1.0)
		} else {
			s.setTagNumeric(ext.EventSampleRate, 0.0)
		}
	case ext.ManualDrop:
		if v {
			s.setTagNumeric(ext.SamplingPriority, ext.PriorityUserReject)
		}
	case ext.ManualKeep:
		if v {
			s.setTagNumeric(ext.SamplingPriority, ext.PriorityUserKeep)
		}
	default:
		if v {
			s.setTagString(key, "true")
		} else {
			s.setTagString(key, "false")
		}
	}
}

// setTagNumeric sets a numeric tag, in our case called a metric. This method
// is not safe for concurrent use.
func (s *span) setTagNumeric(key string, v float64) {
	switch key {
	case ext.SamplingPriority:
		// setting sampling priority per spec
		s.Metrics[keySamplingPriority] = v
		s.context.setSamplingPriority(int(v))
	default:
		s.Metrics[key] = v
	}
}

// FinishWithOptionsExt closes this Span (but not its children) providing the duration
// of its part of the tracing session.
func (s *span) FinishWithOptionsExt(opts ...ddtrace.FinishOption) {
	var cfg ddtrace.FinishConfig
	for _, fn := range opts {
		fn(&cfg)
	}
	var t int64
	if cfg.FinishTime.IsZero() {
		t = now()
	} else {
		t = cfg.FinishTime.UnixNano()
	}
	if cfg.Error != nil {
		s.Lock()
		s.setTagError(cfg.Error, &errorConfig{
			noDebugStack: cfg.NoDebugStack,
			stackFrames:  cfg.StackFrames,
			stackSkip:    cfg.SkipStackFrames,
		})
		s.Unlock()
	}
	s.internalFinish(t)
}

// FinishWithOptions closes this span
func (s *span) FinishWithOptions(opts opentracing.FinishOptions) {
	for _, lr := range opts.LogRecords {
		if len(lr.Fields) > 0 {
			s.LogFields(lr.Fields...)
		}
	}

	s.internalFinish(opts.FinishTime.UnixNano())
}

// Finish closes this span
func (s *span) Finish() {
	s.internalFinish(time.Now().UnixNano())
}

// SetOperationName sets or changes the operation name.
func (s *span) SetOperationName(operationName string) opentracing.Span {
	s.Lock()
	defer s.Unlock()

	s.Name = operationName
	return s
}

func (s *span) internalFinish(finishTime int64) {
	s.Lock()
	defer s.Unlock()
	// We don't lock spans when flushing, so we could have a data race when
	// modifying a span as it's being flushed. This protects us against that
	// race, since spans are marked `finished` before we flush them.
	if s.finished {
		// already finished
		return
	}
	if s.Duration == 0 {
		s.Duration = finishTime - s.Start
	}
	s.finished = true

	if s.context.drop {
		// not sampled by local sampler
		return
	}
	s.context.finish()
}

// String returns a human readable representation of the span. Not for
// production, just debugging.
func (s *span) String() string {
	lines := []string{
		fmt.Sprintf("Name: %s", s.Name),
		fmt.Sprintf("Service: %s", s.Service),
		fmt.Sprintf("Resource: %s", s.Resource),
		fmt.Sprintf("TraceID: %d", s.TraceID),
		fmt.Sprintf("SpanID: %d", s.SpanID),
		fmt.Sprintf("ParentID: %d", s.ParentID),
		fmt.Sprintf("Start: %s", time.Unix(0, s.Start)),
		fmt.Sprintf("Duration: %s", time.Duration(s.Duration)),
		fmt.Sprintf("Error: %d", s.Error),
		fmt.Sprintf("Type: %s", s.Type),
		"Tags:",
	}
	s.RLock()
	for key, val := range s.Meta {
		lines = append(lines, fmt.Sprintf("\t%s:%s", key, val))
	}
	for key, val := range s.Metrics {
		lines = append(lines, fmt.Sprintf("\t%s:%f", key, val))
	}
	lines = append(lines, "Logs:")
	for _, lf := range s.Logs {
		lines = append(lines, fmt.Sprintf("\t%s", lf.time))
		for key, val := range lf.fields {
			lines = append(lines, fmt.Sprintf("\t\t%s:%s", key, val))
		}
	}
	s.RUnlock()
	return strings.Join(lines, "\n")
}

const (
	keySamplingPriority     = "_sampling_priority_v1"
	keySamplingPriorityRate = "_sampling_priority_rate_v1"
	keyOrigin               = "_dd.origin"
	keyHostname             = "_dd.hostname"
)

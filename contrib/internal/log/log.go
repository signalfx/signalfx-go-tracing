package log

import (
	"fmt"
	"github.com/signalfx/signalfx-go-tracing/ddtrace"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"reflect"
	"runtime/debug"
)

// LogError adds log fields to the span from the err
func LogError(span tracer.Span, err error) {
	span.SetTag(ext.Error, "true")

	span.LogFields(
		ddtrace.LogField("event", "error"),
		ddtrace.LogField("error.kind", reflect.TypeOf(err).String()),
		ddtrace.LogField("error.object", fmt.Sprintf("%#v", err)),
		ddtrace.LogField("message", err.Error()),
		// TODO: maybe support xerrors and/or https://godoc.org/github.com/pkg/errors?
		ddtrace.LogField("stack", string(debug.Stack())),
	)
}

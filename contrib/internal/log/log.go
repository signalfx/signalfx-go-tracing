package log

import (
	"fmt"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"reflect"
	"runtime/debug"
	"time"
)

// LogError adds log fields to the span from the err
func LogError(span tracer.Span, err error) {
	now := time.Now()

	span.SetTag(ext.Error, "true")

	span.AddLog("event", "error", now)
	span.AddLog("error.kind", reflect.TypeOf(err).String(), now)
	span.AddLog("error.object", fmt.Sprintf("%#v", err), now)
	span.AddLog("message", err.Error(), now)

	// TODO: maybe support xerrors and/or https://godoc.org/github.com/pkg/errors?
	span.AddLog("stack", string(debug.Stack()), now)
}


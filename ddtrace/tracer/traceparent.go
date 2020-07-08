package tracer

import (
	"fmt"
	"github.com/signalfx/signalfx-go-tracing/ddtrace"
)

func FormatAsTraceParent(context ddtrace.SpanContext) (string,bool) {
	ctx, ok := context.(*spanContext)
	if !ok || ctx.traceID == 0 || ctx.spanID == 0 {
		return "",false
	}
	answer := fmt.Sprintf("traceparent;desc=\"00-%032x-%016x-01\"", ctx.traceID, ctx.spanID)
	return answer,true
}

package tracer

import (
	"testing"
	"regexp"

	"github.com/stretchr/testify/assert"
)

func TestTraceParent(t *testing.T) {                
	tracer := newTracer()
	span := tracer.StartSpan("web.request").(*span)
	traceParent,ok := FormatAsTraceParent(span.Context())

	assert := assert.New(t)
	assert.True(ok)
	matched,_ := regexp.MatchString("^traceparent;desc=\"00-[0-9a-f]{32}-[0-9a-f]{16}-01\"$", traceParent)
	assert.True(matched)
}

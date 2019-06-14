package opentracer

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/signalfx/signalfx-go-tracing/ddtrace"
)

func TestStart(t *testing.T) {
	assert := assert.New(t)
	ot := New()
	dd, ok := ddtrace.GetGlobalTracer().(ddtrace.Tracer)
	assert.True(ok)
	ott, ok := ot.(*opentracer)
	assert.True(ok)
	assert.Equal(ott.Tracer, dd)
}

package testutil

import (
	"github.com/signalfx/golib/trace"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/stretchr/testify/assert"
	"testing"
)

type ErrorAssertion struct {
	KindEquals      string
	MessageContains string
	ObjectContains  string
	StackContains   []string
	StackMinLength  int
}

func AssertSpanWithTags(t *testing.T, span *trace.Span, expected map[string]string) {
	for k,v := range expected {
		assert.Equal(t, span.Tags[k], v)
	}
}

func AssertSpanWithError(t *testing.T, span *trace.Span, err ErrorAssertion) {
	assert.Equal(t, "true", span.Tags[ext.Error])
	assert.Equal(t, err.KindEquals, span.Tags[ext.ErrorKind])
	assert.Contains(t, span.Tags[ext.ErrorMsg], err.MessageContains)
	assert.Contains(t, span.Tags[ext.ErrorObject], err.ObjectContains)
	assert.Greater(t, len(span.Tags[ext.ErrorStack]), err.StackMinLength)
	for _, s := range err.StackContains {
		assert.Contains(t, span.Tags[ext.ErrorStack], s)
	}
}

func AssertSpanWithNoError(t *testing.T, span *trace.Span) {
	assert.NotContains(t, span.Tags, ext.Error)
}

func AssertSpan(t *testing.T, expected map[string]interface{}, actualSpan *trace.Span) {
	assertIfNotNil(t, expected["kind"], actualSpan.Kind)
	assertIfNotNil(t, expected["name"], actualSpan.Name)
	assert.Equal(t, expected["tags"], actualSpan.Tags)
}

func assertIfNotNil(t *testing.T, expected interface{}, actual *string) {
	if assert.NotNil(t, actual) {
		assert.Equal(t, expected, *actual)
	}
}
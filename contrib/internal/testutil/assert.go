package testutil

import (
	"github.com/signalfx/golib/trace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func AssertSpanWithErrorEvent(t *testing.T, expected map[string]interface{}, actualSpan *trace.Span) {
	AssertSpan(t, expected, actualSpan)
	require.Len(t, actualSpan.Annotations, 1)
}

func AssertSpanWithNoErrorEvent(t *testing.T, expected map[string]interface{}, actualSpan *trace.Span) {
	AssertSpan(t, expected, actualSpan)
	assert.Len(t, actualSpan.Annotations, 0)

}

func AssertSpan(t *testing.T, expected map[string]interface{}, actualSpan *trace.Span) {
	assertIfNotNil(t, expected["kind"], *actualSpan.Kind)
	assertIfNotNil(t, expected["name"], *actualSpan.Name)
	assert.Equal(t, expected["tags"], actualSpan.Tags)
}

func assertIfNotNil(t *testing.T, expected interface{}, actual string) {
	if assert.NotNil(t, actual) {
		assert.Equal(t, expected, actual)
	}
}
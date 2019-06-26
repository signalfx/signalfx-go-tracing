package testutil

import (
	"github.com/signalfx/golib/trace"
	"github.com/stretchr/testify/assert"
)

func AssertSpan(t assert.TestingT, expected map[string]interface{}, actualSpan *trace.Span) {
	assert.Equal(t, expected["kind"], *actualSpan.Kind)
	assert.Equal(t, expected["name"], *actualSpan.Name)
	assert.Equal(t, expected["tags"], actualSpan.Tags)
}

func AssertSpanAnnotations(t assert.TestingT, expected map[string]string, actual map[string]string) {
	for key, val := range expected {
		assert.Equal(t, val, actual[key])
	}
}
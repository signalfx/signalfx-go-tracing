package testutil

import (
	"encoding/json"
	"github.com/signalfx/golib/trace"
	"github.com/stretchr/testify/require"
	"testing"
)

// GetAnnotation returns the deserialized annotation at idx
func GetAnnotation(t *testing.T, span *trace.Span, idx int) map[string]string {
	require.Greater(t, len(span.Annotations), idx)
	var msg map[string]string
	err := json.Unmarshal([]byte(*span.Annotations[idx].Value), &msg)
	require.NoError(t, err)
	return msg
}

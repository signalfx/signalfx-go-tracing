package testutil

import (
	"encoding/json"
	"github.com/signalfx/golib/trace"
	"github.com/stretchr/testify/require"
	"testing"
)

// GetAnnotations returns a map of the annotations set in the given span.
func GetAnnotations(t *testing.T, span *trace.Span) map[string]string {
	annotations := map[string]string{}

	for _, ann := range span.Annotations {
		var msg map[string]string
		err := json.Unmarshal([]byte(*ann.Value), &msg)
		require.NoError(t, err)
		for key, val := range msg {
			annotations[key] = val
		}
	}

	return annotations
}

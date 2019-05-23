package mocktracer_test

import (
	"github.com/signalfx/signalfx-go-tracing/ddtrace/mocktracer"
)

func Example() {
	// Start the mock tracer.
	mt := mocktracer.Start()
	defer mt.Stop()

	// ...run some code with generates spans.

	// Query the mock tracer for finished spans.
	spans := mt.FinishedSpans()
	if len(spans) != 1 {
		// should only have 1 span
	}

	// Run assertions...
}

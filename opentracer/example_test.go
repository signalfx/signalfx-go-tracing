package opentracer

import (
	"github.com/opentracing/opentracing-go"
	"github.com/signalfx/signalfx-go-tracing/tracing"
)

func Example() {
	// Start a Datadog tracer, optionally providing a set of options,
	// returning an opentracing.Tracer which wraps it.
	t := New(tracing.WithEndpointURL("host:port"))

	// Use it with the Opentracing API. The (already started) Datadog tracer
	// may be used in parallel with the Opentracing API if desired.
	opentracing.SetGlobalTracer(t)
}

package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/opentracer"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"os"
)

const (
	signalfxServiceName = "SIGNALFX_SERVICE_NAME"
	signalfxEndpointURL = "SIGNALFX_ENDPOINT_URL"
	signalfxAccessToken = "SIGNALFX_ACCESS_TOKEN"
)

var defaults = map[string]string{
	signalfxServiceName: "SignalFx-Tracing",
	signalfxEndpointURL: "http://localhost:9080/v1/trace",
	signalfxAccessToken: "",
}

type config struct {
	serviceName string
	accessToken string
	url         string
}

// StartOption is a function that configures an option for Start
type StartOption = func(*config)

func defaultConfig() *config {
	return &config{
		serviceName: envOrDefault(signalfxServiceName),
		accessToken: envOrDefault(signalfxAccessToken),
		url:         envOrDefault(signalfxEndpointURL),
	}
}

// envOrDefault gets the given environment variable if set otherwise a default value.
func envOrDefault(envVar string) string {
	if val := os.Getenv(envVar); val != "" {
		return val
	}
	return defaults[envVar]
}

// WithServiceName changes the reported service name
func WithServiceName(serviceName string) StartOption {
	return func(c *config) {
		c.serviceName = serviceName
	}
}

// WithAccessToken configures the SignalFx access token to use when reporting
func WithAccessToken(accessToken string) StartOption {
	return func(c *config) {
		c.accessToken = accessToken
	}
}

// WithEndpointURL configures the URL to send traces to
func WithEndpointURL(url string) StartOption {
	return func(c *config) {
		c.url = url
	}
}

// Start tracing globally
func Start(opts ...StartOption) {
	c := defaultConfig()
	for _, fn := range opts {
		fn(c)
	}

	tracer.Start(
		tracer.WithServiceName(c.serviceName),
		tracer.WithZipkin(c.serviceName, c.url, c.accessToken))
	opentracing.SetGlobalTracer(opentracer.New())
}

// Stop tracing globally
func Stop() {
	tracer.Stop()
	opentracing.SetGlobalTracer(&opentracing.NoopTracer{})
}

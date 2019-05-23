package tracing

import (
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
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

type StartOption = func(*config)


func defaultConfig() *config {
	return &config{
		serviceName: defaults[signalfxServiceName],
		accessToken: defaults[signalfxAccessToken],
		url:         defaults[signalfxEndpointURL],
	}
}

func WithServiceName(serviceName string) StartOption {
	return func(c *config) {
		c.serviceName = serviceName
	}
}

func WithAccessToken(accessToken string) StartOption {
	return func(c *config) {
		c.accessToken = accessToken
	}
}

func WithEndpointURL(url string) StartOption {
	return func(c *config) {
		c.url = url
	}
}

func Start(opts ...StartOption) {
	c := defaultConfig()
	for _, fn := range opts {
		fn(c)
	}

	tracer.Start(
		tracer.WithServiceName(c.serviceName),
		tracer.WithAgentAddr(c.url),
		tracer.WithAccessToken(c.accessToken))
}

func Stop() {
	tracer.Stop()
}

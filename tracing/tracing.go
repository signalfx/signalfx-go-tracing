package tracing

import (
	"os"
	"strconv"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/opentracer"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
)

const (
	signalfxServiceName            = "SIGNALFX_SERVICE_NAME"
	signalfxEndpointURL            = "SIGNALFX_ENDPOINT_URL"
	signalfxAccessToken            = "SIGNALFX_ACCESS_TOKEN"
	signalfxSpanTags               = "SIGNALFX_SPAN_TAGS"
	signalfxRecordedValueMaxLength = "SIGNALFX_RECORDED_VALUE_MAX_LENGTH"
)

const defaultRecordedValueMaxLength int = 1200

var defaults = map[string]string{
	signalfxServiceName: "unnamed-go-service",
	signalfxEndpointURL: "http://localhost:9080/v1/trace",
	signalfxAccessToken: "",
}

type config struct {
	serviceName string
	accessToken string
	url         string
	// Because there can be multiple global tags added via environment variable
	// or calls to WithGlobalTag, store them in the required StartOption format to
	// call tracer.Start() in the variadic format.
	globalTags []tracer.StartOption

	// flag to disable injecting library tags
	disableLibraryTags bool

	recordedValueMaxLength *int
}

// StartOption is a function that configures an option for Start
type StartOption = func(*config)

func defaultConfig() *config {
	return &config{
		serviceName:            envOrDefault(signalfxServiceName),
		accessToken:            envOrDefault(signalfxAccessToken),
		url:                    envOrDefault(signalfxEndpointURL),
		globalTags:             envGlobalTags(),
		recordedValueMaxLength: envRecordedValueMaxLength(),
	}
}

func envRecordedValueMaxLength() *int {
	val := os.Getenv(signalfxRecordedValueMaxLength)
	num, err := strconv.Atoi(val)
	if err != nil {
		num = defaultRecordedValueMaxLength
	}
	return &num
}

// envGlobalTags extract global tags from the environment variable and parses the value in the expected format
// key1:value1,
func envGlobalTags() []tracer.StartOption {
	var globalTags []tracer.StartOption
	var val string

	if val = os.Getenv(signalfxSpanTags); val == "" {
		return globalTags
	}

	tags := strings.Split(val, ",")
	for _, tag := range tags {
		// TODO: Currently this assumes "<stringb>" where "<stringa>:<stringb>" has no ":" in the
		// string. The TODO is to fix this logic to allow for "<stringb> to have colons, ":', in it.
		pair := strings.Split(tag, ":")
		if len(pair) == 2 {
			key := strings.TrimSpace(pair[0])
			value := strings.TrimSpace(pair[1])
			// Empty keys aren't valid in Zipkin.
			// https://github.com/openzipkin/zipkin-api/blob/d3324ac79d1aa8f5c6f0ea4febb299402e50480f/zipkin-jsonv2.proto#L50-L51
			if key == "" {
				continue
			}
			globalTag := tracer.WithGlobalTag(key, value)
			globalTags = append(globalTags, globalTag)
		}
	}

	return globalTags
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

// WithoutLibraryTags prevents the tracer from injecting
// tracing library metadata as span tags.
func WithoutLibraryTags() StartOption {
	return func(c *config) {
		c.disableLibraryTags = true
	}
}

// WithGlobalTag sets a tag with the given key/value on all spans created by the
// tracer. This option may be used multiple times.
// Note: Since the underlying transport is Zipkin, only values with strings
// are accepted.
func WithGlobalTag(k string, v string) StartOption {
	return func(c *config) {
		globalTag := tracer.WithGlobalTag(k, v)
		c.globalTags = append(c.globalTags, globalTag)
	}
}

// WithRecordedValueMaxLength specifies the maximum length a tag/log value
// can have.
// Values are completely truncated when set to 0.
// Ignored when set to -1.
func WithRecordedValueMaxLength(l int) StartOption {
	return func(c *config) {
		c.recordedValueMaxLength = &l
	}
}

// Start tracing globally
func Start(opts ...StartOption) {
	c := defaultConfig()
	for _, fn := range opts {
		fn(c)
	}

	startOptions := append(c.globalTags, tracer.WithServiceName(c.serviceName))
	startOptions = append(startOptions, tracer.WithZipkin(c.serviceName, c.url, c.accessToken))
	if c.disableLibraryTags {
		startOptions = append(startOptions, tracer.WithoutLibraryTags())
	}
	if c.recordedValueMaxLength != nil {
		startOptions = append(startOptions, tracer.WithTracerRecordedValueMaxLength(*c.recordedValueMaxLength))
	}
	tracer.Start(
		startOptions...,
	)

	opentracing.SetGlobalTracer(opentracer.New())
}

// Stop tracing globally
func Stop() {
	tracer.Stop()
	opentracing.SetGlobalTracer(&opentracing.NoopTracer{})
}

package redigo // import "github.com/signalfx/signalfx-go-tracing/contrib/gomodule/redigo"

import "github.com/signalfx/signalfx-go-tracing/ddtrace"

type dialConfig struct {
	serviceName   string
	analyticsRate float64
	spanOpts      []ddtrace.StartSpanOption
}

// DialOption represents an option that can be passed to Dial.
type DialOption func(*dialConfig)

func defaults(cfg *dialConfig) {
	cfg.serviceName = "redis.conn"
	// cfg.analyticsRate = globalconfig.AnalyticsRate()
}

// WithServiceName sets the given service name for the dialled connection.
func WithServiceName(name string) DialOption {
	return func(cfg *dialConfig) {
		cfg.serviceName = name
	}
}

// WithAnalytics enables Trace Analytics for all started spans.
func WithAnalytics(on bool) DialOption {
	if on {
		return WithAnalyticsRate(1.0)
	}
	return WithAnalyticsRate(0.0)
}

// WithAnalyticsRate sets the sampling rate for Trace Analytics events
// correlated to started spans.
func WithAnalyticsRate(rate float64) DialOption {
	return func(cfg *dialConfig) {
		cfg.analyticsRate = rate
	}
}

// WithSpanOptions defines a set of additional ddtrace.StartSpanOption to be added
// to spans started by the integration.
func WithSpanOptions(opts ...ddtrace.StartSpanOption) DialOption {
	return func(cfg *dialConfig) {
		cfg.spanOpts = append(cfg.spanOpts, opts...)
	}
}

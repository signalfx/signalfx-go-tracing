package tracer

import (
	"io"
	"net"
	"net/http"
	"time"
)

var (
	defaultRoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
)

const (
	defaultHTTPTimeout = time.Second             // defines the current timeout before giving up with the send process
	defaultAddress = "http://localhost:9080/v1/trace"
	// FIXME
	tracerVersion = "0.0.0"
)

type transport interface {
	// send sends the payload p to the agent using the transport set up.
	// It returns a non-nil response body when no error occurred.
	send(p *payload) (body io.ReadCloser, err error)
}

type httpTransport struct {
	traceURL string            // the delivery URL for traces
	client   *http.Client      // the HTTP client used in the POST
	headers  map[string]string // the Transport headers
}

func (*httpTransport) send(p *payload) (body io.ReadCloser, err error) {
	panic("implement me")
}

// newHTTPTransport returns an httpTransport for the given endpoint
func newHTTPTransport(url string, roundTripper http.RoundTripper) *httpTransport {
	// initialize the default EncoderPool with Encoder headers
	defaultHeaders := map[string]string{
		"Content-Type":                  "application/json",
	}
	return &httpTransport{
		traceURL: url,
		client: &http.Client{
			Transport: roundTripper,
			Timeout:   defaultHTTPTimeout,
		},
		headers: defaultHeaders,
	}
}

func newTransport(addr string, roundTripper http.RoundTripper) transport {
	if roundTripper == nil {
		roundTripper = defaultRoundTripper
	}
	return newHTTPTransport(addr, roundTripper)
}

// newDefaultTransport return a default transport for this tracing client
func newDefaultTransport() transport {
	return newHTTPTransport(defaultAddress, defaultRoundTripper)
}

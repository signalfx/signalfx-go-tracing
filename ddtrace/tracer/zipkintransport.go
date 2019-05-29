package tracer

import (
	"io"
	"net/http"
)

type zipkinHTTPTransport struct {
	traceURL string            // the delivery URL for traces
	client   *http.Client      // the HTTP client used in the POST
	headers  map[string]string // the Transport headers
}

func (*zipkinHTTPTransport) send(p *payload) (body io.ReadCloser, err error) {
	panic("implement me")
}

// newHTTPTransport returns an zipkinHTTPTransport for the given endpoint
func newZipkinHTTPTransport(url string, roundTripper http.RoundTripper) *zipkinHTTPTransport {
	// initialize the default EncoderPool with Encoder headers
	defaultHeaders := map[string]string{
		"Content-Type":                  "application/json",
	}
	return &zipkinHTTPTransport{
		traceURL: url,
		client: &http.Client{
			Transport: roundTripper,
			Timeout:   defaultHTTPTimeout,
		},
		headers: defaultHeaders,
	}
}

func newZipkinTransport(addr string, roundTripper http.RoundTripper) transport {
	if roundTripper == nil {
		roundTripper = defaultRoundTripper
	}
	return newHTTPTransport(addr, roundTripper)
}

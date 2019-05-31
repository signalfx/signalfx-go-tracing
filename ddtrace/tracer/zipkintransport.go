package tracer

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

type zipkinHTTPTransport struct {
	traceURL string            // the delivery URL for traces
	client   *http.Client      // the HTTP client used in the POST
	headers  map[string]string // the Transport headers
}

func (t *zipkinHTTPTransport) send(p encoder) (body io.ReadCloser, err error) {
	// prepare the client and send the payload
	req, err := http.NewRequest("POST", t.traceURL, p)
	if err != nil {
		return nil, fmt.Errorf("cannot create http request: %v", err)
	}
	for header, value := range t.headers {
		req.Header.Set(header, value)
	}
	req.Header.Set("Content-Length", strconv.Itoa(p.size()))
	response, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request to %s failed: %s", t.traceURL, err)
	}
	if code := response.StatusCode; code >= 400 {
		msg, err := ioutil.ReadAll(response.Body)
		_ = response.Body.Close()
		txt := http.StatusText(code)
		if err == nil {
			return nil, fmt.Errorf("%s (Status: %s, URL: %s)", msg, txt, t.traceURL)
		}
		return nil, fmt.Errorf("error reading response body: %s (Status: %s, URL: %s)", err, txt, t.traceURL)
	}
	return response.Body, nil
}

// newHTTPTransport returns an zipkinHTTPTransport for the given endpoint
func newZipkinTransport(url string, accessToken string, roundTripper http.RoundTripper) *zipkinHTTPTransport {
	// initialize the default EncoderPool with Encoder headers
	defaultHeaders := map[string]string{
		"Content-Type": "application/json",
	}

	if accessToken != "" {
		defaultHeaders["X-SF-Token"] = accessToken
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

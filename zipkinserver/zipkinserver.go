package zipkinserver

import (
	"github.com/mailru/easyjson"
	traceformat "github.com/signalfx/golib/trace/format"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// ZipkinServer is an embedded Zipkin server
type ZipkinServer struct {
	server *httptest.Server
	spans  traceformat.Trace
	lock   sync.Mutex
}

// URL of the Zipkin server
func (z *ZipkinServer) URL() string {
	return z.server.URL + "/v1/trace"
}

// Stop the embedded Zipkin server
func (z *ZipkinServer) Stop() {
	z.server.Close()
}

// Reset received spans
func (z *ZipkinServer) Reset() {
	z.lock.Lock()
	z.spans = nil
	z.lock.Unlock()
}

// WaitForSpans waits for numSpans to become available
func (z *ZipkinServer) WaitForSpans(t *testing.T, numSpans int) traceformat.Trace {
	start := time.Now()
	deadline := start.Add(3 * time.Second)
	var spans traceformat.Trace

	for time.Now().Before(deadline) {
		z.lock.Lock()
		spans = z.spans
		z.lock.Unlock()

		switch {
		case len(spans) == numSpans:
			return spans
		case len(spans) > numSpans:
			t.Fatalf("received %d spans, expected %d", len(spans), numSpans)
			return nil
		default:
			time.Sleep(250 * time.Millisecond)
		}
	}

	t.Fatalf("timed out waiting for spans, received %d while expecting %d", len(spans), numSpans)
	return nil
}

// Start embedded Zipkin server
func Start() *ZipkinServer {
	zipkin := &ZipkinServer{}
	zipkin.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/trace" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("content-type") != "application/json" {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}

		var trace traceformat.Trace

		if err := easyjson.UnmarshalFromReader(r.Body, &trace); err != nil {
			_, err = io.WriteString(w, err.Error())
			if err != nil {
				// Probably can't successfully write the err to the response so just
				// panic since this is used for testing.
				panic(err)
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		zipkin.lock.Lock()
		zipkin.spans = append(zipkin.spans, trace...)
		zipkin.lock.Unlock()
	}))
	return zipkin
}

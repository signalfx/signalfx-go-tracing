package tracer

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

type mockAPIHandler struct {
	t *testing.T
}

func (m mockAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	require := require.New(m.t)

	header := r.Header.Get("x-sf-token")
	require.NotEmptyf(header, "x-sf-token is missing")
}

func mockAPINewServer(t *testing.T) *httptest.Server {
	handler := mockAPIHandler{t: t}
	server := httptest.NewServer(handler)
	return server
}

func TestZipkinTransportResponse(t *testing.T) {
	for name, tt := range map[string]struct {
		status int
		body   string
		err    string
	}{
		"ok": {
			status: http.StatusOK,
			body:   "Hello world!",
		},
		"bad": {
			status: http.StatusBadRequest,
			body:   strings.Repeat("X", 1002),
			err:    fmt.Sprintf("%s (Status: Bad Request)", strings.Repeat("X", 1000)),
		},
	} {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			ln, err := net.Listen("tcp4", ":0")
			require.NoError(err)
			go http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				w.Write([]byte(tt.body))
			}))
			defer ln.Close()
			transport := newZipkinTransport(
				fmt.Sprintf("http://%s/v1/trace", ln.Addr().String()), "abcd", defaultRoundTripper)
			rc, err := transport.send(newZipkinPayload())
			if tt.err != "" {
				require.Equal(tt.err, err.Error())
				return
			}
			require.NoError(err)
			slurp, err := ioutil.ReadAll(rc)
			rc.Close()
			require.NoError(err)
			require.Equal(tt.body, string(slurp))
		})
	}
}

func (r *recordingRoundTripper) ZipkinRoundTrip(req *http.Request) (*http.Response, error) {
	r.reqs = append(r.reqs, req)
	return defaultRoundTripper.RoundTrip(req)
}

func TestZipkinTransport(t *testing.T) {
	require := require.New(t)

	receiver := mockAPINewServer(t)
	defer receiver.Close()

	customRoundTripper := recordingRoundTripper{}
	transport := newZipkinTransport(receiver.URL, "abcdef", &customRoundTripper)

	p, err := encodeZipkin(getTestTrace(1, 1))
	require.NoError(err)

	_, err = transport.send(p)
	require.NoError(err)

	// make sure our custom round tripper was used
	require.Len(customRoundTripper.reqs, 1)

	req := customRoundTripper.reqs[0]
	require.Equal(strconv.Itoa(p.size()), req.Header.Get("content-length"))
}

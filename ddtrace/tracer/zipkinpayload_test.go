package tracer

import (
	"bytes"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"

	"github.com/mailru/easyjson"
	sfxtrace "github.com/signalfx/golib/trace"
	traceformat "github.com/signalfx/golib/trace/format"

	"github.com/stretchr/testify/require"
)

func encodeZipkin(traces [][]*span) (*zipkinPayload, error) {
	p := newZipkinPayload("test-service")
	for _, trace := range traces {
		if err := p.push(trace); err != nil {
			return nil, err
		}
	}

	return p, nil
}

// TestZipkinPayloadIntegrity tests that whatever we push into the payload
// allows us to read the same content as would have been encoded by
// the codec.
func TestZipkinPayloadIntegrity(t *testing.T) {
	require := require.New(t)
	p := newZipkinPayload("test-service")
	want := new(bytes.Buffer)
	for _, n := range []int{10, 1 << 10, 1 << 15} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			p.reset()
			lists := make(spanLists, n)
			for i := 0; i < n; i++ {
				list := newSpanList(i)
				lists[i] = list
				require.NoError(p.push(list))
			}
			want.Reset()

			var total traceformat.Trace
			count := 0

			for _, lst := range lists {
				for _, span := range p.convertSpans(lst) {
					s := sfxtrace.Span(*span)
					total = append(total, &s)
					count++
				}
			}

			_, err := easyjson.MarshalToWriter(total, want)
			require.NoError(err)
			require.Equal(want.Len(), p.size())
			require.Equal(count, p.itemCount())

			got, err := ioutil.ReadAll(p)
			require.NoError(err)
			require.Equal(want.Bytes(), got)
		})
	}
}

func TestEmptyZipkinPayload(t *testing.T) {
	require := require.New(t)
	p := newZipkinPayload("test-service")
	data, err := ioutil.ReadAll(p)
	require.NoError(err)
	require.Equal("[]", string(data))
}

func TestSpanKindRemovedFromTags(t *testing.T) {
	require := require.New(t)
	payload := newZipkinPayload("test-service")

	spanList := []*span{
		&span{Meta: map[string]string{"span.kind": "server"}},
		&span{Meta: map[string]string{"span.kind": "client"}},
		&span{Meta: map[string]string{"span.kind": "producer"}},
		&span{Meta: map[string]string{"span.kind": "consumer"}},
	}

	converted := payload.convertSpans(spanList)
	require.Equal(map[string]string{}, converted[0].Tags)
	require.Equal("SERVER", *(converted[0].Kind))

	require.Equal(map[string]string{}, converted[1].Tags)
	require.Equal("CLIENT", *(converted[1].Kind))

	require.Equal(map[string]string{}, converted[2].Tags)
	require.Equal("PRODUCER", *(converted[2].Kind))

	require.Equal(map[string]string{}, converted[3].Tags)
	require.Equal("CONSUMER", *(converted[3].Kind))
}

// TestZipkinPayloadDecode ensures that whatever we push into the payload can
// be decoded by the codec.
func TestZipkinPayloadDecode(t *testing.T) {
	require := require.New(t)
	p := newZipkinPayload("test-service")
	for _, n := range []int{10, 1 << 10} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			p.reset()
			for i := 0; i < n; i++ {
				require.NoError(p.push(newSpanList(i)))
			}
			var got traceformat.Trace
			err := easyjson.UnmarshalFromReader(p, &got)
			require.NoError(err)
		})
	}
}

func BenchmarkZipkinPayloadThroughput(b *testing.B) {
	b.Run("10K", benchmarkZipkinPayloadThroughput(1))
	b.Run("100K", benchmarkZipkinPayloadThroughput(10))
	b.Run("1MB", benchmarkZipkinPayloadThroughput(100))
}

// benchmarkPayloadThroughput benchmarks the throughput of the payload by subsequently
// pushing a trace containing count spans of approximately 10KB in size each.
func benchmarkZipkinPayloadThroughput(count int) func(*testing.B) {
	return func(b *testing.B) {
		require := require.New(b)
		p := newZipkinPayload("test-service")
		s := newBasicSpan("X")
		s.Meta["key"] = strings.Repeat("X", 10*1024)
		trace := make(spanList, count)
		for i := 0; i < count; i++ {
			trace[i] = s
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.reset()
			for p.size() < payloadMaxLimit {
				require.NoError(p.push(trace))
			}
		}
	}
}

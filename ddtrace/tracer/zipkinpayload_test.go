package tracer

import (
	"bytes"
	"github.com/mailru/easyjson"
	sfxtrace "github.com/signalfx/golib/trace"
	traceformat "github.com/signalfx/golib/trace/format"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestZipkinPayloadIntegrity tests that whatever we push into the payload
// allows us to read the same content as would have been encoded by
// the codec.
func TestZipkinPayloadIntegrity(t *testing.T) {
	assert := assert.New(t)
	p := newZipkinPayload()
	want := new(bytes.Buffer)
	for _, n := range []int{10, 1 << 10, 1 << 17,
	} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			p.reset()
			lists := make(spanLists, n)
			for i := 0; i < n; i++ {
				list := newSpanList(i)
				lists[i] = list
				assert.NoError(p.push(list))
			}
			want.Reset()

			var total traceformat.Trace

			for _, lst := range lists {
				for _, span := range convertSpans(lst) {
					s := sfxtrace.Span(*span)
					total = append(total, &s)
				}
			}

			_, err := easyjson.MarshalToWriter(total, want)
			assert.NoError(err)
			assert.Equal(want.Len(), p.size())
			assert.Equal(n, p.itemCount())

			got, err := ioutil.ReadAll(p)
			assert.NoError(err)
			assert.Equal(want.Bytes(), got)
		})
	}
}

// TestZipkinPayloadDecode ensures that whatever we push into the payload can
// be decoded by the codec.
func TestZipkinPayloadDecode(t *testing.T) {
	assert := assert.New(t)
	p := newZipkinPayload()
	for _, n := range []int{10, 1 << 10} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			p.reset()
			for i := 0; i < n; i++ {
				assert.NoError(p.push(newSpanList(i)))
			}
			var got traceformat.Trace
			err := easyjson.UnmarshalFromReader(p, &got)
			assert.NoError(err)
		})
	}
}

func BenchmarkZipkinPayloadThroughput(b *testing.B) {
	b.Run("10K", benchmarkPayloadThroughput(1))
	b.Run("100K", benchmarkPayloadThroughput(10))
	b.Run("1MB", benchmarkPayloadThroughput(100))
}

// benchmarkPayloadThroughput benchmarks the throughput of the payload by subsequently
// pushing a trace containing count spans of approximately 10KB in size each.
func benchmarkZipkinPayloadThroughput(count int) func(*testing.B) {
	return func(b *testing.B) {
		assert := assert.New(b)
		p := newZipkinPayload()
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
				assert.NoError(p.push(trace))
			}
		}
	}
}

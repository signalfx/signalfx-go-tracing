package tracer

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mailru/easyjson"
	"github.com/signalfx/golib/pointer"
	sfxtrace "github.com/signalfx/golib/trace"
	"github.com/signalfx/golib/trace/format"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"strings"
)

const (
	spanKind       = "span.kind"
	spanKindServer = "SERVER"
	spanKindClient = "CLIENT"
)

var _ encoder = (*zipkinPayload)(nil)

type zipkinPayload struct {
	buf        bytes.Buffer
	spanCount  int
	reader     *bytes.Reader

}

func (p *zipkinPayload) Read(b []byte) (n int, err error) {
	if p.reader == nil {
		p.buf.WriteByte(']')
		p.reader = bytes.NewReader(p.buf.Bytes())
	}
	return p.reader.Read(b)
}

func newZipkinPayload() *zipkinPayload {
	payload := &zipkinPayload{}
	payload.reset()
	return payload
}

func (p *zipkinPayload) push(t spanList) error {
	if p.reader != nil {
		return errors.New("zipkinPayload must reset before pushing additional traces")
	}
	for _, span := range convertSpans(t) {
		data, err := easyjson.Marshal(span)
		if err != nil {
			return err
		}
		if p.spanCount > 0 {
			p.buf.WriteByte(',')
		}
		p.buf.Write(data)
		p.spanCount++
	}

	return nil
}

func idToHex(id uint64) string {
	return fmt.Sprintf("%016x", id)
}

func idToHexPtr(id uint64) *string {
	if id == 0 {
		return nil
	}
	return pointer.String(idToHex(id))
}

func formatTags(tags map[string]string) {
	if tags["error.type"] != "" || tags["error.msg"] != "" || tags["error.stack"] != "" {
		tags["error"] = "true"
	}

	delete(tags, ext.ServiceName)
	delete(tags, ext.ResourceName)
	delete(tags, ext.Pid)
}

func convertSpans(spans spanList) []*traceformat.Span {
	sfxSpans := make([]*traceformat.Span, 0, len(spans))

	for _, span := range spans {
		sfxSpan := traceformat.Span{}
		localEndpoint := &sfxtrace.Endpoint{ServiceName: pointer.String(span.Service)}
		tags := map[string]string{}

		for key, val := range span.Meta {
			if key == spanKind {
				continue
			}
			tags[key] = val
		}

		sfxSpan.TraceID = idToHex(span.TraceID)
		sfxSpan.Name = pointer.String(span.Name)
		sfxSpan.ParentID = idToHexPtr(span.ParentID)
		sfxSpan.ID = idToHex(span.SpanID)
		sfxSpan.Kind = deriveKind(span)
		sfxSpan.LocalEndpoint = localEndpoint
		sfxSpan.Timestamp = pointer.Int64(span.Start / 1000)
		sfxSpan.Duration = pointer.Int64(span.Duration / 1000)

		if span.Resource != "" && sfxSpan.Kind != nil && *sfxSpan.Kind == spanKindServer {
			sfxSpan.Name = pointer.String(span.Resource)
		}

		if tags["component"] == "" && span.Type != "" {
			tags["component"] = span.Type
		}

		if sfxSpan.Kind != nil && *sfxSpan.Kind != "" {
			tags["span.kind"] = strings.ToLower(*sfxSpan.Kind)
		}

		formatTags(tags)
		sfxSpan.Tags = tags

		sfxSpans = append(sfxSpans, &sfxSpan)
	}

	return sfxSpans
}

func deriveKind(s *span) *string {
	if kind, ok := s.Meta[spanKind]; ok {
		return pointer.String(kind)
	}

	switch s.Type {
	case ext.SpanTypeHTTP:
		return pointer.String(spanKindClient)
	case ext.SpanTypeWeb:
		return pointer.String(spanKindServer)
	}

	return nil
}

func (p *zipkinPayload) itemCount() int {
	return p.spanCount
}

func (p *zipkinPayload) size() int {
	// Pretty hacky. If reader is non-nil then closing ] has been added. Otherwise it's yet to be added.
	size := p.buf.Len()
	if p.reader != nil {
		return size
	}
	return size + 1
}

func (p *zipkinPayload) reset() {
	p.buf.Reset()
	p.buf.WriteByte('[')
	p.reader = nil
	p.spanCount = 0
}

package tracer

import (
	sfxtrace "github.com/signalfx/golib/trace"
	"io"
)

type payload struct {
	count int
	trace sfxtrace.Trace
}

func (*payload) Read(p []byte) (n int, err error) {
	panic("implement me")
}

func newPayload() *payload {
	return &payload{}
}

func (p *payload) push(t spanList) error {
	panic("implement me")
	return nil
}

func (p *payload) itemCount() int {
	return p.count
}

func (p *payload) size() int {
	panic("implement me")
	return -1
}

func (p *payload) reset() {
	panic("implement me")
}

func (p *payload) updateHeader() {
	panic("implement me")
}

var _ io.Reader = (*payload)(nil)

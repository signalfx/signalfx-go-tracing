package memcache_test

import (
	"context"

	"github.com/bradfitz/gomemcache/memcache"
	memcachetrace "github.com/signalfx/signalfx-go-tracing/contrib/bradfitz/gomemcache/memcache"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
)

func Example() {
	span, ctx := tracer.StartSpanFromContext(context.Background(), "parent.request",
		tracer.ServiceName("web"),
		tracer.ResourceName("/home"),
	)
	defer span.Finish()

	mc := memcachetrace.WrapClient(memcache.New("127.0.0.1:11211"))
	// you can use WithContext to set the parent span
	mc.WithContext(ctx).Set(&memcache.Item{Key: "my key", Value: []byte("my value")})

}

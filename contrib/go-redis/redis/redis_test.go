package redis

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"

	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/mocktracer"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"github.com/signalfx/signalfx-go-tracing/internal/globalconfig"
)

const debug = false

func TestMain(m *testing.M) {
	_, ok := os.LookupEnv("INTEGRATION")
	if !ok {
		fmt.Println("--- SKIP: to enable integration test, set the INTEGRATION environment variable")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestClientEvalSha(t *testing.T) {
	opts := &redis.Options{Addr: "127.0.0.1:6379"}
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()

	client := NewClient(opts, WithServiceName("my-redis"))

	sha1 := client.ScriptLoad("return {KEYS[1],KEYS[2],ARGV[1],ARGV[2]}").Val()
	mt.Reset()

	client.EvalSha(sha1, []string{"key1", "key2", "first", "second"})

	spans := mt.FinishedSpans()
	assert.Len(spans, 1)

	span := spans[0]
	assert.Equal("redis.command", span.OperationName())
	assert.Equal(ext.SpanTypeRedis, span.Tag(ext.SpanType))
	assert.Equal(ext.SpanKindClient, span.Tag(ext.SpanKind))
	assert.Equal("my-redis", span.Tag(ext.ServiceName))
	assert.Equal("127.0.0.1", span.Tag(ext.TargetHost))
	assert.Equal("6379", span.Tag(ext.TargetPort))
	assert.Equal("evalsha", span.Tag(ext.ResourceName))
	assert.Equal("redis", span.Tag(ext.DBType))
}

// https://github.com/DataDog/dd-trace-go/issues/387
func TestIssue387(t *testing.T) {
	opts := &redis.Options{Addr: "127.0.0.1:6379"}
	client := NewClient(opts, WithServiceName("my-redis"))
	n := 1000

	client.Set("test_key", "test_value", 0)

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			client.WithContext(context.Background()).Get("test_key").Result()
		}()
	}
	wg.Wait()

	// should not result in a race
}

func TestClient(t *testing.T) {
	opts := &redis.Options{Addr: "127.0.0.1:6379"}
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()

	client := NewClient(opts, WithServiceName("my-redis"))
	client.Set("test_key", "test_value", 0)

	spans := mt.FinishedSpans()
	assert.Len(spans, 1)

	span := spans[0]
	assert.Equal("redis.command", span.OperationName())
	assert.Equal(ext.SpanTypeRedis, span.Tag(ext.SpanType))
	assert.Equal(ext.SpanKindClient, span.Tag(ext.SpanKind))
	assert.Equal("my-redis", span.Tag(ext.ServiceName))
	assert.Equal("127.0.0.1", span.Tag(ext.TargetHost))
	assert.Equal("6379", span.Tag(ext.TargetPort))
	assert.Equal("set test_key ?", span.Tag("redis.raw_command"))
	assert.Equal("2", span.Tag("redis.args_length"))
	assert.Equal("redis", span.Tag(ext.DBType))
}

func TestPipeline(t *testing.T) {
	opts := &redis.Options{Addr: "127.0.0.1:6379"}
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()

	client := NewClient(opts, WithServiceName("my-redis"))
	pipeline := client.Pipeline()
	pipeline.Expire("pipeline_counter", time.Hour)

	// Exec with context test
	pipeline.(*Pipeliner).ExecWithContext(context.Background())

	spans := mt.FinishedSpans()
	assert.Len(spans, 1)

	span := spans[0]
	assert.Equal("redis.command", span.OperationName())
	assert.Equal(ext.SpanTypeRedis, span.Tag(ext.SpanType))
	assert.Equal(ext.SpanKindClient, span.Tag(ext.SpanKind))
	assert.Equal("my-redis", span.Tag(ext.ServiceName))
	assert.Equal("expire pipeline_counter 3600: false\n", span.Tag(ext.ResourceName))
	assert.Equal("127.0.0.1", span.Tag(ext.TargetHost))
	assert.Equal("6379", span.Tag(ext.TargetPort))
	assert.Equal("1", span.Tag("redis.pipeline_length"))
	assert.Equal("redis", span.Tag(ext.DBType))

	mt.Reset()
	pipeline.Expire("pipeline_counter", time.Hour)
	pipeline.Expire("pipeline_counter_1", time.Minute)

	// Rewriting Exec
	pipeline.Exec()

	spans = mt.FinishedSpans()
	assert.Len(spans, 1)

	span = spans[0]
	assert.Equal("redis.command", span.OperationName())
	assert.Equal(ext.SpanTypeRedis, span.Tag(ext.SpanType))
	assert.Equal(ext.SpanKindClient, span.Tag(ext.SpanKind))
	assert.Equal("my-redis", span.Tag(ext.ServiceName))
	assert.Equal("expire pipeline_counter 3600: false\nexpire pipeline_counter_1 60: false\n", span.Tag(ext.ResourceName))
	assert.Equal("2", span.Tag("redis.pipeline_length"))
	assert.Equal("redis", span.Tag(ext.DBType))
}

func TestChildSpan(t *testing.T) {
	opts := &redis.Options{Addr: "127.0.0.1:6379"}
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()

	// Parent span
	client := NewClient(opts, WithServiceName("my-redis"))
	root, ctx := tracer.StartSpanFromContext(context.Background(), "parent.span")
	client = client.WithContext(ctx)
	client.Set("test_key", "test_value", 0)
	root.Finish()

	spans := mt.FinishedSpans()
	assert.Len(spans, 2)

	var child, parent mocktracer.Span
	for _, s := range spans {
		// order of traces in buffer is not garanteed
		switch s.OperationName() {
		case "redis.command":
			child = s
		case "parent.span":
			parent = s
		}
	}
	assert.NotNil(parent)
	assert.NotNil(child)

	assert.Equal(child.ParentID(), parent.SpanID())
	assert.Equal(child.Tag(ext.TargetHost), "127.0.0.1")
	assert.Equal(child.Tag(ext.TargetPort), "6379")
}

func TestMultipleCommands(t *testing.T) {
	opts := &redis.Options{Addr: "127.0.0.1:6379"}
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()

	client := NewClient(opts, WithServiceName("my-redis"))
	client.Set("test_key", "test_value", 0)
	client.Get("test_key")
	client.Incr("int_key")
	client.ClientList()

	spans := mt.FinishedSpans()
	assert.Len(spans, 4)

	// Checking all commands were recorded
	var commands [4]string
	for i := 0; i < 4; i++ {
		commands[i] = spans[i].Tag("redis.raw_command").(string)
	}
	assert.Contains(commands, "set test_key ?")
	assert.Contains(commands, "get test_key: ")
	assert.Contains(commands, "incr int_key: 0")
	assert.Contains(commands, "client list: ")
}

func TestError(t *testing.T) {
	t.Run("wrong-port", func(t *testing.T) {
		opts := &redis.Options{Addr: "127.0.0.1:6378"} // wrong port
		assert := assert.New(t)
		mt := mocktracer.Start()
		defer mt.Stop()

		client := NewClient(opts, WithServiceName("my-redis"))
		_, err := client.Get("key").Result()

		spans := mt.FinishedSpans()
		assert.Len(spans, 1)
		span := spans[0]

		assert.Equal("redis.command", span.OperationName())
		assert.NotNil(err)
		assert.Equal(err, span.Tag(ext.Error))
		assert.Equal("127.0.0.1", span.Tag(ext.TargetHost))
		assert.Equal("6378", span.Tag(ext.TargetPort))
		assert.Equal("get key: ", span.Tag("redis.raw_command"))
	})

	t.Run("nil", func(t *testing.T) {
		opts := &redis.Options{Addr: "127.0.0.1:6379"}
		assert := assert.New(t)
		mt := mocktracer.Start()
		defer mt.Stop()

		client := NewClient(opts, WithServiceName("my-redis"))
		_, err := client.Get("non_existent_key").Result()

		spans := mt.FinishedSpans()
		assert.Len(spans, 1)
		span := spans[0]

		assert.Equal(redis.Nil, err)
		assert.Equal("redis.command", span.OperationName())
		assert.Empty(span.Tag(ext.Error))
		assert.Equal("127.0.0.1", span.Tag(ext.TargetHost))
		assert.Equal("6379", span.Tag(ext.TargetPort))
		assert.Equal("get non_existent_key: ", span.Tag("redis.raw_command"))
	})
}
func TestAnalyticsSettings(t *testing.T) {
	assertRate := func(t *testing.T, mt mocktracer.Tracer, rate interface{}, opts ...ClientOption) {
		client := NewClient(&redis.Options{Addr: "127.0.0.1:6379"}, opts...)
		client.Set("test_key", "test_value", 0)

		spans := mt.FinishedSpans()
		assert.Len(t, spans, 1)
		s := spans[0]
		assert.Equal(t, rate, s.Tag(ext.EventSampleRate))
	}

	t.Run("defaults", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		assertRate(t, mt, nil)
	})

	t.Run("global", func(t *testing.T) {
		t.Skip("global flag disabled")
		mt := mocktracer.Start()
		defer mt.Stop()

		rate := globalconfig.AnalyticsRate()
		defer globalconfig.SetAnalyticsRate(rate)
		globalconfig.SetAnalyticsRate(0.4)

		assertRate(t, mt, 0.4)
	})

	t.Run("enabled", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		assertRate(t, mt, 1.0, WithAnalytics(true))
	})

	t.Run("disabled", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		assertRate(t, mt, nil, WithAnalytics(false))
	})

	t.Run("override", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		rate := globalconfig.AnalyticsRate()
		defer globalconfig.SetAnalyticsRate(rate)
		globalconfig.SetAnalyticsRate(0.4)

		assertRate(t, mt, 0.23, WithAnalyticsRate(0.23))
	})
}

func TestCommandsToStr(t *testing.T) {
	opts := &redis.Options{Addr: "127.0.0.1:6379"}
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()

	cases := []struct {
		command func(*Client)
		tag     string
	}{
		{
			func(c *Client) { c.Get("test_key") },
			"get test_key: ",
		},
		{
			func(c *Client) { c.Do("append") },
			"append",
		},
		{
			func(c *Client) { c.Do("append test_list test_value") },
			"append test_list ?",
		},
		{
			func(c *Client) { c.Append("test_list", "test_value") },
			"append test_list ? ?",
		},
		{
			func(c *Client) { c.Set("test_key", "test_value", 0) },
			"set test_key ?",
		},
		{
			func(c *Client) { c.SetNX("test_key", "test_value", 0) },
			"setnx test_key ? ?",
		},
		{
			func(c *Client) { c.MSet("t1", "v1", "v2") },
			"mset t1 ? ?",
		},
		{
			func(c *Client) { c.MSet("t1", "v1", "v2", "v3", "v4", "v5") },
			"mset t1 ? ? ? ? ?",
		},
		{
			func(c *Client) { c.MSetNX("t1", "v1", "v2", "v3", "v4", "v5") },
			"msetnx t1 ? ? ? ? ? ?",
		},
		{
			func(c *Client) { c.RPush("t2", "v1", "v2") },
			"rpush t2 ? ? ?",
		},
		{
			func(c *Client) { c.RPushX("t2", "v1") },
			"rpushx t2 ? ?",
		},
		{
			func(c *Client) { c.LPush("t2", "v1", "v2") },
			"lpush t2 ? ? ?",
		},
		{
			func(c *Client) { c.LPushX("t2", "v1") },
			"lpushx t2 ? ?",
		},
		{
			func(c *Client) { c.LInsert("t1", "AFTER", "v2", "v9") },
			"linsert t1 AFTER ? ? ?",
		},
		{
			func(c *Client) { c.LSet("t1", 1, "v10") },
			"lset t1 1 ?",
		},
		{
			func(c *Client) { c.SIsMember("t1", "v10") },
			"sismember t1 ? ?",
		},
		{
			func(c *Client) { c.HSet("t1", "field", "value") },
			"hset t1 field ? ?",
		},
		{
			func(c *Client) { c.HSetNX("t1", "field", "value") },
			"hsetnx t1 field ? ?",
		},
		{
			func(c *Client) { c.HMSet("t1", map[string]interface{}{"k1": "v1", "k2": "v2"}) },
			"hmset t1 ? ? ? ?",
		},
	}

	client := NewClient(opts, WithServiceName("my-redis"))
	for _, tc := range cases {
		tc.command(client)
		spans := mt.FinishedSpans()
		assert.Len(spans, 1)
		assert.Equal(tc.tag, spans[0].Tags()["redis.raw_command"])
		mt.Reset()
	}
}

func Test_cmdStrToTagStr(t *testing.T) {

	cases := []struct {
		in, raw, cmd string
		args         int
	}{
		{
			"get test_key: ",
			"get test_key: ",
			"get",
			1,
		},
		{
			"append",
			"append",
			"append",
			0,
		},
		{
			"append test_list tesakjsd",
			"append test_list ?",
			"append",
			2,
		},
		{
			"set test_key 33nkjkj",
			"set test_key ?",
			"set",
			2,
		},
		{
			"gibberish",
			"gibberish",
			"gibberish",
			0,
		},
		{
			"",
			"",
			"",
			0,
		},
	}

	for _, tc := range cases {
		raw, cmd, args := cmdStrToTagStr(tc.in)
		assert.Equal(t, tc.raw, raw)
		assert.Equal(t, tc.cmd, cmd)
		assert.Equal(t, tc.args, args)
	}
}

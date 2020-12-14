// Package redis provides tracing functions for tracing the go-redis/redis package (https://github.com/go-redis/redis).
package redis

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/signalfx/signalfx-go-tracing/ddtrace"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"

	"github.com/go-redis/redis"
)

// Client is used to trace requests to a redis server.
type Client struct {
	*redis.Client
	*params

	mu  sync.RWMutex // guards ctx
	ctx context.Context
}

var _ redis.Cmdable = (*Client)(nil)

// Pipeliner is used to trace pipelines executed on a Redis server.
type Pipeliner struct {
	redis.Pipeliner
	*params

	ctx context.Context
}

var _ redis.Pipeliner = (*Pipeliner)(nil)

// params holds the tracer and a set of parameters which are recorded with every trace.
type params struct {
	host   string
	port   string
	db     string
	config *clientConfig
}

// NewClient returns a new Client that is traced with the default tracer under
// the service name "redis".
func NewClient(opt *redis.Options, opts ...ClientOption) *Client {
	return WrapClient(redis.NewClient(opt), opts...)
}

// WrapClient wraps a given redis.Client with a tracer under the given service name.
func WrapClient(c *redis.Client, opts ...ClientOption) *Client {
	cfg := new(clientConfig)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}
	opt := c.Options()
	host, port, err := net.SplitHostPort(opt.Addr)
	if err != nil {
		host = opt.Addr
		port = "6379"
	}
	params := &params{
		host:   host,
		port:   port,
		db:     strconv.Itoa(opt.DB),
		config: cfg,
	}
	tc := &Client{Client: c, params: params}
	tc.Client.WrapProcess(createWrapperFromClient(tc))
	return tc
}

// Pipeline creates a Pipeline from a Client
func (c *Client) Pipeline() redis.Pipeliner {
	c.mu.RLock()
	ctx := c.ctx
	c.mu.RUnlock()
	return &Pipeliner{c.Client.Pipeline(), c.params, ctx}
}

// ExecWithContext calls Pipeline.Exec(). It ensures that the resulting Redis calls
// are traced, and that emitted spans are children of the given Context.
func (c *Pipeliner) ExecWithContext(ctx context.Context) ([]redis.Cmder, error) {
	return c.execWithContext(ctx)
}

// Exec calls Pipeline.Exec() ensuring that the resulting Redis calls are traced.
func (c *Pipeliner) Exec() ([]redis.Cmder, error) {
	return c.execWithContext(c.ctx)
}

func (c *Pipeliner) execWithContext(ctx context.Context) ([]redis.Cmder, error) {
	p := c.params
	opts := []ddtrace.StartSpanOption{
		tracer.SpanType(ext.SpanTypeRedis),
		tracer.ServiceName(p.config.serviceName),
		tracer.ResourceName("redis"),
		tracer.Tag(ext.DBType, "redis"),
		tracer.Tag(ext.TargetHost, p.host),
		tracer.Tag(ext.TargetPort, p.port),
		tracer.Tag(ext.SpanKind, ext.SpanKindClient),
		tracer.Tag("out.db", p.db),
	}
	if rate := p.config.analyticsRate; rate > 0 {
		opts = append(opts, tracer.Tag(ext.EventSampleRate, rate))
	}
	if ctx == nil {
		ctx = context.Background()
	}
	span, _ := tracer.StartSpanFromContext(ctx, "redis.command", opts...)
	cmds, err := c.Pipeliner.Exec()
	span.SetTag(ext.ResourceName, commandsToString(cmds))
	span.SetTag("redis.pipeline_length", strconv.Itoa(len(cmds)))
	var finishOpts []ddtrace.FinishOption
	if err != redis.Nil {
		finishOpts = append(finishOpts, tracer.WithError(err))
	}
	span.FinishWithOptionsExt(finishOpts...)

	return cmds, err
}

// commandsToString returns a string representation of a slice of redis Commands, separated by newlines.
func commandsToString(cmds []redis.Cmder) string {
	var b bytes.Buffer
	for _, cmd := range cmds {
		b.WriteString(cmderToString(cmd))
		b.WriteString("\n")
	}
	return b.String()
}

// WithContext sets a context on a Client. Use it to ensure that emitted spans have the correct parent.
func (c *Client) WithContext(ctx context.Context) *Client {
	c.mu.Lock()
	c.ctx = ctx
	c.mu.Unlock()
	return c
}

// Context returns the active context in the client.
func (c *Client) Context() context.Context {
	c.mu.RLock()
	ctx := c.ctx
	c.mu.RUnlock()
	return ctx
}

// createWrapperFromClient returns a new createWrapper function which wraps the processor with tracing
// information obtained from the provided Client. To understand this functionality better see the
// documentation for the github.com/go-redis/redis.(*baseClient).WrapProcess function.
func createWrapperFromClient(tc *Client) func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
	return func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			tc.mu.RLock()
			ctx := tc.ctx
			tc.mu.RUnlock()
			raw, cmdName, argsLength := cmderToCmdStrAndLength(cmd)
			p := tc.params
			opts := []ddtrace.StartSpanOption{
				tracer.SpanType(ext.SpanTypeRedis),
				tracer.ServiceName(p.config.serviceName),
				tracer.ResourceName(cmdName),
				tracer.Tag(ext.DBType, "redis"),
				tracer.Tag(ext.TargetHost, p.host),
				tracer.Tag(ext.TargetPort, p.port),
				tracer.Tag(ext.SpanKind, ext.SpanKindClient),
				tracer.Tag("out.db", p.db),
				tracer.Tag("redis.raw_command", raw),
				tracer.Tag("redis.args_length", strconv.Itoa(argsLength)),
			}
			if rate := p.config.analyticsRate; rate > 0 {
				opts = append(opts, tracer.Tag(ext.EventSampleRate, rate))
			}
			if ctx == nil {
				ctx = context.Background()
			}
			span, _ := tracer.StartSpanFromContext(ctx, "redis.command", opts...)
			err := oldProcess(cmd)
			var finishOpts []ddtrace.FinishOption
			if err != redis.Nil {
				finishOpts = append(finishOpts, tracer.WithError(err))
			}
			span.FinishWithOptionsExt(finishOpts...)
			return err
		}
	}
}

func cmderToCmdStrAndLength(cmder redis.Cmder) (string, string, int) {
	return cmdStrToTagStr(_cmderToString(cmder))
}

func cmderToString(cmder redis.Cmder) string {
	raw, _, _ := cmdStrToTagStr(_cmderToString(cmder))
	return raw
}

func cmdStrToTagStr(cmdStr string) (string, string, int) {
	parts := strings.Split(strings.Trim(cmdStr, " "), " ")
	if len(parts) == 0 {
		return cmdStr, cmdStr, 0
	}

	cmd := parts[0]
	argsLength := len(parts) - 1
	numArgs, ok := cmdCaptureArguments[cmd]
	if !ok {
		return cmdStr, cmd, argsLength
	}

	keep := parts
	if len(parts) > numArgs {
		keep = parts[:numArgs]
	}
	remaining := len(parts) - numArgs
	for i := 0; i < remaining; i++ {
		keep = append(keep, "?")
	}
	return strings.Join(keep, " "), cmd, argsLength
}

func _cmderToString(cmd redis.Cmder) string {
	// We want to support multiple versions of the go-redis library. In
	// older versions Cmder implements the Stringer interface, while in
	// newer versions that was removed, and this String method which
	// sometimes returns an error is used instead. By doing a type assertion
	// we can support both versions.
	switch v := cmd.(type) {
	case fmt.Stringer:
		return v.String()
	case interface{ String() (string, error) }:
		str, err := v.String()
		if err == nil {
			return str
		}
	}
	args := cmd.Args()
	if len(args) == 0 {
		return ""
	}
	if str, ok := args[0].(string); ok {
		return str
	}
	return ""
}

// cmdCaptureArguments specifies how many arguments should the instrumentation
// capture into the raw_command or db.statement tag per redis command.
var cmdCaptureArguments = map[string]int{
	"append":    2,
	"set":       2,
	"setnx":     2,
	"setrange":  3,
	"mset":      2,
	"msetnx":    2,
	"rpush":     2,
	"rpushx":    2,
	"lpush":     2,
	"lpushx":    2,
	"linsert":   3,
	"lset":      3,
	"sismember": 2,
	"hset":      3,
	"hsetnx":    3,
	"hmset":     2,
}

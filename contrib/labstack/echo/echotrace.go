// Package echo provides functions to trace the labstack/echo package (https://github.com/labstack/echo).
package echo

import (
	"strconv"

	"github.com/signalfx/signalfx-go-tracing/ddtrace"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"github.com/signalfx/signalfx-go-tracing/internal/utils"

	"github.com/labstack/echo"
)

// Middleware returns echo middleware which will trace incoming requests.
func Middleware(opts ...Option) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		cfg := new(config)
		defaults(cfg)
		for _, fn := range opts {
			fn(cfg)
		}
		return func(c echo.Context) error {
			request := c.Request()
			operationName := c.Path()
			opts := []ddtrace.StartSpanOption{
				tracer.ServiceName(cfg.serviceName),
				tracer.ResourceName(operationName),
				tracer.SpanType(ext.SpanTypeEcho),
				tracer.Tag(ext.SpanKind, ext.SpanKindServer),
				tracer.Tag(ext.HTTPMethod, request.Method),
				tracer.Tag(ext.HTTPURL, utils.GetURL(request)),
			}

			if spanctx, err := tracer.Extract(tracer.HTTPHeadersCarrier(request.Header)); err == nil {
				opts = append(opts, tracer.ChildOf(spanctx))
			}
			span, ctx := tracer.StartSpanFromContext(request.Context(), operationName, opts...)
			defer span.Finish()

			// pass the span through the request context
			c.SetRequest(request.WithContext(ctx))

			// serve the request to the next middleware
			err := next(c)

			span.SetTag(ext.HTTPCode, strconv.Itoa(c.Response().Status))
			if err != nil {
				span.SetTag(ext.Error, err)
			}
			return err
		}
	}
}

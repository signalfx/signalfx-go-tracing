// Package gin provides functions to trace the gin-gonic/gin package (https://github.com/gin-gonic/gin).
package gin // import "github.com/signalfx/signalfx-go-tracing/contrib/gin-gonic/gin"

import (
	"fmt"
	"strconv"

	"github.com/signalfx/signalfx-go-tracing/ddtrace"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"github.com/signalfx/signalfx-go-tracing/internal/utils"

	"github.com/gin-gonic/gin"
)

// Middleware returns middleware that will trace incoming requests.
func Middleware(service string, opts ...Option) gin.HandlerFunc {
	cfg := newConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return func(c *gin.Context) {
		operationName := c.FullPath()
		opts := []ddtrace.StartSpanOption{
			tracer.ServiceName(service),
			tracer.ResourceName(operationName),
			tracer.SpanType(ext.SpanTypeGin),
			tracer.Tag(ext.SpanKind, ext.SpanKindServer),
			tracer.Tag(ext.HTTPMethod, c.Request.Method),
			tracer.Tag(ext.HTTPURL, utils.GetURL(c.Request)),
		}
		if cfg.analyticsRate > 0 {
			opts = append(opts, tracer.Tag(ext.EventSampleRate, cfg.analyticsRate))
		}
		if spanctx, err := tracer.Extract(tracer.HTTPHeadersCarrier(c.Request.Header)); err == nil {
			opts = append(opts, tracer.ChildOf(spanctx))
		}
		span, ctx := tracer.StartSpanFromContext(c.Request.Context(), operationName, opts...)
		defer span.Finish()

		// pass the span through the request context
		c.Request = c.Request.WithContext(ctx)

		// serve the request to the next middleware
		c.Next()

		span.SetTag(ext.HTTPCode, strconv.Itoa(c.Writer.Status()))

		if len(c.Errors) > 0 {
			span.SetTag(ext.Error, c.Errors[0])
		}
	}
}

// HTML will trace the rendering of the template as a child of the span in the given context.
func HTML(c *gin.Context, code int, name string, obj interface{}) {
	span, _ := tracer.StartSpanFromContext(c.Request.Context(), "gin.render.html")
	span.SetTag("go.template", name)
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("error rendering tmpl:%s: %s", name, r)
			span.FinishWithOptionsExt(tracer.WithError(err))
			panic(r)
		} else {
			span.Finish()
		}
	}()
	c.HTML(code, name, obj)
}

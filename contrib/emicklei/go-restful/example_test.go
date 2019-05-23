package restful_test

import (
	"io"
	"log"
	"net/http"

	"github.com/emicklei/go-restful"
	restfultrace "github.com/signalfx/signalfx-go-tracing/contrib/emicklei/go-restful"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
)

// To start tracing requests, add the trace filter to your go-restful router.
func Example() {
	// create new go-restful service
	ws := new(restful.WebService)

	// create the Datadog filter
	filter := restfultrace.FilterFunc(
		restfultrace.WithServiceName("my-service"),
	)

	// use it
	ws.Filter(filter)

	// set endpoint
	ws.Route(ws.GET("/hello").To(
		func(request *restful.Request, response *restful.Response) {
			io.WriteString(response, "world")
		}))
	restful.Add(ws)

	// serve request
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Example_spanFromContext() {
	ws := new(restful.WebService)
	ws.Filter(restfultrace.Filter)

	ws.Route(ws.GET("/image/encode").To(
		func(request *restful.Request, response *restful.Response) {
			// create a child span to track operation timing.
			encodeSpan, _ := tracer.StartSpanFromContext(request.Request.Context(), "image.encode")
			// encode a image
			encodeSpan.Finish()
		}))
}

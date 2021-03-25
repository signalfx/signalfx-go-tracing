# SignalFx Tracing Library for Go

[![GitHub Release](https://img.shields.io/github/v/release/signalfx/signalfx-go-tracing)](https://github.com/signalfx/signalfx-go-tracing/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/signalfx/signalfx-go-tracing.svg)](https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing)
[![go.mod](https://img.shields.io/github/go-mod/go-version/signalfx/signalfx-go-tracing)](go.mod)
[![CircleCI](https://circleci.com/gh/signalfx/signalfx-go-tracing/tree/master.svg?style=svg)](https://circleci.com/gh/signalfx/signalfx-go-tracing/tree/master)

The SignalFx Tracing Library for Go helps you automatically instrument
Go applications with instrumented library helpers and a tracer to capture
and report distributed traces to SignalFx. You can also use the library to
manually instrument Go applications with the OpenTracing API.

## Requirements and supported software

These are the requirements to use the SignalFx Tracing Library for Go:

* Go 1.12+
* [Go Modules](https://golang.org/ref/mod) as dependency management

These are the supported libraries you can instrument:

| Library | Version |
| ------- | ------- |
| [database/sql](contrib/database/sql) | Standard Library |
| [github.com/gin-gonic/gin](contrib/gin-gonic/gin) | v1.6.2 |
| [github.com/gorilla/mux](contrib/gorilla/mux) | v1.7.4 |
| [github.com/labstack/echo](contrib/labstack/echo.v4) | v4.2.1 |
| [go.mongodb.org/mongo-driver](contrib/mongodb/mongo-go-driver) | v1.3.2 |
| [net/http](contrib/net/http) | Standard Library |

Other libraries are available to instrument, but are in beta and aren't
officially supported. You can view all the libraries in the
[contrib](contrib) directory.

## Configuration values

If the default configuration values don't apply for your environment, override them before running the process you instrument.

| Code | Environment Variable | Default Value | Notes |
| ---  | ---                  | ---           | ---   |
| [WithServiceName](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing/#WithServiceName) | `SIGNALFX_SERVICE_NAME` | `SignalFx-Tracing` | The name of the service. |
| [WithEndpointURL](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing/#WithEndpointURL) | `SIGNALFX_ENDPOINT_URL` | `http://localhost:9080/v1/trace` | The URL to send traces to. Send spans to a Smart Agent, OpenTelemetry Collector, or a SignalFx ingest endpoint.  |
| [WithAccessToken](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing/#WithAccessToken) | `SIGNALFX_ACCESS_TOKEN` | none | The access token for your SignalFx organization. |
| [WithGlobalTag](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing/#WithGlobalTag) | `SIGNALFX_SPAN_TAGS` | none | Comma-separated list of tags included in every reported span. For example, "key1:val1,key2:val2". Use only string values for tags.|
| [WithRecordedValueMaxLength](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing/#WithRecordedValueMaxLength) | `SIGNALFX_RECORDED_VALUE_MAX_LENGTH` | 1200 | The maximum number of characters for any Zipkin-encoded tagged or logged value. Behaviour disabled when set to -1. |
| - | `SIGNALFX_SERVER_TIMING_CONTEXT` | none | When set to `true`, adds `Server-Timing` header to HTTP responses for [net/http](contrib/net/http) and [github.com/gorilla/mux](contrib/gorilla/mux) instrumentations. |

## Instrument a Go application

Follow these steps to instrument target libraries with provided instrumentors.

For more information about how to instrument a Go application, see the
[examples](https://github.com/signalfx/tracing-examples/tree/master/signalfx-tracing/signalfx-go-tracing).

1. Import `github.com/signalfx/signalfx-go-tracing` to your `go mod`.
2. Import the instrumentor for the target library you want to instrument and
replace utilities with their traced equivalents.

   Find an instrumentor for each supported target library in the [contrib](contrib)
   directory. Each instrumentor has an `example_test.go` file that demonstrates
   how to instrument a target library. Don't import examples directly in your application.
3. Enable tracing globally with
[tracing.Start()](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing/#Start).
This creates a tracer and registers it as the OpenTracing global tracer.

## Inject trace context in HTTP requests

Link individual log entries with trace IDs and span IDs associated with corresponding events by passing metadata in HTTP reqeusts.

To inject trace context in HTTP requests, follow these steps:

1. Replace the default HTTP multiplexer:

   ```go
   import (
     httptrace "github.com/signalfx/signalfx-go-tracing/contrib/net/http"
     "github.com/signalfx/signalfx-go-tracing/ddtrace/tracer" // global tracer
     "github.com/signalfx/signalfx-go-tracing/tracing" // helper
   )
   ```

2. Create the new HTTP multiplexer and attach it to your HTTP instance:

   ```go
   mux := httptrace.NewServeMux()
   mux.HandleFunc("/", handler) // handler to your function
   // add all other routes
   http.ListenAndServe(":8888", mux)
   ```

3. Use a helper for the tracer to get the span ID and trace ID from each request:

   ```go
   func TraceIdFromCtx(ctx context.Context) (result string) {
     if span, ok := tracer.SpanFromContext(ctx); ok {
       result = tracer.TraceIDHex(span.Context())
     }
     return
   }
   ```

   It cooperates with the global tracer instance.

### Example log injection configuration

This is an example of what injecting span IDs and trace IDs can look like in your environment:

```go
package main

import (
  httptrace "github.com/signalfx/signalfx-go-tracing/contrib/net/http"
  "github.com/signalfx/signalfx-go-tracing/ddtrace/tracer" // global tracer
  "github.com/signalfx/signalfx-go-tracing/tracing" // helper

  "context"
  "fmt"
  "log"
  "net/http"
)

func TraceIdFromCtx(ctx context.Context) (result string) {
  if span, ok := tracer.SpanFromContext(ctx); ok {
    result = tracer.TraceIDHex(span.Context())
  }
  return
}

func handler(w http.ResponseWriter, r *http.Request) {
  ctx := r.Context()
  log.Print("Recevied request for ", r.URL.Path[1:], " ", TraceIdFromCtx(ctx))
}

func main() {
  tracing.Start()
  defer tracing.Stop()
  mux.HandleFunc("/", handler)
  http.ListenAndServe(":8888", mux)
}

```

## Testing

To run integration tests locally, you should set the `INTEGRATION` environment
variable. The dependencies of the integration tests are best run via Docker.
To get an idea about the versions and the set-up take a look at our
[CI config](https://github.com/signalfx/signalfx-go-tracing/blob/master/.circleci/config.yml).

The best way to run the entire test suite is using the
[CircleCI CLI](https://circleci.com/docs/2.0/local-jobs/). Simply run
`circleci build` in the repository root. Note that you might have to increase
the resources dedicated to Docker to around 4GB.

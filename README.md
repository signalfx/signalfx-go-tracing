[![CircleCI](https://circleci.com/gh/signalfx/signalfx-go-tracing/tree/master.svg?style=svg)](https://circleci.com/gh/signalfx/signalfx-go-tracing/tree/master)
[![GoDoc](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing?status.svg)](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing)

# SignalFx Tracing Library for Go

The SignalFx Tracing Library for Go helps you instrument. Go applications
with the OpenTracing API 2.0 to capture and report distributed traces to
SignalFx.

## Requirements and supported software

These are the requirements to use the SignalFx Tracing Library for Go:

* Go 1.12+
* `mod` or `dep` dependency management tool

These are the supported libraries you can instrument:

| Library | Version |
| ------- | ------- |
| [database/sql](contrib/database/sql) | Standard Library |
| [gin-gonic/gin](contrib/gin-gonic/gin) | 1.5.0 |
| [globalsign/mgo](contrib/globalsign/mgo) | r2018.06.15+ |
| [gorilla/mux](contrib/gorilla/mux) | 1.7+ |
| [labstack/echo](contrib/labstack/echo) | 4.0+ |
| [mongodb/mongo-go-driver](contrib/mongodb/mongo-go-driver) | 1.0+ |
| [net/http](contrib/net/http) | Standard Library |

Other libraries are available to instrument, but are in beta and aren't
officially supported. You can view all the libraries in the
[contrib](contrib) directory.

## Configuration values

Set configuration values from environment variables or directly in your code:

| Code | Environment Variable | Default Value | Notes |
| ---  | ---                  | ---           | ---   |
| [WithServiceName](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing/#WithServiceName) | `SIGNALFX_SERVICE_NAME` | `SignalFx-Tracing` | The name of the service. |
| [WithEndpointURL](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing/#WithEndpointURL) | `SIGNALFX_ENDPOINT_URL` | `http://localhost:9080/v1/trace` | The URL to send traces to.  |
| [WithAccessToken](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing/#WithAccessToken) | `SIGNALFX_ACCESS_TOKEN` | none | The access token for your SignalFx organization. |
| [WithGlobalTag](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing/#WithGlobalTag) | `SIGNALFX_SPAN_TAGS` | none | Comma-separated list of tags included in every reported span. For example, "key1:val1,key2:val2". Use only string values for tags.|

## Instrument a Go application

Follow these steps to instrument target libraries with provided wrapper libraries. 

For more information about how to instrument a Go application, see the
[examples](https://github.com/signalfx/tracing-examples/tree/master/signalfx-tracing/signalfx-go-tracing).

1. Add `github.com/signalfx/signalfx-go-tracing` to your `go mod` or `dep`
dependencies.
2. Specify the name of the service you're instrumenting:
   ```bash
    $ export SIGNALFX_SERVICE_NAME = "your_service"
   ```
3. If you're sending trace data to an endpoint different than `http://localhost:9080/v1/trace`,
set the endpoint URL:
   ```bash
   $ export SIGNALFX_ENDPOINT_URL = "your_endpoint"
   ```
4. If you're sending trace data directly to a SignalFx ingest endpoint or to an
OpenTelemetry Collector, set the access token for your SignalFx organization:
   ```bash
   $ export SIGNALFX_ACCESS_TOKEN = "your_access_token"
   ```
5. Import the wrapper library for the target library you want to instrument and
replace utilities with their traced equivalents. 

   Find a wrapper library for each supported target library in the [contrib](contrib)
   directory. Each wrapper library has a `example_test.go` file that demonstrates
   how to instrument a target library. Don't use the example in your application.
6. Enable tracing globally with
[tracing.Start()](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing/#Start).
This creates a tracer and registers it as the OpenTracing global tracer. 

## API

See the SignalFx Tracing Library for Go API on
[godoc](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing).

## Testing

Test locally using the Go toolset. The grpc.v12 integration will fail because
it covers for deprecated methods. This is expected. In the CI environment we
vendor this version of the library inside the integration. Under normal
circumstances this is not something that we want to do, because users using
this integration might be running versions different from the vendored one,
creating hard to debug conflicts.

To run integration tests locally, you should set the `INTEGRATION` environment
variable. The dependencies of the integration tests are best run via Docker.
To get an idea about the versions and the set-up take a look at our
[CI config](https://github.com/signalfx/signalfx-go-tracing/blob/master/.circleci/config.yml).

The best way to run the entire test suite is using the
[CircleCI CLI](https://circleci.com/docs/2.0/local-jobs/). Simply run
`circleci build` in the repository root. Note that you might have to increase
the resources dedicated to Docker to around 4GB.

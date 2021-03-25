# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Migrate from dep to Go Modules. ([#67](https://github.com/signalfx/signalfx-go-tracing/pull/67))
- Instrument `github.com/Shopify/sarama` instead of deprecated `gopkg.in/Shopify/sarama.v1`. ([#67](https://github.com/signalfx/signalfx-go-tracing/pull/67))

## [1.6.2] - 2021-02-12

### Fixed

- Fix MongoDB instrumentation. ([#65](https://github.com/signalfx/signalfx-go-tracing/pull/65))

## [1.6.1] - 2020-12-14

### Added

- Add `db.type` tag to Redis instrumentations. ([#53](https://github.com/signalfx/signalfx-go-tracing/pull/53))

## [1.6.0] - 2020-09-29

### Changed

- go-redis instrumentation will try to substitute any values from commands with `?`. ([#50](https://github.com/signalfx/signalfx-go-tracing/pull/50))

## [1.5.0] - 2020-09-17

### Added

- Add `TraceIDHex` and `SpanIDHex` functions that return IDs from `ddtrace.SpanContext`. ([#48](https://github.com/signalfx/signalfx-go-tracing/pull/48))

## [1.4.2] - 2020-09-14

### Fixed

- Fix grpc instrumentation to correctly add span.kind tags. ([#47](https://github.com/signalfx/signalfx-go-tracing/pull/47))

## [1.4.1] - 2020-09-02

### Fixed

- Fix version number returned in code.

## [1.4.0] - 2020-08-27

Limiting the size of span tags/attributesThis can be done in the following ways:

1. Set SIGNALFX_RECORDED_VALUE_MAX_LENGTH environment variable to a number to apply the setting globally.
2. Pass tracing.WithRecordedValueMaxLength(number) option to tracing.Start() function.
3. Pass tracer.WithRecordedValueMaxLength(number) option to tracer.StartSpan() function.

### Added

- Add support for SIGNALFX_RECORDED_VALUE_MAX_LENGTH. ([#42](https://github.com/signalfx/signalfx-go-tracing/pull/42))

## [1.3.1] - 2020-08-12

### Fixed

- Ensure that redundant span.kind tag is not set for Zipkin and that its kind field is always uppercase. ([#41](https://github.com/signalfx/signalfx-go-tracing/pull/41))

## [1.3.0] - 2020-08-05

In order to support echo/v4, the instrumentation needed to import echo
from the updated v4 path and also provide an alternate import path for
itself.

When using echo/v4, users must import the instrumentation middleware
from:

```go
"github.com/signalfx/signalfx-go-tracing/contrib/labstack/echo.v4"
```

instead of:

```go
"github.com/signalfx/signalfx-go-tracing/contrib/labstack/echo"
```

### Added

- Add support for echo/v4. ([#40](https://github.com/signalfx/signalfx-go-tracing/pull/40))

## [1.2.3] - 2020-07-31

### Fixed

- Fix version number returned in code.

## [1.2.2] - 2020-07-31

### Fixed

- Redis instrumentation correctly adds span.kind=client tag to database operation spans. ([#39](https://github.com/signalfx/signalfx-go-tracing/pull/39))

## [1.2.1] - 2020-07-30

### Added

- If `SIGNALFX_SERVER_TIMING_CONTEXT` environemntal variable is `true`, pass trace context in [traceparent form](https://www.w3.org/TR/trace-context/#traceparent-header) over a [Server-Timing header](https://www.w3.org/TR/server-timing/). ([#34](https://github.com/signalfx/signalfx-go-tracing/pull/34))

### Fixed

- Fix an issue with opentracing-go 1.2 where span.Tracer would not return a compatible opentracing instance. ([#37](https://github.com/signalfx/signalfx-go-tracing/pull/37))

## [1.2.0] - 2020-05-20

### Added

- Add `signalfx.tracing.library` and `signalfx.tracing.version` tags to local root span for each trace. ([#33](https://github.com/signalfx/signalfx-go-tracing/pull/33))

### Changed

- Change default service name to unnamed-go-service. ([#33](https://github.com/signalfx/signalfx-go-tracing/pull/33))

## [1.1.0] - 2020-04-21

Errors will now be recorded in spans as attributes. The following
attributes will be used to represent errors:

- `error`: A boolean field set to true in case an operation resulted in an error.
- `sfx.error.message`: The human readable error message contained by the error.
- `sfx.error.object`: The type of an error e.g., "echo.HTTPError".
- `sfx.error.kind`:  Short form representation of `sfx.error.object`
- `sfx.error.stack`: A stack trace in platform-conventional format; may or may not pertain to an error. E.g.,

```sh
goroutine 43 [running]:
runtime/debug.Stack(0x1502340, 0xc000183380, 0x1590fac)
	/usr/local/go/src/runtime/debug/stack.go:24 +0x9d
github.com/signalfx/signalfx-go-tracing/ddtrace/tracer.(*span).setTagError(0xc0001d4180, 0x15072a0, 0xc000183410, 0xc0001e3af8)
	/Users/olone/Projects/splunk/signalfx-go-tracing/ddtrace/tracer/span.go:184 +0x3a9
github.com/signalfx/signalfx-go-tracing/ddtrace/tracer.(*span).SetTag(0xc0001d4180, 0x158c4d4, 0x5, 0x15072a0, 0xc000183410, 0x0, 0x0)
	/Users/olone/Projects/splunk/signalfx-go-tracing/ddtrace/tracer/span.go:136 +0x51c
github.com/signalfx/signalfx-go-tracing/contrib/labstack/echo.Middleware.func1.1(0x165a840, 0xc0001be380, 0x0, 0x0)
	/Users/olone/Projects/splunk/signalfx-go-tracing/contrib/labstack/echo/echotrace.go:50 +0x603
github.com/labstack/echo.(*Echo).ServeHTTP(0xc0001ce1e0, 0x164c620, 0xc00018a400, 0xc00019e700)
	/Users/olone/go/pkg/mod/github.com/labstack/echo@v3.3.10+incompatible/echo.go:593 +0x222
github.com/signalfx/signalfx-go-tracing/contrib/labstack/echo.TestEchoTracer500Zipkin(0xc0001ae900)
	/Users/olone/Projects/splunk/signalfx-go-tracing/contrib/labstack/echo/echotrace_test.go:224 +0x3da
testing.tRunner(0xc0001ae900, 0x15ae210)
	/usr/local/go/src/testing/testing.go:991 +0xdc
created by testing.(*T).Run
	/usr/local/go/src/testing/testing.go:1042 +0x357
```

### Changed

- Use SignalFx semantic conventions to represent errors in spans as attributes. ([#29](https://github.com/signalfx/signalfx-go-tracing/pull/29))

## [1.0.2] - 2020-04-02

### Added

- Set B3 propagation as default http codec. ([#28](https://github.com/signalfx/signalfx-go-tracing/pull/28))

## [1.0.1] - 2019-10-21

### Added

- Add SpanKind for gRPC Server and Client. ([#25](https://github.com/signalfx/signalfx-go-tracing/pull/25))

[Unreleased]: https://github.com/signalfx/signalfx-go-tracing/compare/v1.6.2...HEAD
[1.6.2]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.6.2
[1.6.1]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.6.1
[1.6.0]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.6.0
[1.5.0]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.5.0
[1.4.2]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.4.2
[1.4.1]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.4.1
[1.4.0]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.4.0
[1.3.1]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.3.1
[1.3.0]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.3.0
[1.2.3]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.2.3
[1.2.2]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.2.2
[1.2.1]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.2.1
[1.2.0]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.2.0
[1.1.0]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.1.0
[1.0.2]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.0.2
[1.0.1]: https://github.com/signalfx/signalfx-go-tracing/releases/tag/v1.0.1

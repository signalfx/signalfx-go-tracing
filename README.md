[![CircleCI](https://circleci.com/gh/signalfx/signalfx-go-tracing/tree/master.svg?style=svg)](https://circleci.com/gh/signalfx/signalfx-go-tracing/tree/master)
[![GoDoc](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing?status.svg)](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing)

### Installing
Add `github.com/signalfx/signalfx-go-tracing` to your `go mod` or `dep` dependencies.

Requires:

* Go 1.12+
* mod or dep

### Configuration
Configuration values can be set either from environmental variables or code:

`SIGNALFX_SERVICE_NAME` / [WithServiceName](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing/#WithServiceName) Name identifying the service as a whole (defaults to `SignalFx-Tracing`)

`SIGNALFX_ENDPOINT_URL` / [WithEndpointURL](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing/#WithEndpointURL) URL to send traces to (defaults to `http://localhost:9080/v1/trace`)

`SIGNALFX_ACCESS_TOKEN` / [WithAccessToken](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing/#WithAccessToken) (no default)

### Getting Started
When your application starts enable tracing globally with
[tracing.Start](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing/#Start).
For Go or 3rd party libraries you will need to use the wrapper libraries
from [contrib](contrib) to have traces automatically emitted.

Here's a basic example of instrumenting Redis:

```go
package main

import (
	"github.com/go-redis/redis"
	redistrace "github.com/signalfx/signalfx-go-tracing/contrib/go-redis/redis"
	"github.com/signalfx/signalfx-go-tracing/tracing"
)

func main() {
	tracing.Start()
	defer tracing.Stop()

	opts := &redis.Options{Addr: "127.0.0.1:6379", Password: "", DB: 0}
	c := redistrace.NewClient(opts)

	c.Set("test_key", "test_value", 0)
}
```

### Instrumentations
These are the currently supported instrumentations with the minimum version requirement.

| Library | Version |
| ------- | ------- |
| [database/sql](contrib/database/sql) | Standard Library |
| [net/http](contrib/net/http) | Standard Library |
| [gin-gonic/gin](contrib/gin-gonic/gin) | 1.4 |
| [globalsign/mgo](contrib/globalsign/mgo) | r2018.06.15 |
| [mongodb/mongo-go-driver](contrib/mongodb/mongo-go-driver) | 1.0 |

### API
The API is documented on [godoc](https://godoc.org/github.com/signalfx/signalfx-go-tracing/tracing).

### Testing
Tests can be run locally using the Go toolset. The grpc.v12 integration will fail (and this is normal), because it covers for deprecated methods. In the CI environment
we vendor this version of the library inside the integration. Under normal circumstances this is not something that we want to do, because users using this integration
might be running versions different from the vendored one, creating hard to debug conflicts.

To run integration tests locally, you should set the `INTEGRATION` environment variable. The dependencies of the integration tests are best run via Docker. To get an
idea about the versions and the set-up take a look at our [CI config](https://github.com/signalfx/signalfx-go-tracing/blob/master/.circleci/config.yml).

The best way to run the entire test suite is using the [CircleCI CLI](https://circleci.com/docs/2.0/local-jobs/). Simply run `circleci build`
in the repository root. Note that you might have to increase the resources dedicated to Docker to around 4GB.

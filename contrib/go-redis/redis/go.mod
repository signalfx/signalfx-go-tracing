module github.com/signalfx/signalfx-go-tracing/contrib/go-redis/redis

go 1.12

require (
	github.com/go-redis/redis v6.15.7+incompatible
	github.com/signalfx/signalfx-go-tracing v1.12.0
	github.com/stretchr/testify v1.8.3
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

module github.com/signalfx/signalfx-go-tracing/contrib/bradfitz/gomemcache/memcache

go 1.12

require (
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b
	github.com/signalfx/signalfx-go-tracing v1.6.1
	github.com/stretchr/testify v1.5.1
)

replace github.com/signalfx/signalfx-go-tracing => ../../../../

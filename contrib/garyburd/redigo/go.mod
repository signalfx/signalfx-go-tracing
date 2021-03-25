module github.com/signalfx/signalfx-go-tracing/contrib/garyburd/redigo

go 1.12

require (
	github.com/garyburd/redigo v1.6.2
	github.com/signalfx/signalfx-go-tracing v1.7.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

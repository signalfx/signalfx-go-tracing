module github.com/signalfx/signalfx-go-tracing/contrib/gomodule/redigo

go 1.12

require (
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/signalfx/signalfx-go-tracing v1.12.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

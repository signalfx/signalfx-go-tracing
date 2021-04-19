module github.com/signalfx/signalfx-go-tracing/contrib/emicklei/go-restful

go 1.12

require (
	github.com/emicklei/go-restful v2.12.0+incompatible
	github.com/signalfx/signalfx-go-tracing v1.9.2
	github.com/stretchr/testify v1.7.0
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

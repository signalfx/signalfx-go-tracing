module github.com/signalfx/signalfx-go-tracing/contrib/julienschmidt/httprouter

go 1.12

require (
	github.com/julienschmidt/httprouter v1.3.0
	github.com/signalfx/signalfx-go-tracing v1.12.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

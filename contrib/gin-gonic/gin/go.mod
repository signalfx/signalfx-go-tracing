module github.com/signalfx/signalfx-go-tracing/contrib/gin-gonic/gin

go 1.12

require (
	github.com/gin-gonic/gin v1.6.2
	github.com/signalfx/signalfx-go-tracing v1.7.0
	github.com/stretchr/testify v1.5.1
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

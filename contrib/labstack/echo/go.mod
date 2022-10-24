module github.com/signalfx/signalfx-go-tracing/contrib/labstack/echo

go 1.12

require (
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/signalfx/signalfx-go-tracing v1.12.0
	github.com/stretchr/testify v1.8.1
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

module github.com/signalfx/signalfx-go-tracing/contrib/gorilla/mux

go 1.12

require (
	github.com/gorilla/mux v1.7.4
	github.com/signalfx/signalfx-go-tracing v1.10.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

module github.com/signalfx/signalfx-go-tracing/contrib/go-chi/chi

go 1.12

require (
	github.com/go-chi/chi v4.1.1+incompatible
	github.com/signalfx/signalfx-go-tracing v1.9.3
	github.com/stretchr/testify v1.7.0
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

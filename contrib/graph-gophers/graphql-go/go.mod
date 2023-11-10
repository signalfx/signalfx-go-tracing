module github.com/signalfx/signalfx-go-tracing/contrib/graph-gophers/graphql-go

go 1.12

require (
	github.com/graph-gophers/graphql-go v1.5.0
	github.com/signalfx/signalfx-go-tracing v1.12.0
	github.com/stretchr/testify v1.7.1
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

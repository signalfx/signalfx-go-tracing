module github.com/signalfx/signalfx-go-tracing/contrib/graph-gophers/graphql-go

go 1.12

require (
	github.com/graph-gophers/graphql-go v0.0.0-20200309224638-dae41bde9ef9
	github.com/signalfx/signalfx-go-tracing v1.9.2
	github.com/stretchr/testify v1.7.0
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

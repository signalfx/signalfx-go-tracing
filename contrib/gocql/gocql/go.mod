module github.com/signalfx/signalfx-go-tracing/contrib/gocql/gocql

go 1.12

require (
	github.com/gocql/gocql v0.0.0-20200410100145-b454769479c6
	github.com/signalfx/signalfx-go-tracing v1.9.1
	github.com/stretchr/testify v1.7.0
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

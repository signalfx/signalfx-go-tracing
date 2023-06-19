module github.com/signalfx/signalfx-go-tracing/contrib/gocql/gocql

go 1.12

require (
	github.com/gocql/gocql v1.5.2
	github.com/signalfx/signalfx-go-tracing v1.12.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

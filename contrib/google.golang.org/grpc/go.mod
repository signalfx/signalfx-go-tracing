module github.com/signalfx/signalfx-go-tracing/contrib/google.golang.org/grpc

go 1.12

require (
	github.com/golang/protobuf v1.4.0
	github.com/signalfx/signalfx-go-tracing v1.12.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.7.0
	google.golang.org/grpc v1.28.1
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

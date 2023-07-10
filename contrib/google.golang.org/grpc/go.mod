module github.com/signalfx/signalfx-go-tracing/contrib/google.golang.org/grpc

go 1.12

require (
	github.com/golang/protobuf v1.5.3
	github.com/signalfx/signalfx-go-tracing v1.12.0
	github.com/stretchr/testify v1.8.3
	golang.org/x/net v0.9.0
	google.golang.org/grpc v1.56.2
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

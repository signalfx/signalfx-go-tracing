module github.com/signalfx/signalfx-go-tracing/contrib/google.golang.org/grpc

go 1.12

require (
	github.com/golang/protobuf v1.5.3
	github.com/signalfx/signalfx-go-tracing v1.12.0
	github.com/stretchr/testify v1.8.1
	golang.org/x/net v0.8.0
	google.golang.org/grpc v1.55.0
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

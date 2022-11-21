module github.com/signalfx/signalfx-go-tracing/contrib/google.golang.org/grpc

go 1.12

require (
	github.com/golang/protobuf v1.5.2
	github.com/signalfx/signalfx-go-tracing v1.12.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20220722155237-a158d28d115b
	google.golang.org/grpc v1.51.0
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

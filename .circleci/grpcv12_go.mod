module github.com/signalfx/signalfx-go-tracing/contrib/google.golang.org/grpc.v12

go 1.12

require (
    github.com/golang/protobuf v1.4.0
    github.com/signalfx/signalfx-go-tracing v1.0.2
    github.com/stretchr/testify v1.5.1
    golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
    google.golang.org/grpc v1.28.1
)

replace (
    google.golang.org/grpc => google.golang.org/grpc v1.2.1
    github.com/signalfx/signalfx-go-tracing => ../../../
)
module github.com/signalfx/signalfx-go-tracing/contrib/google.golang.org/grpc.v12

go 1.12

require (
	github.com/golang/protobuf v1.4.0
	github.com/signalfx/signalfx-go-tracing v1.0.2
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
	google.golang.org/grpc v1.28.1
)

replace (
	github.com/signalfx/signalfx-go-tracing => ../../../
	google.golang.org/grpc => google.golang.org/grpc v1.2.1
	gopkg.in/Shopify/sarama.v1 => github.com/Shopify/sarama v1.26.1
)

module github.com/signalfx/signalfx-go-tracing/contrib/Shopify/sarama

go 1.12

require (
	github.com/Shopify/sarama v1.30.1
	github.com/signalfx/signalfx-go-tracing v1.12.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

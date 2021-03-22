module github.com/signalfx/signalfx-go-tracing/confluentinc/confluent-kafka-go/kafka

go 1.12

require (
	github.com/confluentinc/confluent-kafka-go v1.4.0
	github.com/signalfx/signalfx-go-tracing v1.6.1
	github.com/stretchr/testify v1.5.1
)

replace github.com/signalfx/signalfx-go-tracing => ../../../../

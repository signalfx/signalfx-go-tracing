module github.com/signalfx/signalfx-go-tracing/contrib/mongodb/mongo-go-driver/mongo

go 1.12

require (
	github.com/signalfx/signalfx-go-tracing v1.11.0
	github.com/stretchr/testify v1.7.0
	go.mongodb.org/mongo-driver v1.3.2
)

replace github.com/signalfx/signalfx-go-tracing => ../../../../

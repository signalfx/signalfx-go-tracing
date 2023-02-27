module github.com/signalfx/signalfx-go-tracing/contrib/google.golang.org/api

go 1.12

require (
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/signalfx/signalfx-go-tracing v1.12.0
	github.com/signalfx/signalfx-go-tracing/contrib/net/http v1.12.0
	github.com/stretchr/testify v1.8.1
	golang.org/x/oauth2 v0.5.0
	google.golang.org/api v0.103.0
)

replace (
	github.com/signalfx/signalfx-go-tracing => ../../../
	github.com/signalfx/signalfx-go-tracing/contrib/net/http => ../../net/http
)

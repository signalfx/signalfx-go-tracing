module github.com/signalfx/signalfx-go-tracing/contrib/google.golang.org/api

go 1.12

require (
	github.com/signalfx/signalfx-go-tracing v1.9.2
	github.com/signalfx/signalfx-go-tracing/contrib/net/http v1.9.2
	github.com/stretchr/testify v1.7.0
	golang.org/x/oauth2 v0.0.0-20210413134643-5e61552d6c78
	google.golang.org/api v0.45.0
)

replace (
	github.com/signalfx/signalfx-go-tracing => ../../../
	github.com/signalfx/signalfx-go-tracing/contrib/net/http => ../../net/http
)

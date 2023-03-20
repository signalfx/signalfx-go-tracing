module github.com/signalfx/signalfx-go-tracing/contrib/k8s.io/client-go/kubernetes

go 1.12

require (
	github.com/signalfx/signalfx-go-tracing v1.12.0
	github.com/signalfx/signalfx-go-tracing/contrib/net/http v1.12.0
	github.com/stretchr/testify v1.8.1
	k8s.io/apimachinery v0.27.0-beta.0
	k8s.io/client-go v0.27.0-beta.0
)

replace (
	github.com/signalfx/signalfx-go-tracing => ../../../../
	github.com/signalfx/signalfx-go-tracing/contrib/net/http => ../../../net/http
)

module github.com/signalfx/signalfx-go-tracing/contrib/k8s.io/client-go/kubernetes

go 1.12

require (
	github.com/signalfx/signalfx-go-tracing v1.7.0
	github.com/signalfx/signalfx-go-tracing/contrib/net/http v1.7.0
	github.com/stretchr/testify v1.7.0
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go v0.0.0-20190819141724-e14f31a72a77
)

replace (
	github.com/signalfx/signalfx-go-tracing => ../../../../
	github.com/signalfx/signalfx-go-tracing/contrib/net/http => ../../../net/http
)

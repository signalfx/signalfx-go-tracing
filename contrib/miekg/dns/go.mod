module github.com/signalfx/signalfx-go-tracing/contrib/miekg/dns

go 1.12

require (
	github.com/miekg/dns v1.1.41
	github.com/signalfx/signalfx-go-tracing v1.7.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

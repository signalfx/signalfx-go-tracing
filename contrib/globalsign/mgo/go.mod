module github.com/signalfx/signalfx-go-tracing/contrib/globalsign/mgo

go 1.12

require (
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/signalfx/signalfx-go-tracing v1.11.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

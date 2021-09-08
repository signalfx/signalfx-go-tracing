module github.com/signalfx/signalfx-go-tracing/contrib/olivere/elastic

go 1.12

require (
	github.com/signalfx/signalfx-go-tracing v1.11.0
	github.com/stretchr/testify v1.7.0
	gopkg.in/olivere/elastic.v3 v3.0.75
	gopkg.in/olivere/elastic.v5 v5.0.85
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

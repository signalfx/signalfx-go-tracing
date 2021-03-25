module github.com/signalfx/signalfx-go-tracing/contrib/jinzhu/gorm

go 1.12

require (
	github.com/go-sql-driver/mysql v1.5.0
	github.com/jinzhu/gorm v1.9.12
	github.com/lib/pq v1.10.0
	github.com/signalfx/signalfx-go-tracing v1.7.0
	github.com/signalfx/signalfx-go-tracing/contrib/database/sql v1.7.0
)

replace (
	github.com/signalfx/signalfx-go-tracing => ../../../
	github.com/signalfx/signalfx-go-tracing/contrib/database/sql => ../../database/sql
)

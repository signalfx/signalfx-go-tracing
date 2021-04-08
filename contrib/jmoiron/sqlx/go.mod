module github.com/signalfx/signalfx-go-tracing/contrib/jmoiron/sqlx

go 1.12

require (
	github.com/go-sql-driver/mysql v1.5.0
	github.com/jmoiron/sqlx v1.3.3
	github.com/lib/pq v1.2.0
	github.com/signalfx/signalfx-go-tracing v1.7.0
	github.com/signalfx/signalfx-go-tracing/contrib/database/sql v1.7.0
)

replace (
	github.com/signalfx/signalfx-go-tracing => ../../../
	github.com/signalfx/signalfx-go-tracing/contrib/database/sql => ../../database/sql
)

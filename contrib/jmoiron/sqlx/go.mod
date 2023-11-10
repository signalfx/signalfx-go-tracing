module github.com/signalfx/signalfx-go-tracing/contrib/jmoiron/sqlx

go 1.12

require (
	github.com/go-sql-driver/mysql v1.6.0
	github.com/jmoiron/sqlx v1.3.5
	github.com/lib/pq v1.2.0
	github.com/signalfx/signalfx-go-tracing v1.12.0
	github.com/signalfx/signalfx-go-tracing/contrib/database/sql v1.12.0
)

replace (
	github.com/signalfx/signalfx-go-tracing => ../../../
	github.com/signalfx/signalfx-go-tracing/contrib/database/sql => ../../database/sql
)

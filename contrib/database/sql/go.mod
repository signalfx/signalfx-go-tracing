module github.com/signalfx/signalfx-go-tracing/contrib/database/sql

go 1.12

require (
	github.com/go-sql-driver/mysql v1.5.0
	github.com/lib/pq v1.2.0
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/signalfx/signalfx-go-tracing v1.12.0
	github.com/stretchr/testify v1.7.0
	gotest.tools v2.2.0+incompatible
)

replace github.com/signalfx/signalfx-go-tracing => ../../../

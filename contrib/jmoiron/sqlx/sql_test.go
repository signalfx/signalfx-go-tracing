package sqlx

import (
	"fmt"
	"log"
	"os"
	"testing"

	sqltrace "github.com/signalfx/signalfx-go-tracing/contrib/database/sql"
	"github.com/signalfx/signalfx-go-tracing/contrib/internal/sqltest"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
)

// tableName holds the SQL table that these tests will be run against. It must be unique cross-repo.
const tableName = "testsqlx"

func TestMain(m *testing.M) {
	_, ok := os.LookupEnv("INTEGRATION")
	if !ok {
		fmt.Println("--- SKIP: to enable integration test, set the INTEGRATION environment variable")
		os.Exit(0)
	}
	defer sqltest.Prepare(tableName)()
	os.Exit(m.Run())
}

func TestMySQL(t *testing.T) {
	sqltrace.Register("mysql", &mysql.MySQLDriver{}, sqltrace.WithServiceName("mysql-test"))
	dbx, err := Open("mysql", "test:test@tcp(127.0.0.1:3306)/test")
	if err != nil {
		log.Fatal(err)
	}
	defer dbx.Close()

	testConfig := &sqltest.Config{
		DB:         dbx.DB,
		DBName:     "test",
		DBPort:     3306,
		DBUser:     "test",
		DriverName: "mysql",
		TableName:  tableName,
		ExpectName: "Query",
		ExpectTags: map[string]interface{}{
			ext.ServiceName: "mysql-test",
			ext.SpanType:    ext.SpanTypeSQL,
			ext.TargetHost:  "127.0.0.1",
			ext.TargetPort:  "3306",
			ext.DBUser:      "test",
			ext.DBInstance:  "test",
		},
	}
	sqltest.RunAll(t, testConfig)
}

func TestPostgres(t *testing.T) {
	sqltrace.Register("postgres", &pq.Driver{})
	dbx, err := Open("postgres", "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer dbx.Close()

	testConfig := &sqltest.Config{
		DB:         dbx.DB,
		DBName:     "postgres",
		DBPort:     5432,
		DBUser:     "postgres",
		DriverName: "postgres",
		TableName:  tableName,
		ExpectName: "Query",
		ExpectTags: map[string]interface{}{
			ext.ServiceName: "postgres.db",
			ext.SpanType:    ext.SpanTypeSQL,
			ext.TargetHost:  "127.0.0.1",
			ext.TargetPort:  "5432",
			ext.DBUser:      "postgres",
			ext.DBInstance:  "postgres",
		},
	}
	sqltest.RunAll(t, testConfig)
}

package sqltest // import "github.com/signalfx/signalfx-go-tracing/contrib/internal/sqltest"

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/signalfx/signalfx-go-tracing/contrib/internal/testutil"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/ext"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/mocktracer"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"github.com/signalfx/signalfx-go-tracing/tracing"
	"github.com/signalfx/signalfx-go-tracing/zipkinserver"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Prepare sets up a table with the given name in both the MySQL and Postgres databases and returns
// a teardown function which will drop it.
func Prepare(tableName string) func() {
	queryDrop := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	queryCreate := fmt.Sprintf("CREATE TABLE %s (id integer NOT NULL DEFAULT '0', name text)", tableName)
	mysql, err := sql.Open("mysql", "test:test@tcp(127.0.0.1:3306)/test")
	defer mysql.Close()
	if err != nil {
		log.Fatal(err)
	}
	mysql.Exec(queryDrop)
	mysql.Exec(queryCreate)
	postgres, err := sql.Open("postgres", "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable")
	defer postgres.Close()
	if err != nil {
		log.Fatal(err)
	}
	postgres.Exec(queryDrop)
	postgres.Exec(queryCreate)
	return func() {
		mysql.Exec(queryDrop)
		postgres.Exec(queryDrop)
	}
}

// RunAll applies a sequence of unit tests to check the correct tracing of sql features.
func RunAll(t *testing.T, cfg *Config) {
	cfg.mockTracer = mocktracer.Start()
	defer cfg.mockTracer.Stop()

	for name, test := range map[string]func(*Config) func(*testing.T){
		"Ping":          testPing,
		"Query":         testQuery,
		"Statement":     testStatement,
		"BeginRollback": testBeginRollback,
		"Exec":          testExec,
	} {
		t.Run(name, test(cfg))
	}

	cfg.mockTracer.Stop()
	t.Run("Zipkin", testZipkin(cfg))
}

func testZipkin(cfg *Config) func(t *testing.T) {
	return func(t *testing.T) {
		query := fmt.Sprintf("SELECT id, name FROM %s LIMIT 5", cfg.TableName)

		zipkin := zipkinserver.Start()
		defer zipkin.Stop()

		tracing.Start(
			tracing.WithEndpointURL(zipkin.URL()),
			tracing.WithServiceName("sql-service"),
			tracing.WithoutLibraryTags(),
		)
		defer tracing.Stop()

		t.Run("error", func(t *testing.T) {
			zipkin.Reset()
			assert := assert.New(t)
			require := require.New(t)

			rows, err := cfg.DB.Query("invalid query")
			require.Error(err)
			require.Nil(rows)

			tracer.ForceFlush()
			spans := zipkin.WaitForSpans(t, 1)

			span := spans[0]

			assert.Equal("CLIENT", *span.Kind)
			assert.Equal("Query", *span.Name)
			testutil.AssertSpanWithTags(t, span, map[string]string{
				"component":     "sql",
				"db.type":       cfg.DriverName,
				"db.instance":   cfg.DBName,
				"db.statement":  "invalid query",
				"db.user":       cfg.DBUser,
				"peer.hostname": "127.0.0.1",
				"peer.port":     strconv.Itoa(cfg.DBPort),
			})

			ea := testutil.ErrorAssertion{
				StackContains:  []string{"goroutine"},
				StackMinLength: 50,
			}

			switch cfg.DriverName {
			case "mysql":
				ea.KindEquals = "*mysql.MySQLError"
				ea.MessageContains = "You have an error in your SQL syntax"
				ea.ObjectContains = "&mysql.MySQLError"
			case "postgres":
				ea.KindEquals = "*pq.Error"
				ea.MessageContains = `pq: syntax error at or near "invalid"`
				ea.ObjectContains = "&pq.Error"
			default:
				panic(cfg.DriverName + "unsupported")
			}
			testutil.AssertSpanWithError(t, span, ea)
		})

		t.Run("query", func(t *testing.T) {
			zipkin.Reset()
			assert := assert.New(t)
			require := require.New(t)

			rows, err := cfg.DB.Query(query)
			require.NoError(err)
			defer rows.Close()

			tracer.ForceFlush()
			spans := zipkin.WaitForSpans(t, 1)

			span := spans[0]

			assert.Equal("CLIENT", *span.Kind)
			assert.Equal("Query", *span.Name)
			assert.Equal(map[string]string{
				"component":     "sql",
				"db.type":       cfg.DriverName,
				"db.instance":   cfg.DBName,
				"db.statement":  query,
				"db.user":       cfg.DBUser,
				"peer.hostname": "127.0.0.1",
				"peer.port":     strconv.Itoa(cfg.DBPort),
			}, span.Tags)
		})
	}
}

func testPing(cfg *Config) func(*testing.T) {
	return func(t *testing.T) {
		cfg.mockTracer.Reset()
		assert := assert.New(t)
		err := cfg.DB.Ping()
		assert.Nil(err)
		spans := cfg.mockTracer.FinishedSpans()
		assert.Len(spans, 1)

		span := spans[0]
		assert.Equal("Ping", span.OperationName())
		for k, v := range cfg.ExpectTags {
			assert.Equal(v, span.Tag(k), "Value mismatch on tag %s", k)
		}
	}
}

func testQuery(cfg *Config) func(*testing.T) {
	query := fmt.Sprintf("SELECT id, name FROM %s LIMIT 5", cfg.TableName)
	return func(t *testing.T) {
		cfg.mockTracer.Reset()
		assert := assert.New(t)
		rows, err := cfg.DB.Query(query)
		defer rows.Close()
		assert.Nil(err)

		spans := cfg.mockTracer.FinishedSpans()
		assert.Len(spans, 1)

		span := spans[0]
		assert.Equal(cfg.ExpectName, span.OperationName())
		for k, v := range cfg.ExpectTags {
			assert.Equal(v, span.Tag(k), "Value mismatch on tag %s", k)
		}
	}
}

func testStatement(cfg *Config) func(*testing.T) {
	query := "INSERT INTO %s(name) VALUES(%s)"
	switch cfg.DriverName {
	case "postgres":
		query = fmt.Sprintf(query, cfg.TableName, "$1")
	case "mysql":
		query = fmt.Sprintf(query, cfg.TableName, "?")
	}
	return func(t *testing.T) {
		cfg.mockTracer.Reset()
		assert := assert.New(t)
		stmt, err := cfg.DB.Prepare(query)
		assert.Equal(nil, err)

		spans := cfg.mockTracer.FinishedSpans()
		assert.Len(spans, 1)

		span := spans[0]
		assert.Equal("Prepare", span.OperationName())
		for k, v := range cfg.ExpectTags {
			assert.Equal(v, span.Tag(k), "Value mismatch on tag %s", k)
		}

		cfg.mockTracer.Reset()
		_, err2 := stmt.Exec("New York")
		assert.Equal(nil, err2)

		spans = cfg.mockTracer.FinishedSpans()
		assert.Len(spans, 1)
		span = spans[0]
		assert.Equal("Exec", span.OperationName())
		for k, v := range cfg.ExpectTags {
			assert.Equal(v, span.Tag(k), "Value mismatch on tag %s", k)
		}
	}
}

func testBeginRollback(cfg *Config) func(*testing.T) {
	return func(t *testing.T) {
		cfg.mockTracer.Reset()
		assert := assert.New(t)

		tx, err := cfg.DB.Begin()
		assert.Equal(nil, err)

		spans := cfg.mockTracer.FinishedSpans()
		assert.Len(spans, 1)

		span := spans[0]
		assert.Equal("Begin", span.OperationName())
		for k, v := range cfg.ExpectTags {
			assert.Equal(v, span.Tag(k), "Value mismatch on tag %s", k)
		}

		cfg.mockTracer.Reset()
		err = tx.Rollback()
		assert.Equal(nil, err)

		spans = cfg.mockTracer.FinishedSpans()
		assert.Len(spans, 1)
		span = spans[0]
		assert.Equal("Rollback", span.OperationName())
		for k, v := range cfg.ExpectTags {
			assert.Equal(v, span.Tag(k), "Value mismatch on tag %s", k)
		}
	}
}

func testExec(cfg *Config) func(*testing.T) {
	return func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)
		query := fmt.Sprintf("INSERT INTO %s(name) VALUES('New York')", cfg.TableName)

		parent, ctx := tracer.StartSpanFromContext(context.Background(), "test.parent",
			tracer.ServiceName("test"),
			tracer.ResourceName("parent"),
		)

		cfg.mockTracer.Reset()
		tx, err := cfg.DB.BeginTx(ctx, nil)
		assert.Equal(nil, err)
		_, err = tx.ExecContext(ctx, query)
		assert.Equal(nil, err)
		err = tx.Commit()
		assert.Equal(nil, err)

		parent.Finish() // flush children

		spans := cfg.mockTracer.FinishedSpans()
		assert.Len(spans, 4)

		var span mocktracer.Span
		for _, s := range spans {
			if s.OperationName() == "Exec" && s.Tag(ext.ResourceName) == query {
				span = s
			}
		}
		require.NotNil(span, "span not found")
		for k, v := range cfg.ExpectTags {
			assert.Equal(v, span.Tag(k), "Value mismatch on tag %s", k)
		}
		for _, s := range spans {
			if s.OperationName() == cfg.ExpectName && s.Tag(ext.ResourceName) == "Commit" {
				span = s
			}
		}
		require.NotNil(span, "span not found")
		for k, v := range cfg.ExpectTags {
			assert.Equal(v, span.Tag(k), "Value mismatch on tag %s", k)
		}
	}
}

// Config holds the test configuration.
type Config struct {
	*sql.DB
	mockTracer mocktracer.Tracer
	DriverName string
	TableName  string
	ExpectName string
	ExpectTags map[string]interface{}
	DBName     string
	DBUser     string
	DBPort     int
}

package buntdb_test

import (
	"context"
	"log"

	buntdbtrace "github.com/signalfx/signalfx-go-tracing/contrib/tidwall/buntdb"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
)

func Example() {
	db, err := buntdbtrace.Open(":memory:")
	if err != nil {
		log.Fatal(err)
	}

	// Create a root span, giving name, server and resource.
	_, ctx := tracer.StartSpanFromContext(context.Background(), "my-query",
		tracer.ServiceName("my-db"),
		tracer.ResourceName("initial-access"),
	)

	// use WithContext to associate the span with the parent
	db.WithContext(ctx).
		Update(func(tx *buntdbtrace.Tx) error {
			_, _, err := tx.Set("key", "value", nil)
			return err
		})

	db.View(func(tx *buntdbtrace.Tx) error {
		// you can also use WithContext on the transaction
		val, err := tx.WithContext(ctx).Get("key")
		log.Println(val)
		return err
	})
}

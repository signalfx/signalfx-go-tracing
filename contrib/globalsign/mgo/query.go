package mgo

import (
	"time"

	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"

	"github.com/globalsign/mgo"
)

// Query is an mgo.Query instance along with the data necessary for tracing.
type Query struct {
	*mgo.Query
	cfg  *mongoConfig
	tags map[string]string
}

// Iter invokes and traces Query.Iter
func (q *Query) Iter() *Iter {
	q.tags = setTagsForCommand(q.tags, "query.iter")
	span := newChildSpanFromContext(q.cfg, q.tags)
	iter := q.Query.Iter()
	span.Finish()
	return &Iter{
		Iter: iter,
		cfg:  q.cfg,
	}
}

// All invokes and traces Query.All
func (q *Query) All(result interface{}) error {
	q.tags = setTagsForCommand(q.tags, "query.all")
	span := newChildSpanFromContext(q.cfg, q.tags)
	err := q.Query.All(result)
	span.Finish(tracer.WithError(err))
	return err
}

// Apply invokes and traces Query.Apply. Triggered by update/upsert or remove
func (q *Query) Apply(change mgo.Change, result interface{}) (info *mgo.ChangeInfo, err error) {
	q.tags = setTagsForCommand(q.tags, getChangeCommand(info))
	span := newChildSpanFromContext(q.cfg, q.tags)
	info, err = q.Query.Apply(change, result)
	span.Finish(tracer.WithError(err))
	return info, err
}

// Count invokes and traces Query.Count
func (q *Query) Count() (n int, err error) {
	q.tags = setTagsForCommand(q.tags, "query.count")
	span := newChildSpanFromContext(q.cfg, q.tags)
	n, err = q.Query.Count()
	span.Finish(tracer.WithError(err))
	return n, err
}

// Distinct invokes and traces Query.Distinct
func (q *Query) Distinct(key string, result interface{}) error {
	q.tags = setTagsForCommand(q.tags, "query.distinct")
	span := newChildSpanFromContext(q.cfg, q.tags)
	err := q.Query.Distinct(key, result)
	span.Finish(tracer.WithError(err))
	return err
}

// Explain invokes and traces Query.Explain
func (q *Query) Explain(result interface{}) error {
	q.tags = setTagsForCommand(q.tags, "query.explain")
	span := newChildSpanFromContext(q.cfg, q.tags)
	err := q.Query.Explain(result)
	span.Finish(tracer.WithError(err))
	return err
}

// For invokes and traces Query.For
func (q *Query) For(result interface{}, f func() error) error {
	q.tags = setTagsForCommand(q.tags, "query.for")
	span := newChildSpanFromContext(q.cfg, q.tags)
	err := q.Query.For(result, f)
	span.Finish(tracer.WithError(err))
	return err
}

// MapReduce invokes and traces Query.MapReduce
func (q *Query) MapReduce(job *mgo.MapReduce, result interface{}) (info *mgo.MapReduceInfo, err error) {
	q.tags = setTagsForCommand(q.tags, "query.mapreduce")
	span := newChildSpanFromContext(q.cfg, q.tags)
	info, err = q.Query.MapReduce(job, result)
	span.Finish(tracer.WithError(err))
	return info, err
}

// One invokes and traces Query.One
func (q *Query) One(result interface{}) error {
	q.tags = setTagsForCommand(q.tags, "query.findone")
	span := newChildSpanFromContext(q.cfg, q.tags)
	err := q.Query.One(result)
	span.Finish(tracer.WithError(err))
	return err
}

// Tail invokes and traces Query.Tail
func (q *Query) Tail(timeout time.Duration) *Iter {
	q.tags = setTagsForCommand(q.tags, "query.tail")
	span := newChildSpanFromContext(q.cfg, q.tags)
	iter := q.Query.Tail(timeout)
	span.Finish()
	return &Iter{
		Iter: iter,
		cfg:  q.cfg,
	}
}

func getChangeCommand(change *mgo.ChangeInfo) string {
	if change.Updated > 0 {
		return "query.update"
	}
	if change.Removed > 0 {
		return "query.remove"
	}
	return "query.upsert"
}

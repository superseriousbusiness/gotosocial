package bun

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun/schema"
)

type RawQuery struct {
	baseQuery

	query   string
	args    []interface{}
	comment string
}

func NewRawQuery(db *DB, query string, args ...interface{}) *RawQuery {
	return &RawQuery{
		baseQuery: baseQuery{
			db: db,
		},
		query: query,
		args:  args,
	}
}

func (q *RawQuery) Conn(db IConn) *RawQuery {
	q.setConn(db)
	return q
}

func (q *RawQuery) Err(err error) *RawQuery {
	q.setErr(err)
	return q
}

func (q *RawQuery) Exec(ctx context.Context, dest ...interface{}) (sql.Result, error) {
	return q.scanOrExec(ctx, dest, len(dest) > 0)
}

func (q *RawQuery) Scan(ctx context.Context, dest ...interface{}) error {
	_, err := q.scanOrExec(ctx, dest, true)
	return err
}

// Comment adds a comment to the query, wrapped by /* ... */.
func (q *RawQuery) Comment(comment string) *RawQuery {
	q.comment = comment
	return q
}

func (q *RawQuery) scanOrExec(
	ctx context.Context, dest []interface{}, hasDest bool,
) (sql.Result, error) {
	if q.err != nil {
		return nil, q.err
	}

	var model Model
	var err error

	if hasDest {
		model, err = q.getModel(dest)
		if err != nil {
			return nil, err
		}
	}

	// if a comment is propagated via the context, use it
	setCommentFromContext(ctx, q)

	query := q.db.format(q.query, q.args)
	var res sql.Result

	if hasDest {
		res, err = q.scan(ctx, q, query, model, hasDest)
	} else {
		res, err = q.exec(ctx, q, query)
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (q *RawQuery) AppendQuery(fmter schema.Formatter, b []byte) ([]byte, error) {
	b = appendComment(b, q.comment)

	return fmter.AppendQuery(b, q.query, q.args...), nil
}

func (q *RawQuery) Operation() string {
	return "SELECT"
}

func (q *RawQuery) String() string {
	buf, err := q.AppendQuery(q.db.Formatter(), nil)
	if err != nil {
		panic(err)
	}

	return string(buf)
}

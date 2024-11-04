package bun

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/dialect/feature"
	"github.com/uptrace/bun/dialect/sqltype"
	"github.com/uptrace/bun/internal"
	"github.com/uptrace/bun/schema"
)

type CreateTableQuery struct {
	baseQuery

	temp        bool
	ifNotExists bool
	fksFromRel  bool // Create foreign keys captured in table's relations.

	// varchar changes the default length for VARCHAR columns.
	// Because some dialects require that length is always specified for VARCHAR type,
	// we will use the exact user-defined type if length is set explicitly, as in `bun:",type:varchar(5)"`,
	// but assume the new default length when it's omitted, e.g. `bun:",type:varchar"`.
	varchar int

	fks         []schema.QueryWithArgs
	partitionBy schema.QueryWithArgs
	tablespace  schema.QueryWithArgs
}

var _ Query = (*CreateTableQuery)(nil)

func NewCreateTableQuery(db *DB) *CreateTableQuery {
	q := &CreateTableQuery{
		baseQuery: baseQuery{
			db:   db,
			conn: db.DB,
		},
		varchar: db.Dialect().DefaultVarcharLen(),
	}
	return q
}

func (q *CreateTableQuery) Conn(db IConn) *CreateTableQuery {
	q.setConn(db)
	return q
}

func (q *CreateTableQuery) Model(model interface{}) *CreateTableQuery {
	q.setModel(model)
	return q
}

func (q *CreateTableQuery) Err(err error) *CreateTableQuery {
	q.setErr(err)
	return q
}

// ------------------------------------------------------------------------------

func (q *CreateTableQuery) Table(tables ...string) *CreateTableQuery {
	for _, table := range tables {
		q.addTable(schema.UnsafeIdent(table))
	}
	return q
}

func (q *CreateTableQuery) TableExpr(query string, args ...interface{}) *CreateTableQuery {
	q.addTable(schema.SafeQuery(query, args))
	return q
}

func (q *CreateTableQuery) ModelTableExpr(query string, args ...interface{}) *CreateTableQuery {
	q.modelTableName = schema.SafeQuery(query, args)
	return q
}

func (q *CreateTableQuery) ColumnExpr(query string, args ...interface{}) *CreateTableQuery {
	q.addColumn(schema.SafeQuery(query, args))
	return q
}

// ------------------------------------------------------------------------------

func (q *CreateTableQuery) Temp() *CreateTableQuery {
	q.temp = true
	return q
}

func (q *CreateTableQuery) IfNotExists() *CreateTableQuery {
	q.ifNotExists = true
	return q
}

// Varchar sets default length for VARCHAR columns.
func (q *CreateTableQuery) Varchar(n int) *CreateTableQuery {
	if n <= 0 {
		q.setErr(fmt.Errorf("bun: illegal VARCHAR length: %d", n))
		return q
	}
	q.varchar = n
	return q
}

func (q *CreateTableQuery) ForeignKey(query string, args ...interface{}) *CreateTableQuery {
	q.fks = append(q.fks, schema.SafeQuery(query, args))
	return q
}

func (q *CreateTableQuery) PartitionBy(query string, args ...interface{}) *CreateTableQuery {
	q.partitionBy = schema.SafeQuery(query, args)
	return q
}

func (q *CreateTableQuery) TableSpace(tablespace string) *CreateTableQuery {
	q.tablespace = schema.UnsafeIdent(tablespace)
	return q
}

// WithForeignKeys adds a FOREIGN KEY clause for each of the model's existing relations.
func (q *CreateTableQuery) WithForeignKeys() *CreateTableQuery {
	q.fksFromRel = true
	return q
}

// ------------------------------------------------------------------------------

func (q *CreateTableQuery) Operation() string {
	return "CREATE TABLE"
}

func (q *CreateTableQuery) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	if q.err != nil {
		return nil, q.err
	}
	if q.table == nil {
		return nil, errNilModel
	}

	b = append(b, "CREATE "...)
	if q.temp {
		b = append(b, "TEMP "...)
	}
	b = append(b, "TABLE "...)
	if q.ifNotExists && fmter.HasFeature(feature.TableNotExists) {
		b = append(b, "IF NOT EXISTS "...)
	}
	b, err = q.appendFirstTable(fmter, b)
	if err != nil {
		return nil, err
	}

	b = append(b, " ("...)

	for i, field := range q.table.Fields {
		if i > 0 {
			b = append(b, ", "...)
		}

		b = append(b, field.SQLName...)
		b = append(b, " "...)
		b = q.appendSQLType(b, field)
		if field.NotNull && q.db.dialect.Name() != dialect.Oracle {
			b = append(b, " NOT NULL"...)
		}

		if (field.Identity && fmter.HasFeature(feature.GeneratedIdentity)) ||
			(field.AutoIncrement && (fmter.HasFeature(feature.AutoIncrement) || fmter.HasFeature(feature.Identity))) {
			b = q.db.dialect.AppendSequence(b, q.table, field)
		}

		if field.SQLDefault != "" {
			b = append(b, " DEFAULT "...)
			b = append(b, field.SQLDefault...)
		}
	}

	for i, col := range q.columns {
		// Only pre-pend the comma if we are on subsequent iterations, or if there were fields/columns appended before
		// this. This way if we are only appending custom column expressions we will not produce a syntax error with a
		// leading comma.
		if i > 0 || len(q.table.Fields) > 0 {
			b = append(b, ", "...)
		}
		b, err = col.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	}

	// In SQLite AUTOINCREMENT is only valid for INTEGER PRIMARY KEY columns, so it might be that
	// a primary key constraint has already been created in dialect.AppendSequence() call above.
	// See sqldialect.Dialect.AppendSequence() for more details.
	if len(q.table.PKs) > 0 && !bytes.Contains(b, []byte("PRIMARY KEY")) {
		b = q.appendPKConstraint(b, q.table.PKs)
	}
	b = q.appendUniqueConstraints(fmter, b)

	if q.fksFromRel {
		b, err = q.appendFKConstraintsRel(fmter, b)
		if err != nil {
			return nil, err
		}
	}
	b, err = q.appendFKConstraints(fmter, b)
	if err != nil {
		return nil, err
	}

	b = append(b, ")"...)

	if !q.partitionBy.IsZero() {
		b = append(b, " PARTITION BY "...)
		b, err = q.partitionBy.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	}

	if !q.tablespace.IsZero() {
		b = append(b, " TABLESPACE "...)
		b, err = q.tablespace.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

func (q *CreateTableQuery) appendSQLType(b []byte, field *schema.Field) []byte {
	// Most of the time these two will match, but for the cases where DiscoveredSQLType is dialect-specific,
	// e.g. pgdialect would change sqltype.SmallInt to pgTypeSmallSerial for columns that have `bun:",autoincrement"`
	if !strings.EqualFold(field.CreateTableSQLType, field.DiscoveredSQLType) {
		return append(b, field.CreateTableSQLType...)
	}

	// For all common SQL types except VARCHAR, both UserDefinedSQLType and DiscoveredSQLType specify the correct type,
	// and we needn't modify it. For VARCHAR columns, we will stop to check if a valid length has been set in .Varchar(int).
	if !strings.EqualFold(field.CreateTableSQLType, sqltype.VarChar) || q.varchar <= 0 {
		return append(b, field.CreateTableSQLType...)
	}

	if q.db.dialect.Name() == dialect.Oracle {
		b = append(b, "VARCHAR2"...)
	} else {
		b = append(b, sqltype.VarChar...)
	}
	b = append(b, "("...)
	b = strconv.AppendInt(b, int64(q.varchar), 10)
	b = append(b, ")"...)
	return b
}

func (q *CreateTableQuery) appendUniqueConstraints(fmter schema.Formatter, b []byte) []byte {
	unique := q.table.Unique

	keys := make([]string, 0, len(unique))
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if key == "" {
			for _, field := range unique[key] {
				b = q.appendUniqueConstraint(fmter, b, key, field)
			}
			continue
		}
		b = q.appendUniqueConstraint(fmter, b, key, unique[key]...)
	}

	return b
}

func (q *CreateTableQuery) appendUniqueConstraint(
	fmter schema.Formatter, b []byte, name string, fields ...*schema.Field,
) []byte {
	if name != "" {
		b = append(b, ", CONSTRAINT "...)
		b = fmter.AppendIdent(b, name)
	} else {
		b = append(b, ","...)
	}
	b = append(b, " UNIQUE ("...)
	b = appendColumns(b, "", fields)
	b = append(b, ")"...)
	return b
}

// appendFKConstraintsRel appends a FOREIGN KEY clause for each of the model's existing relations.
func (q *CreateTableQuery) appendFKConstraintsRel(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	for _, rel := range q.tableModel.Table().Relations {
		if rel.References() {
			b, err = q.appendFK(fmter, b, schema.QueryWithArgs{
				Query: "(?) REFERENCES ? (?) ? ?",
				Args: []interface{}{
					Safe(appendColumns(nil, "", rel.BasePKs)),
					rel.JoinTable.SQLName,
					Safe(appendColumns(nil, "", rel.JoinPKs)),
					Safe(rel.OnUpdate),
					Safe(rel.OnDelete),
				},
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return b, nil
}

func (q *CreateTableQuery) appendFK(fmter schema.Formatter, b []byte, fk schema.QueryWithArgs) (_ []byte, err error) {
	b = append(b, ", FOREIGN KEY "...)
	return fk.AppendQuery(fmter, b)
}

func (q *CreateTableQuery) appendFKConstraints(
	fmter schema.Formatter, b []byte,
) (_ []byte, err error) {
	for _, fk := range q.fks {
		if b, err = q.appendFK(fmter, b, fk); err != nil {
			return nil, err
		}
	}
	return b, nil
}

func (q *CreateTableQuery) appendPKConstraint(b []byte, pks []*schema.Field) []byte {
	b = append(b, ", PRIMARY KEY ("...)
	b = appendColumns(b, "", pks)
	b = append(b, ")"...)
	return b
}

// ------------------------------------------------------------------------------

func (q *CreateTableQuery) Exec(ctx context.Context, dest ...interface{}) (sql.Result, error) {
	if err := q.beforeCreateTableHook(ctx); err != nil {
		return nil, err
	}

	queryBytes, err := q.AppendQuery(q.db.fmter, q.db.makeQueryBytes())
	if err != nil {
		return nil, err
	}

	query := internal.String(queryBytes)

	res, err := q.exec(ctx, q, query)
	if err != nil {
		return nil, err
	}

	if q.table != nil {
		if err := q.afterCreateTableHook(ctx); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (q *CreateTableQuery) beforeCreateTableHook(ctx context.Context) error {
	if hook, ok := q.table.ZeroIface.(BeforeCreateTableHook); ok {
		if err := hook.BeforeCreateTable(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (q *CreateTableQuery) afterCreateTableHook(ctx context.Context) error {
	if hook, ok := q.table.ZeroIface.(AfterCreateTableHook); ok {
		if err := hook.AfterCreateTable(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (q *CreateTableQuery) String() string {
	buf, err := q.AppendQuery(q.db.Formatter(), nil)
	if err != nil {
		panic(err)
	}

	return string(buf)
}

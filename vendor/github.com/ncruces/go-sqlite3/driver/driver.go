// Package driver provides a database/sql driver for SQLite.
//
// Importing package driver registers a [database/sql] driver named "sqlite3".
// You may also need to import package embed.
//
//	import _ "github.com/ncruces/go-sqlite3/driver"
//	import _ "github.com/ncruces/go-sqlite3/embed"
//
// The data source name for "sqlite3" databases can be a filename or a "file:" [URI].
//
// # Default transaction mode
//
// The [TRANSACTION] mode can be specified using "_txlock":
//
//	sql.Open("sqlite3", "file:demo.db?_txlock=immediate")
//
// Possible values are: "deferred" (the default), "immediate", "exclusive".
// Regardless of "_txlock":
//   - a [linearizable] transaction is always "exclusive";
//   - a [serializable] transaction is always "immediate";
//   - a [read-only] transaction is always "deferred".
//
// # Working with time
//
// The time encoding/decoding format can be specified using "_timefmt":
//
//	sql.Open("sqlite3", "file:demo.db?_timefmt=sqlite")
//
// Possible values are: "auto" (the default), "sqlite", "rfc3339";
//   - "auto" encodes as RFC 3339 and decodes any [format] supported by SQLite;
//   - "sqlite" encodes as SQLite and decodes any [format] supported by SQLite;
//   - "rfc3339" encodes and decodes RFC 3339 only.
//
// If you encode as RFC 3339 (the default),
// consider using the TIME [collating sequence] to produce a time-ordered sequence.
//
// To scan values in other formats, [sqlite3.TimeFormat.Scanner] may be helpful.
// To bind values in other formats, [sqlite3.TimeFormat.Encode] them before binding.
//
// When using a custom time struct, you'll have to implement
// [database/sql/driver.Valuer] and [database/sql.Scanner].
//
// The Value method should ideally encode to a time [format] supported by SQLite.
// This ensures SQL date and time functions work as they should,
// and that your schema works with other SQLite tools.
// [sqlite3.TimeFormat.Encode] may help.
//
// The Scan method needs to take into account that the value it receives can be of differing types.
// It can already be a [time.Time], if the driver decoded the value according to "_timefmt" rules.
// Or it can be a: string, int64, float64, []byte, or nil,
// depending on the column type and what whoever wrote the value.
// [sqlite3.TimeFormat.Decode] may help.
//
// # Setting PRAGMAs
//
// [PRAGMA] statements can be specified using "_pragma":
//
//	sql.Open("sqlite3", "file:demo.db?_pragma=busy_timeout(10000)")
//
// If no PRAGMAs are specified, a busy timeout of 1 minute is set.
//
// Order matters:
// encryption keys, busy timeout and locking mode should be the first PRAGMAs set,
// in that order.
//
// [URI]: https://sqlite.org/uri.html
// [PRAGMA]: https://sqlite.org/pragma.html
// [TRANSACTION]: https://sqlite.org/lang_transaction.html#deferred_immediate_and_exclusive_transactions
// [linearizable]: https://pkg.go.dev/database/sql#TxOptions
// [serializable]: https://pkg.go.dev/database/sql#TxOptions
// [read-only]: https://pkg.go.dev/database/sql#TxOptions
// [format]: https://sqlite.org/lang_datefunc.html#time_values
// [collating sequence]: https://sqlite.org/datatype3.html#collating_sequences
package driver

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/ncruces/go-sqlite3"
	"github.com/ncruces/go-sqlite3/internal/util"
)

// This variable can be replaced with -ldflags:
//
//	go build -ldflags="-X github.com/ncruces/go-sqlite3/driver.driverName=sqlite"
var driverName = "sqlite3"

func init() {
	if driverName != "" {
		sql.Register(driverName, &SQLite{})
	}
}

// Open opens the SQLite database specified by dataSourceName as a [database/sql.DB].
//
// Open accepts zero, one, or two callbacks (nil callbacks are ignored).
// The first callback is called when the driver opens a new connection.
// The second callback is called before the driver closes a connection.
// The [sqlite3.Conn] can be used to execute queries, register functions, etc.
func Open(dataSourceName string, fn ...func(*sqlite3.Conn) error) (*sql.DB, error) {
	if len(fn) > 2 {
		return nil, sqlite3.MISUSE
	}
	var init, term func(*sqlite3.Conn) error
	if len(fn) > 1 {
		term = fn[1]
	}
	if len(fn) > 0 {
		init = fn[0]
	}
	c, err := newConnector(dataSourceName, init, term)
	if err != nil {
		return nil, err
	}
	return sql.OpenDB(c), nil
}

// SQLite implements [database/sql/driver.Driver].
type SQLite struct{}

var (
	// Ensure these interfaces are implemented:
	_ driver.DriverContext = &SQLite{}
)

// Open implements [database/sql/driver.Driver].
func (d *SQLite) Open(name string) (driver.Conn, error) {
	c, err := newConnector(name, nil, nil)
	if err != nil {
		return nil, err
	}
	return c.Connect(context.Background())
}

// OpenConnector implements [database/sql/driver.DriverContext].
func (d *SQLite) OpenConnector(name string) (driver.Connector, error) {
	return newConnector(name, nil, nil)
}

func newConnector(name string, init, term func(*sqlite3.Conn) error) (*connector, error) {
	c := connector{name: name, init: init, term: term}

	var txlock, timefmt string
	if strings.HasPrefix(name, "file:") {
		if _, after, ok := strings.Cut(name, "?"); ok {
			query, err := url.ParseQuery(after)
			if err != nil {
				return nil, err
			}
			txlock = query.Get("_txlock")
			timefmt = query.Get("_timefmt")
			c.pragmas = query.Has("_pragma")
		}
	}

	switch txlock {
	case "", "deferred", "concurrent", "immediate", "exclusive":
		c.txLock = txlock
	default:
		return nil, fmt.Errorf("sqlite3: invalid _txlock: %s", txlock)
	}

	switch timefmt {
	case "":
		c.tmRead = sqlite3.TimeFormatAuto
		c.tmWrite = sqlite3.TimeFormatDefault
	case "sqlite":
		c.tmRead = sqlite3.TimeFormatAuto
		c.tmWrite = sqlite3.TimeFormat3
	case "rfc3339":
		c.tmRead = sqlite3.TimeFormatDefault
		c.tmWrite = sqlite3.TimeFormatDefault
	default:
		c.tmRead = sqlite3.TimeFormat(timefmt)
		c.tmWrite = sqlite3.TimeFormat(timefmt)
	}
	return &c, nil
}

type connector struct {
	init    func(*sqlite3.Conn) error
	term    func(*sqlite3.Conn) error
	name    string
	txLock  string
	tmRead  sqlite3.TimeFormat
	tmWrite sqlite3.TimeFormat
	pragmas bool
}

func (n *connector) Driver() driver.Driver {
	return &SQLite{}
}

func (n *connector) Connect(ctx context.Context) (res driver.Conn, err error) {
	c := &conn{
		txLock:  n.txLock,
		tmRead:  n.tmRead,
		tmWrite: n.tmWrite,
	}

	c.Conn, err = sqlite3.OpenContext(ctx, n.name)
	if err != nil {
		return nil, err
	}
	defer func() {
		if res == nil {
			c.Close()
		}
	}()

	old := c.Conn.SetInterrupt(ctx)
	defer c.Conn.SetInterrupt(old)

	if !n.pragmas {
		err = c.Conn.BusyTimeout(time.Minute)
		if err != nil {
			return nil, err
		}
	}
	if n.init != nil {
		err = n.init(c.Conn)
		if err != nil {
			return nil, err
		}
	}
	if n.pragmas || n.init != nil {
		s, _, err := c.Conn.Prepare(`PRAGMA query_only`)
		if err != nil {
			return nil, err
		}
		defer s.Close()
		if s.Step() && s.ColumnBool(0) {
			c.readOnly = '1'
		} else {
			c.readOnly = '0'
		}
		err = s.Close()
		if err != nil {
			return nil, err
		}
	}
	if n.term != nil {
		err = c.Conn.Trace(sqlite3.TRACE_CLOSE, func(sqlite3.TraceEvent, any, any) error {
			return n.term(c.Conn)
		})
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

// Conn is implemented by the SQLite [database/sql] driver connections.
//
// It can be used to access SQLite features like [online backup]:
//
//	db, err := driver.Open("temp.db")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
//	conn, err := db.Conn(context.TODO())
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer conn.Close()
//
//	err = conn.Raw(func(driverConn any) error {
//		conn := driverConn.(driver.Conn)
//		return conn.Raw().Backup("main", "backup.db")
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
// [online backup]: https://sqlite.org/backup.html
type Conn interface {
	Raw() *sqlite3.Conn
	driver.Conn
	driver.ConnBeginTx
	driver.ConnPrepareContext
}

type conn struct {
	*sqlite3.Conn
	txLock   string
	txReset  string
	tmRead   sqlite3.TimeFormat
	tmWrite  sqlite3.TimeFormat
	readOnly byte
}

var (
	// Ensure these interfaces are implemented:
	_ Conn                 = &conn{}
	_ driver.ExecerContext = &conn{}
)

func (c *conn) Raw() *sqlite3.Conn {
	return c.Conn
}

// Deprecated: use BeginTx instead.
func (c *conn) Begin() (driver.Tx, error) {
	// notest
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	var txLock string
	switch opts.Isolation {
	default:
		return nil, util.IsolationErr
	case driver.IsolationLevel(sql.LevelLinearizable):
		txLock = "exclusive"
	case driver.IsolationLevel(sql.LevelSerializable):
		txLock = "immediate"
	case driver.IsolationLevel(sql.LevelDefault):
		if !opts.ReadOnly {
			txLock = c.txLock
		}
	}

	c.txReset = ``
	txBegin := `BEGIN ` + txLock
	if opts.ReadOnly {
		txBegin += ` ; PRAGMA query_only=on`
		c.txReset = `; PRAGMA query_only=` + string(c.readOnly)
	}

	old := c.Conn.SetInterrupt(ctx)
	defer c.Conn.SetInterrupt(old)

	err := c.Conn.Exec(txBegin)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *conn) Commit() error {
	err := c.Conn.Exec(`COMMIT` + c.txReset)
	if err != nil && !c.Conn.GetAutocommit() {
		c.Rollback()
	}
	return err
}

func (c *conn) Rollback() error {
	err := c.Conn.Exec(`ROLLBACK` + c.txReset)
	if errors.Is(err, sqlite3.INTERRUPT) {
		old := c.Conn.SetInterrupt(context.Background())
		defer c.Conn.SetInterrupt(old)
		err = c.Conn.Exec(`ROLLBACK` + c.txReset)
	}
	return err
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	// notest
	return c.PrepareContext(context.Background(), query)
}

func (c *conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	old := c.Conn.SetInterrupt(ctx)
	defer c.Conn.SetInterrupt(old)

	s, tail, err := c.Conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	if notWhitespace(tail) {
		s.Close()
		return nil, util.TailErr
	}
	return &stmt{Stmt: s, tmRead: c.tmRead, tmWrite: c.tmWrite, inputs: -2}, nil
}

func (c *conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if len(args) != 0 {
		// Slow path.
		return nil, driver.ErrSkip
	}

	if savept, ok := ctx.(*saveptCtx); ok {
		// Called from driver.Savepoint.
		savept.Savepoint = c.Conn.Savepoint()
		return resultRowsAffected(0), nil
	}

	old := c.Conn.SetInterrupt(ctx)
	defer c.Conn.SetInterrupt(old)

	err := c.Conn.Exec(query)
	if err != nil {
		return nil, err
	}

	return newResult(c.Conn), nil
}

func (c *conn) CheckNamedValue(arg *driver.NamedValue) error {
	// Fast path: short circuit argument verification.
	// Arguments will be rejected by conn.ExecContext.
	return nil
}

type stmt struct {
	*sqlite3.Stmt
	tmWrite sqlite3.TimeFormat
	tmRead  sqlite3.TimeFormat
	inputs  int
}

var (
	// Ensure these interfaces are implemented:
	_ driver.StmtExecContext   = &stmt{}
	_ driver.StmtQueryContext  = &stmt{}
	_ driver.NamedValueChecker = &stmt{}
)

func (s *stmt) NumInput() int {
	if s.inputs >= -1 {
		return s.inputs
	}
	n := s.Stmt.BindCount()
	for i := 1; i <= n; i++ {
		if s.Stmt.BindName(i) != "" {
			s.inputs = -1
			return -1
		}
	}
	s.inputs = n
	return n
}

// Deprecated: use ExecContext instead.
func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	// notest
	return s.ExecContext(context.Background(), namedValues(args))
}

// Deprecated: use QueryContext instead.
func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	// notest
	return s.QueryContext(context.Background(), namedValues(args))
}

func (s *stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	err := s.setupBindings(args)
	if err != nil {
		return nil, err
	}

	old := s.Stmt.Conn().SetInterrupt(ctx)
	defer s.Stmt.Conn().SetInterrupt(old)

	err = s.Stmt.Exec()
	s.Stmt.ClearBindings()
	if err != nil {
		return nil, err
	}

	return newResult(s.Stmt.Conn()), nil
}

func (s *stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	err := s.setupBindings(args)
	if err != nil {
		return nil, err
	}
	return &rows{ctx: ctx, stmt: s}, nil
}

func (s *stmt) setupBindings(args []driver.NamedValue) (err error) {
	var ids [3]int
	for _, arg := range args {
		ids := ids[:0]
		if arg.Name == "" {
			ids = append(ids, arg.Ordinal)
		} else {
			for _, prefix := range [...]string{":", "@", "$"} {
				if id := s.Stmt.BindIndex(prefix + arg.Name); id != 0 {
					ids = append(ids, id)
				}
			}
		}

		for _, id := range ids {
			switch a := arg.Value.(type) {
			case bool:
				err = s.Stmt.BindBool(id, a)
			case int:
				err = s.Stmt.BindInt(id, a)
			case int64:
				err = s.Stmt.BindInt64(id, a)
			case float64:
				err = s.Stmt.BindFloat(id, a)
			case string:
				err = s.Stmt.BindText(id, a)
			case []byte:
				err = s.Stmt.BindBlob(id, a)
			case sqlite3.ZeroBlob:
				err = s.Stmt.BindZeroBlob(id, int64(a))
			case time.Time:
				err = s.Stmt.BindTime(id, a, s.tmWrite)
			case util.JSON:
				err = s.Stmt.BindJSON(id, a.Value)
			case util.PointerUnwrap:
				err = s.Stmt.BindPointer(id, util.UnwrapPointer(a))
			case nil:
				err = s.Stmt.BindNull(id)
			default:
				panic(util.AssertErr())
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *stmt) CheckNamedValue(arg *driver.NamedValue) error {
	switch arg.Value.(type) {
	case bool, int, int64, float64, string, []byte,
		time.Time, sqlite3.ZeroBlob,
		util.JSON, util.PointerUnwrap,
		nil:
		return nil
	default:
		return driver.ErrSkip
	}
}

func newResult(c *sqlite3.Conn) driver.Result {
	rows := c.Changes()
	if rows != 0 {
		id := c.LastInsertRowID()
		if id != 0 {
			return result{id, rows}
		}
	}
	return resultRowsAffected(rows)
}

type result struct{ lastInsertId, rowsAffected int64 }

func (r result) LastInsertId() (int64, error) {
	return r.lastInsertId, nil
}

func (r result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

type resultRowsAffected int64

func (r resultRowsAffected) LastInsertId() (int64, error) {
	return 0, nil
}

func (r resultRowsAffected) RowsAffected() (int64, error) {
	return int64(r), nil
}

type rows struct {
	ctx context.Context
	*stmt
	names []string
	types []string
	nulls []bool
	scans []scantype
}

type scantype byte

const (
	_ANY  scantype = iota
	_INT  scantype = scantype(sqlite3.INTEGER)
	_REAL scantype = scantype(sqlite3.FLOAT)
	_TEXT scantype = scantype(sqlite3.TEXT)
	_BLOB scantype = scantype(sqlite3.BLOB)
	_NULL scantype = scantype(sqlite3.NULL)
	_BOOL scantype = iota
	_TIME
)

var (
	// Ensure these interfaces are implemented:
	_ driver.RowsColumnTypeDatabaseTypeName = &rows{}
	_ driver.RowsColumnTypeNullable         = &rows{}
)

func (r *rows) Close() error {
	r.Stmt.ClearBindings()
	return r.Stmt.Reset()
}

func (r *rows) Columns() []string {
	if r.names == nil {
		count := r.Stmt.ColumnCount()
		names := make([]string, count)
		for i := range names {
			names[i] = r.Stmt.ColumnName(i)
		}
		r.names = names
	}
	return r.names
}

func (r *rows) loadColumnMetadata() {
	if r.nulls == nil {
		count := r.Stmt.ColumnCount()
		nulls := make([]bool, count)
		types := make([]string, count)
		scans := make([]scantype, count)
		for i := range nulls {
			if col := r.Stmt.ColumnOriginName(i); col != "" {
				types[i], _, nulls[i], _, _, _ = r.Stmt.Conn().TableColumnMetadata(
					r.Stmt.ColumnDatabaseName(i),
					r.Stmt.ColumnTableName(i),
					col)
				types[i] = strings.ToUpper(types[i])
				// These types are only used before we have rows,
				// and otherwise as type hints.
				// The first few ensure STRICT tables are strictly typed.
				// The other two are type hints for booleans and time.
				switch types[i] {
				case "INT", "INTEGER":
					scans[i] = _INT
				case "REAL":
					scans[i] = _REAL
				case "TEXT":
					scans[i] = _TEXT
				case "BLOB":
					scans[i] = _BLOB
				case "BOOLEAN":
					scans[i] = _BOOL
				case "DATE", "TIME", "DATETIME", "TIMESTAMP":
					scans[i] = _TIME
				}
			}
		}
		r.nulls = nulls
		r.types = types
		r.scans = scans
	}
}

func (r *rows) declType(index int) string {
	if r.types == nil {
		count := r.Stmt.ColumnCount()
		types := make([]string, count)
		for i := range types {
			types[i] = strings.ToUpper(r.Stmt.ColumnDeclType(i))
		}
		r.types = types
	}
	return r.types[index]
}

func (r *rows) ColumnTypeDatabaseTypeName(index int) string {
	r.loadColumnMetadata()
	decltype := r.types[index]
	if len := len(decltype); len > 0 && decltype[len-1] == ')' {
		if i := strings.LastIndexByte(decltype, '('); i >= 0 {
			decltype = decltype[:i]
		}
	}
	return strings.TrimSpace(decltype)
}

func (r *rows) ColumnTypeNullable(index int) (nullable, ok bool) {
	r.loadColumnMetadata()
	if r.nulls[index] {
		return false, true
	}
	return true, false
}

func (r *rows) ColumnTypeScanType(index int) (typ reflect.Type) {
	r.loadColumnMetadata()
	scan := r.scans[index]

	if r.Stmt.Busy() {
		// SQLite is dynamically typed and we now have a row.
		// Always use the type of the value itself,
		// unless the scan type is more specific
		// and can scan the actual value.
		val := scantype(r.Stmt.ColumnType(index))
		useValType := true
		switch {
		case scan == _TIME && val != _BLOB && val != _NULL:
			t := r.Stmt.ColumnTime(index, r.tmRead)
			useValType = t == time.Time{}
		case scan == _BOOL && val == _INT:
			i := r.Stmt.ColumnInt64(index)
			useValType = i != 0 && i != 1
		case scan == _BLOB && val == _NULL:
			useValType = false
		}
		if useValType {
			scan = val
		}
	}

	switch scan {
	case _INT:
		return reflect.TypeOf(int64(0))
	case _REAL:
		return reflect.TypeOf(float64(0))
	case _TEXT:
		return reflect.TypeOf("")
	case _BLOB:
		return reflect.TypeOf([]byte{})
	case _BOOL:
		return reflect.TypeOf(false)
	case _TIME:
		return reflect.TypeOf(time.Time{})
	default:
		return reflect.TypeOf((*any)(nil)).Elem()
	}
}

func (r *rows) Next(dest []driver.Value) error {
	old := r.Stmt.Conn().SetInterrupt(r.ctx)
	defer r.Stmt.Conn().SetInterrupt(old)

	if !r.Stmt.Step() {
		if err := r.Stmt.Err(); err != nil {
			return err
		}
		return io.EOF
	}

	data := unsafe.Slice((*any)(unsafe.SliceData(dest)), len(dest))
	err := r.Stmt.Columns(data...)
	for i := range dest {
		if t, ok := r.decodeTime(i, dest[i]); ok {
			dest[i] = t
		}
	}
	return err
}

func (r *rows) decodeTime(i int, v any) (_ time.Time, ok bool) {
	switch v := v.(type) {
	case int64, float64:
		// could be a time value
	case string:
		if r.tmWrite != "" && r.tmWrite != time.RFC3339 && r.tmWrite != time.RFC3339Nano {
			break
		}
		t, ok := maybeTime(v)
		if ok {
			return t, true
		}
	default:
		return
	}
	switch r.declType(i) {
	case "DATE", "TIME", "DATETIME", "TIMESTAMP":
		// could be a time value
	default:
		return
	}
	t, err := r.tmRead.Decode(v)
	return t, err == nil
}

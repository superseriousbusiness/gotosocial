package otelsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"

	"go.opentelemetry.io/otel/trace"
)

// Open is a wrapper over sql.Open that instruments the sql.DB to record executed queries
// using OpenTelemetry API.
func Open(driverName, dsn string, opts ...Option) (*sql.DB, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	return patchDB(db, dsn, opts...)
}

func patchDB(db *sql.DB, dsn string, opts ...Option) (*sql.DB, error) {
	dbDriver := db.Driver()

	// Close the db since we are about to open a new one.
	if err := db.Close(); err != nil {
		return nil, err
	}

	d := newDriver(dbDriver, opts)

	if _, ok := dbDriver.(driver.DriverContext); ok {
		connector, err := d.OpenConnector(dsn)
		if err != nil {
			return nil, err
		}
		return sqlOpenDB(connector, d.instrum), nil
	}

	return sqlOpenDB(&dsnConnector{
		driver: d,
		dsn:    dsn,
	}, d.instrum), nil
}

// OpenDB is a wrapper over sql.OpenDB that instruments the sql.DB to record executed queries
// using OpenTelemetry API.
func OpenDB(connector driver.Connector, opts ...Option) *sql.DB {
	instrum := newDBInstrum(opts)
	c := newConnector(connector.Driver(), connector, instrum)
	return sqlOpenDB(c, instrum)
}

func sqlOpenDB(connector driver.Connector, instrum *dbInstrum) *sql.DB {
	db := sql.OpenDB(connector)
	ReportDBStatsMetrics(db, WithMeterProvider(instrum.meterProvider), WithAttributes(instrum.attrs...))
	return db
}

type dsnConnector struct {
	driver *otelDriver
	dsn    string
}

func (c *dsnConnector) Connect(ctx context.Context) (driver.Conn, error) {
	var conn driver.Conn
	err := c.driver.instrum.withSpan(ctx, "db.Connect", "",
		func(ctx context.Context, span trace.Span) error {
			var err error
			conn, err = c.driver.Open(c.dsn)
			return err
		})
	return conn, err
}

func (c *dsnConnector) Driver() driver.Driver {
	return c.driver
}

//------------------------------------------------------------------------------

type otelDriver struct {
	driver    driver.Driver
	driverCtx driver.DriverContext
	instrum   *dbInstrum
}

var _ driver.DriverContext = (*otelDriver)(nil)

func newDriver(dr driver.Driver, opts []Option) *otelDriver {
	driverCtx, _ := dr.(driver.DriverContext)
	d := &otelDriver{
		driver:    dr,
		driverCtx: driverCtx,
		instrum:   newDBInstrum(opts),
	}
	return d
}

func (d *otelDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.driver.Open(name)
	if err != nil {
		return nil, err
	}
	return newConn(conn, d.instrum), nil
}

func (d *otelDriver) OpenConnector(dsn string) (driver.Connector, error) {
	connector, err := d.driverCtx.OpenConnector(dsn)
	if err != nil {
		return nil, err
	}
	return newConnector(d, connector, d.instrum), nil
}

//------------------------------------------------------------------------------

type connector struct {
	driver.Connector
	driver  driver.Driver
	instrum *dbInstrum
}

var _ driver.Connector = (*connector)(nil)

func newConnector(d driver.Driver, c driver.Connector, instrum *dbInstrum) *connector {
	return &connector{
		driver:    d,
		Connector: c,
		instrum:   instrum,
	}
}

func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	var conn driver.Conn
	if err := c.instrum.withSpan(ctx, "db.Connect", "",
		func(ctx context.Context, span trace.Span) error {
			var err error
			conn, err = c.Connector.Connect(ctx)
			return err
		}); err != nil {
		return nil, err
	}
	return newConn(conn, c.instrum), nil
}

func (c *connector) Driver() driver.Driver {
	return c.driver
}

//------------------------------------------------------------------------------

type otelConn struct {
	driver.Conn

	instrum *dbInstrum

	ping            pingFunc
	exec            execFunc
	execCtx         execCtxFunc
	query           queryFunc
	queryCtx        queryCtxFunc
	prepareCtx      prepareCtxFunc
	beginTx         beginTxFunc
	resetSession    resetSessionFunc
	checkNamedValue checkNamedValueFunc
}

var _ driver.Conn = (*otelConn)(nil)

func newConn(conn driver.Conn, instrum *dbInstrum) *otelConn {
	cn := &otelConn{
		Conn:    conn,
		instrum: instrum,
	}

	cn.ping = cn.createPingFunc(conn)
	cn.exec = cn.createExecFunc(conn)
	cn.execCtx = cn.createExecCtxFunc(conn)
	cn.query = cn.createQueryFunc(conn)
	cn.queryCtx = cn.createQueryCtxFunc(conn)
	cn.prepareCtx = cn.createPrepareCtxFunc(conn)
	cn.beginTx = cn.createBeginTxFunc(conn)
	cn.resetSession = cn.createResetSessionFunc(conn)
	cn.checkNamedValue = cn.createCheckNamedValueFunc(conn)

	return cn
}

var _ driver.Pinger = (*otelConn)(nil)

func (c *otelConn) Ping(ctx context.Context) error {
	return c.ping(ctx)
}

type pingFunc func(ctx context.Context) error

func (c *otelConn) createPingFunc(conn driver.Conn) pingFunc {
	if pinger, ok := conn.(driver.Pinger); ok {
		return func(ctx context.Context) error {
			return c.instrum.withSpan(ctx, "db.Ping", "",
				func(ctx context.Context, span trace.Span) error {
					return pinger.Ping(ctx)
				})
		}
	}
	return func(ctx context.Context) error {
		return driver.ErrSkip
	}
}

//------------------------------------------------------------------------------

var _ driver.Execer = (*otelConn)(nil)

func (c *otelConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return c.exec(query, args)
}

type execFunc func(query string, args []driver.Value) (driver.Result, error)

func (c *otelConn) createExecFunc(conn driver.Conn) execFunc {
	if execer, ok := conn.(driver.Execer); ok {
		return func(query string, args []driver.Value) (driver.Result, error) {
			return execer.Exec(query, args)
		}
	}
	return func(query string, args []driver.Value) (driver.Result, error) {
		return nil, driver.ErrSkip
	}
}

//------------------------------------------------------------------------------

var _ driver.ExecerContext = (*otelConn)(nil)

func (c *otelConn) ExecContext(
	ctx context.Context, query string, args []driver.NamedValue,
) (driver.Result, error) {
	return c.execCtx(ctx, query, args)
}

type execCtxFunc func(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error)

func (c *otelConn) createExecCtxFunc(conn driver.Conn) execCtxFunc {
	var fn execCtxFunc

	if execer, ok := conn.(driver.ExecerContext); ok {
		fn = execer.ExecContext
	} else {
		fn = func(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
			vArgs, err := namedValueToValue(args)
			if err != nil {
				return nil, err
			}
			return c.exec(query, vArgs)
		}
	}

	return func(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		var res driver.Result
		if err := c.instrum.withSpan(ctx, "db.Exec", query,
			func(ctx context.Context, span trace.Span) error {
				var err error
				res, err = fn(ctx, query, args)
				if err != nil {
					return err
				}

				if span.IsRecording() {
					rows, err := res.RowsAffected()
					if err == nil {
						span.SetAttributes(dbRowsAffected.Int64(rows))
					}
				}

				return nil
			}); err != nil {
			return nil, err
		}
		return res, nil
	}
}

//------------------------------------------------------------------------------

var _ driver.Queryer = (*otelConn)(nil)

func (c *otelConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return c.query(query, args)
}

type queryFunc func(query string, args []driver.Value) (driver.Rows, error)

func (c *otelConn) createQueryFunc(conn driver.Conn) queryFunc {
	if queryer, ok := c.Conn.(driver.Queryer); ok {
		return func(query string, args []driver.Value) (driver.Rows, error) {
			return queryer.Query(query, args)
		}
	}
	return func(query string, args []driver.Value) (driver.Rows, error) {
		return nil, driver.ErrSkip
	}
}

//------------------------------------------------------------------------------

var _ driver.QueryerContext = (*otelConn)(nil)

func (c *otelConn) QueryContext(
	ctx context.Context, query string, args []driver.NamedValue,
) (driver.Rows, error) {
	return c.queryCtx(ctx, query, args)
}

type queryCtxFunc func(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error)

func (c *otelConn) createQueryCtxFunc(conn driver.Conn) queryCtxFunc {
	var fn queryCtxFunc

	if queryer, ok := c.Conn.(driver.QueryerContext); ok {
		fn = queryer.QueryContext
	} else {
		fn = func(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
			vArgs, err := namedValueToValue(args)
			if err != nil {
				return nil, err
			}
			return c.query(query, vArgs)
		}
	}

	return func(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		var rows driver.Rows
		err := c.instrum.withSpan(ctx, "db.Query", query,
			func(ctx context.Context, span trace.Span) error {
				var err error
				rows, err = fn(ctx, query, args)
				return err
			})
		return rows, err
	}
}

//------------------------------------------------------------------------------

var _ driver.ConnPrepareContext = (*otelConn)(nil)

func (c *otelConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return c.prepareCtx(ctx, query)
}

type prepareCtxFunc func(ctx context.Context, query string) (driver.Stmt, error)

func (c *otelConn) createPrepareCtxFunc(conn driver.Conn) prepareCtxFunc {
	var fn prepareCtxFunc

	if preparer, ok := c.Conn.(driver.ConnPrepareContext); ok {
		fn = preparer.PrepareContext
	} else {
		fn = func(ctx context.Context, query string) (driver.Stmt, error) {
			return c.Conn.Prepare(query)
		}
	}

	return func(ctx context.Context, query string) (driver.Stmt, error) {
		var stmt driver.Stmt
		if err := c.instrum.withSpan(ctx, "db.Prepare", query,
			func(ctx context.Context, span trace.Span) error {
				var err error
				stmt, err = fn(ctx, query)
				return err
			}); err != nil {
			return nil, err
		}
		return newStmt(stmt, query, c.instrum), nil
	}
}

var _ driver.ConnBeginTx = (*otelConn)(nil)

func (c *otelConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return c.beginTx(ctx, opts)
}

type beginTxFunc func(ctx context.Context, opts driver.TxOptions) (driver.Tx, error)

func (c *otelConn) createBeginTxFunc(conn driver.Conn) beginTxFunc {
	var fn beginTxFunc

	if txor, ok := conn.(driver.ConnBeginTx); ok {
		fn = txor.BeginTx
	} else {
		fn = func(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
			return conn.Begin()
		}
	}

	return func(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
		var tx driver.Tx
		if err := c.instrum.withSpan(ctx, "db.Begin", "",
			func(ctx context.Context, span trace.Span) error {
				var err error
				tx, err = fn(ctx, opts)
				return err
			}); err != nil {
			return nil, err
		}
		return newTx(ctx, tx, c.instrum), nil
	}
}

//------------------------------------------------------------------------------

var _ driver.SessionResetter = (*otelConn)(nil)

func (c *otelConn) ResetSession(ctx context.Context) error {
	return c.resetSession(ctx)
}

type resetSessionFunc func(ctx context.Context) error

func (c *otelConn) createResetSessionFunc(conn driver.Conn) resetSessionFunc {
	if resetter, ok := c.Conn.(driver.SessionResetter); ok {
		return func(ctx context.Context) error {
			return resetter.ResetSession(ctx)
		}
	}
	return func(ctx context.Context) error {
		return driver.ErrSkip
	}
}

//------------------------------------------------------------------------------

var _ driver.NamedValueChecker = (*otelConn)(nil)

func (c *otelConn) CheckNamedValue(value *driver.NamedValue) error {
	return c.checkNamedValue(value)
}

type checkNamedValueFunc func(*driver.NamedValue) error

func (c *otelConn) createCheckNamedValueFunc(conn driver.Conn) checkNamedValueFunc {
	if checker, ok := c.Conn.(driver.NamedValueChecker); ok {
		return func(value *driver.NamedValue) error {
			return checker.CheckNamedValue(value)
		}
	}
	return func(value *driver.NamedValue) error {
		return driver.ErrSkip
	}
}

//------------------------------------------------------------------------------

func namedValueToValue(named []driver.NamedValue) ([]driver.Value, error) {
	args := make([]driver.Value, len(named))
	for n, param := range named {
		if len(param.Name) > 0 {
			return nil, errors.New("otelsql: driver does not support named parameters")
		}
		args[n] = param.Value
	}
	return args, nil
}

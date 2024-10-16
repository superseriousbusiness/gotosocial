// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package bundb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"codeberg.org/gruf/go-bytesize"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/metrics"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/tracing"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/migrate"
)

// DBService satisfies the DB interface
type DBService struct {
	db.Account
	db.Admin
	db.AdvancedMigration
	db.Application
	db.Basic
	db.Conversation
	db.Domain
	db.Emoji
	db.HeaderFilter
	db.Instance
	db.Interaction
	db.Filter
	db.List
	db.Marker
	db.Media
	db.Mention
	db.Move
	db.Notification
	db.Poll
	db.Relationship
	db.Report
	db.Rule
	db.Search
	db.Session
	db.SinBinStatus
	db.Status
	db.StatusBookmark
	db.StatusFave
	db.Tag
	db.Thread
	db.Timeline
	db.User
	db.Tombstone
	db.WorkerTask
	db *bun.DB
}

// GetDB returns the underlying database connection pool.
// Should only be used in testing + exceptional circumstance.
func (dbService *DBService) DB() *bun.DB {
	return dbService.db
}

func doMigration(ctx context.Context, db *bun.DB) error {
	migrator := migrate.NewMigrator(db, migrations.Migrations)

	if err := migrator.Init(ctx); err != nil {
		return err
	}

	group, err := migrator.Migrate(ctx)
	if err != nil && !strings.Contains(err.Error(), "no migrations") {
		return err
	}

	if group == nil || group.ID == 0 {
		log.Info(ctx, "there are no new migrations to run")
		return nil
	}

	log.Infof(ctx, "MIGRATED DATABASE TO %s", group)

	if db.Dialect().Name() == dialect.SQLite {
		log.Info(ctx,
			"running ANALYZE to update table and index statistics; this will take somewhere between "+
				"1-10 minutes, or maybe longer depending on your hardware and database size, please be patient",
		)
		_, err := db.ExecContext(ctx, "ANALYZE")
		if err != nil {
			log.Warnf(ctx, "ANALYZE failed, query planner may make poor life choices: %s", err)
		}
	}
	return nil
}

// NewBunDBService returns a bunDB derived from the provided config, which implements the go-fed DB interface.
// Under the hood, it uses https://github.com/uptrace/bun to create and maintain a database connection.
func NewBunDBService(ctx context.Context, state *state.State) (db.DB, error) {
	var db *bun.DB
	var err error
	t := strings.ToLower(config.GetDbType())

	switch t {
	case "postgres":
		db, err = pgConn(ctx)
		if err != nil {
			return nil, err
		}
	case "sqlite":
		db, err = sqliteConn(ctx)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("database type %s not supported for bundb", t)
	}

	// Add database query hooks.
	db.AddQueryHook(queryHook{})
	if config.GetTracingEnabled() {
		db.AddQueryHook(tracing.InstrumentBun())
	}
	if config.GetMetricsEnabled() {
		db.AddQueryHook(metrics.InstrumentBun())
	}

	// table registration is needed for many-to-many, see:
	// https://bun.uptrace.dev/orm/many-to-many-relation/
	for _, t := range []interface{}{
		&gtsmodel.AccountToEmoji{},
		&gtsmodel.ConversationToStatus{},
		&gtsmodel.StatusToEmoji{},
		&gtsmodel.StatusToTag{},
		&gtsmodel.ThreadToStatus{},
	} {
		db.RegisterModel(t)
	}

	// perform any pending database migrations: this includes
	// the very first 'migration' on startup which just creates
	// necessary tables
	if err := doMigration(ctx, db); err != nil {
		return nil, fmt.Errorf("db migration error: %s", err)
	}

	ps := &DBService{
		Account: &accountDB{
			db:    db,
			state: state,
		},
		Admin: &adminDB{
			db:    db,
			state: state,
		},
		AdvancedMigration: &advancedMigrationDB{
			db:    db,
			state: state,
		},
		Application: &applicationDB{
			db:    db,
			state: state,
		},
		Basic: &basicDB{
			db: db,
		},
		Conversation: &conversationDB{
			db:    db,
			state: state,
		},
		Domain: &domainDB{
			db:    db,
			state: state,
		},
		Emoji: &emojiDB{
			db:    db,
			state: state,
		},
		HeaderFilter: &headerFilterDB{
			db:    db,
			state: state,
		},
		Instance: &instanceDB{
			db:    db,
			state: state,
		},
		Interaction: &interactionDB{
			db:    db,
			state: state,
		},
		Filter: &filterDB{
			db:    db,
			state: state,
		},
		List: &listDB{
			db:    db,
			state: state,
		},
		Marker: &markerDB{
			db:    db,
			state: state,
		},
		Media: &mediaDB{
			db:    db,
			state: state,
		},
		Mention: &mentionDB{
			db:    db,
			state: state,
		},
		Move: &moveDB{
			db:    db,
			state: state,
		},
		Notification: &notificationDB{
			db:    db,
			state: state,
		},
		Poll: &pollDB{
			db:    db,
			state: state,
		},
		Relationship: &relationshipDB{
			db:    db,
			state: state,
		},
		Report: &reportDB{
			db:    db,
			state: state,
		},
		Rule: &ruleDB{
			db:    db,
			state: state,
		},
		Search: &searchDB{
			db:    db,
			state: state,
		},
		Session: &sessionDB{
			db: db,
		},
		SinBinStatus: &sinBinStatusDB{
			db:    db,
			state: state,
		},
		Status: &statusDB{
			db:    db,
			state: state,
		},
		StatusBookmark: &statusBookmarkDB{
			db:    db,
			state: state,
		},
		StatusFave: &statusFaveDB{
			db:    db,
			state: state,
		},
		Tag: &tagDB{
			db:    db,
			state: state,
		},
		Thread: &threadDB{
			db:    db,
			state: state,
		},
		Timeline: &timelineDB{
			db:    db,
			state: state,
		},
		User: &userDB{
			db:    db,
			state: state,
		},
		Tombstone: &tombstoneDB{
			db:    db,
			state: state,
		},
		WorkerTask: &workerTaskDB{
			db: db,
		},
		db: db,
	}

	// we can confidently return this useable service now
	return ps, nil
}

func pgConn(ctx context.Context) (*bun.DB, error) {
	opts, err := deriveBunDBPGOptions() //nolint:contextcheck
	if err != nil {
		return nil, fmt.Errorf("could not create bundb postgres options: %w", err)
	}

	cfg := stdlib.RegisterConnConfig(opts)

	sqldb, err := sql.Open("pgx-gts", cfg)
	if err != nil {
		return nil, fmt.Errorf("could not open postgres db: %w", err)
	}

	// Tune db connections for postgres, see:
	// - https://bun.uptrace.dev/guide/running-bun-in-production.html#database-sql
	// - https://www.alexedwards.net/blog/configuring-sqldb
	sqldb.SetMaxOpenConns(maxOpenConns())     // x number of conns per CPU
	sqldb.SetMaxIdleConns(2)                  // assume default 2; if max idle is less than max open, it will be automatically adjusted
	sqldb.SetConnMaxLifetime(5 * time.Minute) // fine to kill old connections

	db := bun.NewDB(sqldb, pgdialect.New())

	// ping to check the db is there and listening
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	log.Info(ctx, "connected to POSTGRES database")
	return db, nil
}

func sqliteConn(ctx context.Context) (*bun.DB, error) {
	// validate db address has actually been set
	address := config.GetDbAddress()
	if address == "" {
		return nil, fmt.Errorf("'%s' was not set when attempting to start sqlite", config.DbAddressFlag())
	}

	// Build SQLite connection address with prefs.
	address, inMem := buildSQLiteAddress(address)

	// Open new DB instance
	sqldb, err := sql.Open("sqlite-gts", address)
	if err != nil {
		return nil, fmt.Errorf("could not open sqlite db with address %s: %w", address, err)
	}

	// Tune db connections for sqlite, see:
	// - https://bun.uptrace.dev/guide/running-bun-in-production.html#database-sql
	// - https://www.alexedwards.net/blog/configuring-sqldb
	sqldb.SetMaxOpenConns(maxOpenConns()) // x number of conns per CPU
	sqldb.SetMaxIdleConns(1)              // only keep max 1 idle connection around
	if inMem {
		log.Warn(nil, "using sqlite in-memory mode; all data will be deleted when gts shuts down; this mode should only be used for debugging or running tests")
		// Don't close aged connections as this may wipe the DB.
		sqldb.SetConnMaxLifetime(0)
	} else {
		sqldb.SetConnMaxLifetime(5 * time.Minute)
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())

	// ping to check the db is there and listening
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("sqlite ping: %w", err)
	}

	log.Infof(ctx, "connected to SQLITE database with address %s", address)

	return db, nil
}

/*
	HANDY STUFF
*/

// maxOpenConns returns multiplier * GOMAXPROCS,
// returning just 1 instead if multiplier < 1.
func maxOpenConns() int {
	multiplier := config.GetDbMaxOpenConnsMultiplier()
	if multiplier < 1 {
		return 1
	}

	// Specifically for SQLite databases with
	// a journal mode of anything EXCEPT "wal",
	// only 1 concurrent connection is supported.
	if strings.ToLower(config.GetDbType()) == "sqlite" {
		journalMode := config.GetDbSqliteJournalMode()
		journalMode = strings.ToLower(journalMode)
		if journalMode != "wal" {
			return 1
		}
	}

	return multiplier * runtime.GOMAXPROCS(0)
}

// deriveBunDBPGOptions takes an application config and returns either a ready-to-use set of options
// with sensible defaults, or an error if it's not satisfied by the provided config.
func deriveBunDBPGOptions() (*pgx.ConnConfig, error) {
	url := config.GetDbPostgresConnectionString()

	// if database URL is defined, ignore other DB related configuration fields
	if url != "" {
		cfg, err := pgx.ParseConfig(url)
		return cfg, err
	}
	// these are all optional, the db adapter figures out defaults
	address := config.GetDbAddress()

	// validate database
	database := config.GetDbDatabase()
	if database == "" {
		return nil, errors.New("no database set")
	}

	var tlsConfig *tls.Config
	switch config.GetDbTLSMode() {
	case "", "disable":
		break // nothing to do
	case "enable":
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		}
	case "require":
		tlsConfig = &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         address,
			MinVersion:         tls.VersionTLS12,
		}
	}

	if certPath := config.GetDbTLSCACert(); tlsConfig != nil && certPath != "" {
		// load the system cert pool first -- we'll append the given CA cert to this
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("error fetching system CA cert pool: %s", err)
		}

		// open the file itself and make sure there's something in it
		caCertBytes, err := os.ReadFile(certPath)
		if err != nil {
			return nil, fmt.Errorf("error opening CA certificate at %s: %s", certPath, err)
		}
		if len(caCertBytes) == 0 {
			return nil, fmt.Errorf("ca cert at %s was empty", certPath)
		}

		// make sure we have a PEM block
		caPem, _ := pem.Decode(caCertBytes)
		if caPem == nil {
			return nil, fmt.Errorf("could not parse cert at %s into PEM", certPath)
		}

		// parse the PEM block into the certificate
		caCert, err := x509.ParseCertificate(caPem.Bytes)
		if err != nil {
			return nil, fmt.Errorf("could not parse cert at %s into x509 certificate: %w", certPath, err)
		}

		// we're happy, add it to the existing pool and then use this pool in our tls config
		certPool.AddCert(caCert)
		tlsConfig.RootCAs = certPool
	}

	cfg, _ := pgx.ParseConfig("")
	if address != "" {
		cfg.Host = address
	}
	if port := config.GetDbPort(); port > 0 {
		if port > math.MaxUint16 {
			return nil, errors.New("invalid port, must be in range 1-65535")
		}
		cfg.Port = uint16(port) // #nosec G115 -- Just validated above.
	}
	if u := config.GetDbUser(); u != "" {
		cfg.User = u
	}
	if p := config.GetDbPassword(); p != "" {
		cfg.Password = p
	}
	if tlsConfig != nil {
		cfg.TLSConfig = tlsConfig
	}
	cfg.Database = database
	cfg.RuntimeParams["application_name"] = config.GetApplicationName()

	return cfg, nil
}

// buildSQLiteAddress will build an SQLite address string from given config input,
// appending user defined SQLite connection preferences (e.g. cache_size, journal_mode etc).
// The returned bool indicates whether this is an in-memory address or not.
func buildSQLiteAddress(addr string) (string, bool) {
	// Notes on SQLite preferences:
	//
	// - SQLite by itself supports setting a subset of its configuration options
	//   via URI query arguments in the connection. Namely `mode` and `cache`.
	//   This is the same situation for the directly transpiled C->Go code in
	//   modernc.org/sqlite, i.e. modernc.org/sqlite/lib, NOT the Go SQL driver.
	//
	// - `modernc.org/sqlite` has a "shim" around it to allow the directly
	//   transpiled C code to be usable with a more native Go API. This is in
	//   the form of a `database/sql/driver.Driver{}` implementation that calls
	//   through to the transpiled C code.
	//
	// - The SQLite shim we interface with adds support for setting ANY of the
	//   configuration options via query arguments, through using a special `_pragma`
	//   query key that specifies SQLite PRAGMAs to set upon opening each connection.
	//   As such you will see below that most config is set with the `_pragma` key.
	//
	// - As for why we're setting these PRAGMAs by connection string instead of
	//   directly executing the PRAGMAs ourselves? That's to ensure that all of
	//   configuration options are set across _all_ of our SQLite connections, given
	//   that we are a multi-threaded (not directly in a C way) application and that
	//   each connection is a separate SQLite instance opening the same database.
	//   And the `database/sql` package provides transparent connection pooling.
	//   Some data is shared between connections, for example the `journal_mode`
	//   as that is set in a bit of the file header, but to be sure with the other
	//   settings we just add them all to the connection URI string.
	//
	// - We specifically set the `busy_timeout` PRAGMA before the `journal_mode`.
	//   When Write-Ahead-Logging (WAL) is enabled, in order to handle the issues
	//   that may arise between separate concurrent read/write threads racing for
	//   the same database file (and write-ahead log), SQLite will sometimes return
	//   an `SQLITE_BUSY` error code, which indicates that the query was aborted
	//   due to a data race and must be retried. The `busy_timeout` PRAGMA configures
	//   a function handler that SQLite can use internally to handle these data races,
	//   in that it will attempt to retry the query until the `busy_timeout` time is
	//   reached. And for whatever reason (:shrug:) SQLite is very particular about
	//   setting this BEFORE the `journal_mode` is set, otherwise you can end up
	//   running into more of these `SQLITE_BUSY` return codes than you might expect.
	//
	// - One final thing (I promise!): `SQLITE_BUSY` is only handled by the internal
	//   `busy_timeout` handler in the case that a data race occurs contending for
	//  table locks. THERE ARE STILL OTHER SITUATIONS IN WHICH THIS MAY BE RETURNED!
	//  As such, we use our wrapping DB{} and Tx{} types (in "db.go") which make use
	//  of our own retry-busy handler.

	// Drop anything fancy from DB address
	addr = strings.Split(addr, "?")[0]       // drop any provided query strings
	addr = strings.TrimPrefix(addr, "file:") // we'll prepend this later ourselves

	// build our own SQLite preferences
	// as a series of URL encoded values
	prefs := make(url.Values)

	// use immediate transaction lock mode to fail quickly if tx can't lock
	// see https://pkg.go.dev/modernc.org/sqlite#Driver.Open
	prefs.Add("_txlock", "immediate")

	inMem := false
	if addr == ":memory:" {
		// Use random name for in-memory instead of ':memory:', so
		// multiple in-mem databases can be created without conflict.
		inMem = true
		addr = "/" + uuid.NewString()
		prefs.Add("vfs", "memdb")
	}

	if dur := config.GetDbSqliteBusyTimeout(); dur > 0 {
		// Set the user provided SQLite busy timeout
		// NOTE: MUST BE SET BEFORE THE JOURNAL MODE.
		prefs.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", dur.Milliseconds()))
	}

	if mode := config.GetDbSqliteJournalMode(); mode != "" {
		// Set the user provided SQLite journal mode.
		prefs.Add("_pragma", fmt.Sprintf("journal_mode(%s)", mode))
	}

	if mode := config.GetDbSqliteSynchronous(); mode != "" {
		// Set the user provided SQLite synchronous mode.
		prefs.Add("_pragma", fmt.Sprintf("synchronous(%s)", mode))
	}

	if sz := config.GetDbSqliteCacheSize(); sz > 0 {
		// Set the user provided SQLite cache size (in kibibytes)
		// Prepend a '-' character to this to indicate to sqlite
		// that we're giving kibibytes rather than num pages.
		// https://www.sqlite.org/pragma.html#pragma_cache_size
		prefs.Add("_pragma", fmt.Sprintf("cache_size(-%d)", uint64(sz/bytesize.KiB)))
	}

	var b strings.Builder
	b.WriteString("file:")
	b.WriteString(addr)
	b.WriteString("?")
	b.WriteString(prefs.Encode())
	return b.String(), inMem
}

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
	"os"
	"runtime"
	"strconv"
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
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/tracing"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/migrate"

	"modernc.org/sqlite"
)

var registerTables = []interface{}{
	&gtsmodel.AccountToEmoji{},
	&gtsmodel.StatusToEmoji{},
	&gtsmodel.StatusToTag{},
}

// DBService satisfies the DB interface
type DBService struct {
	db.Account
	db.Admin
	db.Application
	db.Basic
	db.Domain
	db.Emoji
	db.Instance
	db.List
	db.Marker
	db.Media
	db.Mention
	db.Notification
	db.Relationship
	db.Report
	db.Search
	db.Session
	db.Status
	db.StatusBookmark
	db.StatusFave
	db.Tag
	db.Timeline
	db.User
	db.Tombstone
	db *DB
}

// GetDB returns the underlying database connection pool.
// Should only be used in testing + exceptional circumstance.
func (dbService *DBService) DB() *DB {
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
	return nil
}

// NewBunDBService returns a bunDB derived from the provided config, which implements the go-fed DB interface.
// Under the hood, it uses https://github.com/uptrace/bun to create and maintain a database connection.
func NewBunDBService(ctx context.Context, state *state.State) (db.DB, error) {
	var db *DB
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

	// execute sqlite pragmas *after* adding database hook;
	// this allows the pragma queries to be logged
	if t == "sqlite" {
		if err := sqlitePragmas(ctx, db); err != nil {
			return nil, err
		}
	}

	// table registration is needed for many-to-many, see:
	// https://bun.uptrace.dev/orm/many-to-many-relation/
	for _, t := range registerTables {
		db.RegisterModel(t)
	}

	// perform any pending database migrations: this includes
	// the very first 'migration' on startup which just creates
	// necessary tables
	if err := doMigration(ctx, db.bun); err != nil {
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
		Application: &applicationDB{
			db:    db,
			state: state,
		},
		Basic: &basicDB{
			db: db,
		},
		Domain: &domainDB{
			db:    db,
			state: state,
		},
		Emoji: &emojiDB{
			db:    db,
			state: state,
		},
		Instance: &instanceDB{
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
		Notification: &notificationDB{
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
		Search: &searchDB{
			db:    db,
			state: state,
		},
		Session: &sessionDB{
			db: db,
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
			conn:  db,
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
		db: db,
	}

	// we can confidently return this useable service now
	return ps, nil
}

func pgConn(ctx context.Context) (*DB, error) {
	opts, err := deriveBunDBPGOptions() //nolint:contextcheck
	if err != nil {
		return nil, fmt.Errorf("could not create bundb postgres options: %s", err)
	}

	sqldb := stdlib.OpenDB(*opts)

	// Tune db connections for postgres, see:
	// - https://bun.uptrace.dev/guide/running-bun-in-production.html#database-sql
	// - https://www.alexedwards.net/blog/configuring-sqldb
	sqldb.SetMaxOpenConns(maxOpenConns())     // x number of conns per CPU
	sqldb.SetMaxIdleConns(2)                  // assume default 2; if max idle is less than max open, it will be automatically adjusted
	sqldb.SetConnMaxLifetime(5 * time.Minute) // fine to kill old connections

	db := WrapDB(bun.NewDB(sqldb, pgdialect.New()))

	// ping to check the db is there and listening
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping: %s", err)
	}

	log.Info(ctx, "connected to POSTGRES database")
	return db, nil
}

func sqliteConn(ctx context.Context) (*DB, error) {
	// validate db address has actually been set
	address := config.GetDbAddress()
	if address == "" {
		return nil, fmt.Errorf("'%s' was not set when attempting to start sqlite", config.DbAddressFlag())
	}

	// Drop anything fancy from DB address
	address = strings.Split(address, "?")[0]       // drop any provided query strings
	address = strings.TrimPrefix(address, "file:") // we'll prepend this later ourselves

	// build our own SQLite preferences
	prefs := []string{
		// use immediate transaction lock mode to fail quickly if tx can't lock
		// see https://pkg.go.dev/modernc.org/sqlite#Driver.Open
		"_txlock=immediate",
	}

	if address == ":memory:" {
		log.Warn(ctx, "using sqlite in-memory mode; all data will be deleted when gts shuts down; this mode should only be used for debugging or running tests")

		// Use random name for in-memory instead of ':memory:', so
		// multiple in-mem databases can be created without conflict.
		address = uuid.NewString()

		// in-mem-specific preferences
		prefs = append(prefs, []string{
			"mode=memory",  // indicate in-memory mode using query
			"cache=shared", // shared cache so that tests don't fail
		}...)
	}

	// rebuild address string with our derived preferences
	address = "file:" + address
	for i, q := range prefs {
		var prefix string
		if i == 0 {
			prefix = "?"
		} else {
			prefix = "&"
		}
		address += prefix + q
	}

	// Open new DB instance
	sqldb, err := sql.Open("sqlite", address)
	if err != nil {
		if errWithCode, ok := err.(*sqlite.Error); ok {
			err = errors.New(sqlite.ErrorCodeString[errWithCode.Code()])
		}
		return nil, fmt.Errorf("could not open sqlite db with address %s: %w", address, err)
	}

	// Tune db connections for sqlite, see:
	// - https://bun.uptrace.dev/guide/running-bun-in-production.html#database-sql
	// - https://www.alexedwards.net/blog/configuring-sqldb
	sqldb.SetMaxOpenConns(maxOpenConns()) // x number of conns per CPU
	sqldb.SetMaxIdleConns(1)              // only keep max 1 idle connection around
	sqldb.SetConnMaxLifetime(0)           // don't kill connections due to age

	// Wrap Bun database conn in our own wrapper
	db := WrapDB(bun.NewDB(sqldb, sqlitedialect.New()))

	// ping to check the db is there and listening
	if err := db.PingContext(ctx); err != nil {
		if errWithCode, ok := err.(*sqlite.Error); ok {
			err = errors.New(sqlite.ErrorCodeString[errWithCode.Code()])
		}
		return nil, fmt.Errorf("sqlite ping: %s", err)
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
	return multiplier * runtime.GOMAXPROCS(0)
}

// deriveBunDBPGOptions takes an application config and returns either a ready-to-use set of options
// with sensible defaults, or an error if it's not satisfied by the provided config.
func deriveBunDBPGOptions() (*pgx.ConnConfig, error) {
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
		/* #nosec G402 */
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
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
			return nil, fmt.Errorf("could not parse cert at %s into x509 certificate: %s", certPath, err)
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
		cfg.Port = uint16(port)
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

// sqlitePragmas sets desired sqlite pragmas based on configured values, and
// logs the results of the pragma queries. Errors if something goes wrong.
func sqlitePragmas(ctx context.Context, db *DB) error {
	var pragmas [][]string
	if mode := config.GetDbSqliteJournalMode(); mode != "" {
		// Set the user provided SQLite journal mode
		pragmas = append(pragmas, []string{"journal_mode", mode})
	}

	if mode := config.GetDbSqliteSynchronous(); mode != "" {
		// Set the user provided SQLite synchronous mode
		pragmas = append(pragmas, []string{"synchronous", mode})
	}

	if size := config.GetDbSqliteCacheSize(); size > 0 {
		// Set the user provided SQLite cache size (in kibibytes)
		// Prepend a '-' character to this to indicate to sqlite
		// that we're giving kibibytes rather than num pages.
		// https://www.sqlite.org/pragma.html#pragma_cache_size
		s := "-" + strconv.FormatUint(uint64(size/bytesize.KiB), 10)
		pragmas = append(pragmas, []string{"cache_size", s})
	}

	if timeout := config.GetDbSqliteBusyTimeout(); timeout > 0 {
		t := strconv.FormatInt(timeout.Milliseconds(), 10)
		pragmas = append(pragmas, []string{"busy_timeout", t})
	}

	for _, p := range pragmas {
		pk := p[0]
		pv := p[1]

		if _, err := db.ExecContext(ctx, "PRAGMA ?=?", bun.Ident(pk), bun.Safe(pv)); err != nil {
			return fmt.Errorf("error executing sqlite pragma %s: %w", pk, err)
		}

		var res string
		if err := db.NewRaw("PRAGMA ?", bun.Ident(pk)).Scan(ctx, &res); err != nil {
			return fmt.Errorf("error scanning sqlite pragma %s: %w", pv, err)
		}

		log.Infof(ctx, "sqlite pragma %s set to %s", pk, res)
	}

	return nil
}

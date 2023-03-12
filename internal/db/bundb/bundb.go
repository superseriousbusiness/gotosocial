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
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
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
	db.Basic
	db.Domain
	db.Emoji
	db.Instance
	db.Media
	db.Mention
	db.Notification
	db.Relationship
	db.Report
	db.Session
	db.Status
	db.StatusBookmark
	db.StatusFave
	db.Timeline
	db.User
	db.Tombstone
	conn *DBConn
}

// GetConn returns the underlying bun connection.
// Should only be used in testing + exceptional circumstance.
func (dbService *DBService) GetConn() *DBConn {
	return dbService.conn
}

func doMigration(ctx context.Context, db *bun.DB) error {
	migrator := migrate.NewMigrator(db, migrations.Migrations)

	if err := migrator.Init(ctx); err != nil {
		return err
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		if err.Error() == "migrate: there are no any migrations" {
			return nil
		}
		return err
	}

	if group.ID == 0 {
		log.Info(ctx, "there are no new migrations to run")
		return nil
	}

	log.Infof(ctx, "MIGRATED DATABASE TO %s", group)
	return nil
}

// NewBunDBService returns a bunDB derived from the provided config, which implements the go-fed DB interface.
// Under the hood, it uses https://github.com/uptrace/bun to create and maintain a database connection.
func NewBunDBService(ctx context.Context, state *state.State) (db.DB, error) {
	var conn *DBConn
	var err error
	t := strings.ToLower(config.GetDbType())

	switch t {
	case "postgres":
		conn, err = pgConn(ctx)
		if err != nil {
			return nil, err
		}
	case "sqlite":
		conn, err = sqliteConn(ctx)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("database type %s not supported for bundb", t)
	}

	// Add database query hook
	conn.DB.AddQueryHook(queryHook{})

	// execute sqlite pragmas *after* adding database hook;
	// this allows the pragma queries to be logged
	if t == "sqlite" {
		if err := sqlitePragmas(ctx, conn); err != nil {
			return nil, err
		}
	}

	// table registration is needed for many-to-many, see:
	// https://bun.uptrace.dev/orm/many-to-many-relation/
	for _, t := range registerTables {
		conn.RegisterModel(t)
	}

	// perform any pending database migrations: this includes
	// the very first 'migration' on startup which just creates
	// necessary tables
	if err := doMigration(ctx, conn.DB); err != nil {
		return nil, fmt.Errorf("db migration error: %s", err)
	}

	ps := &DBService{
		Account: &accountDB{
			conn:  conn,
			state: state,
		},
		Admin: &adminDB{
			conn:  conn,
			state: state,
		},
		Basic: &basicDB{
			conn: conn,
		},
		Domain: &domainDB{
			conn:  conn,
			state: state,
		},
		Emoji: &emojiDB{
			conn:  conn,
			state: state,
		},
		Instance: &instanceDB{
			conn: conn,
		},
		Media: &mediaDB{
			conn:  conn,
			state: state,
		},
		Mention: &mentionDB{
			conn:  conn,
			state: state,
		},
		Notification: &notificationDB{
			conn:  conn,
			state: state,
		},
		Relationship: &relationshipDB{
			conn:  conn,
			state: state,
		},
		Report: &reportDB{
			conn:  conn,
			state: state,
		},
		Session: &sessionDB{
			conn: conn,
		},
		Status: &statusDB{
			conn:  conn,
			state: state,
		},
		StatusBookmark: &statusBookmarkDB{
			conn:  conn,
			state: state,
		},
		StatusFave: &statusFaveDB{
			conn:  conn,
			state: state,
		},
		Timeline: &timelineDB{
			conn:  conn,
			state: state,
		},
		User: &userDB{
			conn:  conn,
			state: state,
		},
		Tombstone: &tombstoneDB{
			conn:  conn,
			state: state,
		},
		conn: conn,
	}

	// we can confidently return this useable service now
	return ps, nil
}

func pgConn(ctx context.Context) (*DBConn, error) {
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

	conn := WrapDBConn(bun.NewDB(sqldb, pgdialect.New()))

	// ping to check the db is there and listening
	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping: %s", err)
	}

	log.Info(ctx, "connected to POSTGRES database")
	return conn, nil
}

func sqliteConn(ctx context.Context) (*DBConn, error) {
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
	sqldb.SetMaxOpenConns(1)    // only 1 connection regardless of multiplier, see https://github.com/superseriousbusiness/gotosocial/issues/1407
	sqldb.SetMaxIdleConns(1)    // only keep max 1 idle connection around
	sqldb.SetConnMaxLifetime(0) // don't kill connections due to age

	// Wrap Bun database conn in our own wrapper
	conn := WrapDBConn(bun.NewDB(sqldb, sqlitedialect.New()))

	// ping to check the db is there and listening
	if err := conn.PingContext(ctx); err != nil {
		if errWithCode, ok := err.(*sqlite.Error); ok {
			err = errors.New(sqlite.ErrorCodeString[errWithCode.Code()])
		}
		return nil, fmt.Errorf("sqlite ping: %s", err)
	}
	log.Infof(ctx, "connected to SQLITE database with address %s", address)

	return conn, nil
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
	cfg.PreferSimpleProtocol = true
	cfg.RuntimeParams["application_name"] = config.GetApplicationName()

	return cfg, nil
}

// sqlitePragmas sets desired sqlite pragmas based on configured values, and
// logs the results of the pragma queries. Errors if something goes wrong.
func sqlitePragmas(ctx context.Context, conn *DBConn) error {
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

		if _, err := conn.DB.ExecContext(ctx, "PRAGMA ?=?", bun.Ident(pk), bun.Safe(pv)); err != nil {
			return fmt.Errorf("error executing sqlite pragma %s: %w", pk, err)
		}

		var res string
		if err := conn.DB.NewRaw("PRAGMA ?", bun.Ident(pk)).Scan(ctx, &res); err != nil {
			return fmt.Errorf("error scanning sqlite pragma %s: %w", pv, err)
		}

		log.Infof(ctx, "sqlite pragma %s set to %s", pk, res)
	}

	return nil
}

/*
	CONVERSION FUNCTIONS
*/

func (dbService *DBService) TagStringToTag(ctx context.Context, t string, originAccountID string) (*gtsmodel.Tag, error) {
	protocol := config.GetProtocol()
	host := config.GetHost()
	now := time.Now()

	tag := &gtsmodel.Tag{}
	// we can use selectorinsert here to create the new tag if it doesn't exist already
	// inserted will be true if this is a new tag we just created
	if err := dbService.conn.NewSelect().Model(tag).Where("LOWER(?) = LOWER(?)", bun.Ident("name"), t).Scan(ctx); err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting tag with name %s: %s", t, err)
	}

	if tag.ID == "" {
		// tag doesn't exist yet so populate it
		newID, err := id.NewRandomULID()
		if err != nil {
			return nil, err
		}
		tag.ID = newID
		tag.URL = protocol + "://" + host + "/tags/" + t
		tag.Name = t
		tag.FirstSeenFromAccountID = originAccountID
		tag.CreatedAt = now
		tag.UpdatedAt = now
		useable := true
		tag.Useable = &useable
		listable := true
		tag.Listable = &listable
	}

	// bail already if the tag isn't useable
	if !*tag.Useable {
		return nil, fmt.Errorf("tag %s is not useable", t)
	}
	tag.LastStatusAt = now
	return tag, nil
}

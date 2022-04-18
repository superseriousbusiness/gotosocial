/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

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
	"strings"
	"time"

	"github.com/ReneKroon/ttlcache"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/migrate"

	"modernc.org/sqlite"
)

const (
	dbTypePostgres = "postgres"
	dbTypeSqlite   = "sqlite"

	// dbTLSModeDisable does not attempt to make a TLS connection to the database.
	dbTLSModeDisable = "disable"
	// dbTLSModeEnable attempts to make a TLS connection to the database, but doesn't fail if
	// the certificate passed by the database isn't verified.
	dbTLSModeEnable = "enable"
	// dbTLSModeRequire attempts to make a TLS connection to the database, and requires
	// that the certificate presented by the database is valid.
	dbTLSModeRequire = "require"
	// dbTLSModeUnset means that the TLS mode has not been set.
	dbTLSModeUnset = ""
)

var registerTables = []interface{}{
	&gtsmodel.StatusToEmoji{},
	&gtsmodel.StatusToTag{},
}

// bunDBService satisfies the DB interface
type bunDBService struct {
	db.Account
	db.Admin
	db.Basic
	db.Domain
	db.Instance
	db.Media
	db.Mention
	db.Notification
	db.Relationship
	db.Session
	db.Status
	db.Timeline
	conn *DBConn
}

func doMigration(ctx context.Context, db *bun.DB) error {
	l := logrus.WithField("func", "doMigration")

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
		l.Info("there are no new migrations to run")
		return nil
	}

	l.Infof("MIGRATED DATABASE TO %s", group)
	return nil
}

// NewBunDBService returns a bunDB derived from the provided config, which implements the go-fed DB interface.
// Under the hood, it uses https://github.com/uptrace/bun to create and maintain a database connection.
func NewBunDBService(ctx context.Context) (db.DB, error) {
	var conn *DBConn
	var err error
	dbType := strings.ToLower(viper.GetString(config.Keys.DbType))

	switch dbType {
	case dbTypePostgres:
		conn, err = pgConn(ctx)
		if err != nil {
			return nil, err
		}
	case dbTypeSqlite:
		conn, err = sqliteConn(ctx)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("database type %s not supported for bundb", dbType)
	}

	// add a hook to log queries and the time they take
	// only do this for logging where performance isn't 1st concern
	if logrus.GetLevel() >= logrus.DebugLevel && viper.GetBool(config.Keys.LogDbQueries) {
		conn.DB.AddQueryHook(newDebugQueryHook())
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

	accounts := &accountDB{conn: conn, cache: cache.NewAccountCache()}

	ps := &bunDBService{
		Account: accounts,
		Admin: &adminDB{
			conn: conn,
		},
		Basic: &basicDB{
			conn: conn,
		},
		Domain: &domainDB{
			conn: conn,
		},
		Instance: &instanceDB{
			conn: conn,
		},
		Media: &mediaDB{
			conn: conn,
		},
		Mention: &mentionDB{
			conn:  conn,
			cache: ttlcache.NewCache(),
		},
		Notification: &notificationDB{
			conn:  conn,
			cache: ttlcache.NewCache(),
		},
		Relationship: &relationshipDB{
			conn: conn,
		},
		Session: &sessionDB{
			conn: conn,
		},
		Status: &statusDB{
			conn:     conn,
			cache:    cache.NewStatusCache(),
			accounts: accounts,
		},
		Timeline: &timelineDB{
			conn: conn,
		},
		conn: conn,
	}

	// we can confidently return this useable service now
	return ps, nil
}

func sqliteConn(ctx context.Context) (*DBConn, error) {
	// validate db address has actually been set
	dbAddress := viper.GetString(config.Keys.DbAddress)
	if dbAddress == "" {
		return nil, fmt.Errorf("'%s' was not set when attempting to start sqlite", config.Keys.DbAddress)
	}

	// Drop anything fancy from DB address
	dbAddress = strings.Split(dbAddress, "?")[0]
	dbAddress = strings.TrimPrefix(dbAddress, "file:")

	// Append our own SQLite preferences
	dbAddress = "file:" + dbAddress + "?cache=shared"

	// Open new DB instance
	sqldb, err := sql.Open("sqlite", dbAddress)
	if err != nil {
		if errWithCode, ok := err.(*sqlite.Error); ok {
			err = errors.New(sqlite.ErrorCodeString[errWithCode.Code()])
		}
		return nil, fmt.Errorf("could not open sqlite db: %s", err)
	}

	tweakConnectionValues(sqldb)

	if dbAddress == "file::memory:?cache=shared" {
		logrus.Warn("sqlite in-memory database should only be used for debugging")
		// don't close connections on disconnect -- otherwise
		// the SQLite database will be deleted when there
		// are no active connections
		sqldb.SetConnMaxLifetime(0)
	}

	conn := WrapDBConn(bun.NewDB(sqldb, sqlitedialect.New()))

	// ping to check the db is there and listening
	if err := conn.PingContext(ctx); err != nil {
		if errWithCode, ok := err.(*sqlite.Error); ok {
			err = errors.New(sqlite.ErrorCodeString[errWithCode.Code()])
		}
		return nil, fmt.Errorf("sqlite ping: %s", err)
	}

	logrus.Info("connected to SQLITE database")
	return conn, nil
}

func pgConn(ctx context.Context) (*DBConn, error) {
	opts, err := deriveBunDBPGOptions()
	if err != nil {
		return nil, fmt.Errorf("could not create bundb postgres options: %s", err)
	}

	sqldb := stdlib.OpenDB(*opts)

	tweakConnectionValues(sqldb)

	conn := WrapDBConn(bun.NewDB(sqldb, pgdialect.New()))

	// ping to check the db is there and listening
	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping: %s", err)
	}

	logrus.Info("connected to POSTGRES database")
	return conn, nil
}

/*
	HANDY STUFF
*/

// deriveBunDBPGOptions takes an application config and returns either a ready-to-use set of options
// with sensible defaults, or an error if it's not satisfied by the provided config.
func deriveBunDBPGOptions() (*pgx.ConnConfig, error) {
	keys := config.Keys

	if strings.ToUpper(viper.GetString(keys.DbType)) != db.DBTypePostgres {
		return nil, fmt.Errorf("expected db type of %s but got %s", db.DBTypePostgres, viper.GetString(keys.DbType))
	}

	// these are all optional, the db adapter figures out defaults
	port := viper.GetInt(keys.DbPort)
	address := viper.GetString(keys.DbAddress)
	username := viper.GetString(keys.DbUser)
	password := viper.GetString(keys.DbPassword)

	// validate database
	database := viper.GetString(keys.DbDatabase)
	if database == "" {
		return nil, errors.New("no database set")
	}

	var tlsConfig *tls.Config
	tlsMode := viper.GetString(keys.DbTLSMode)
	switch tlsMode {
	case dbTLSModeDisable, dbTLSModeUnset:
		break // nothing to do
	case dbTLSModeEnable:
		/* #nosec G402 */
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	case dbTLSModeRequire:
		tlsConfig = &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         viper.GetString(keys.DbAddress),
			MinVersion:         tls.VersionTLS12,
		}
	}

	caCertPath := viper.GetString(keys.DbTLSCACert)
	if tlsConfig != nil && caCertPath != "" {
		// load the system cert pool first -- we'll append the given CA cert to this
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("error fetching system CA cert pool: %s", err)
		}

		// open the file itself and make sure there's something in it
		caCertBytes, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, fmt.Errorf("error opening CA certificate at %s: %s", caCertPath, err)
		}
		if len(caCertBytes) == 0 {
			return nil, fmt.Errorf("ca cert at %s was empty", caCertPath)
		}

		// make sure we have a PEM block
		caPem, _ := pem.Decode(caCertBytes)
		if caPem == nil {
			return nil, fmt.Errorf("could not parse cert at %s into PEM", caCertPath)
		}

		// parse the PEM block into the certificate
		caCert, err := x509.ParseCertificate(caPem.Bytes)
		if err != nil {
			return nil, fmt.Errorf("could not parse cert at %s into x509 certificate: %s", caCertPath, err)
		}

		// we're happy, add it to the existing pool and then use this pool in our tls config
		certPool.AddCert(caCert)
		tlsConfig.RootCAs = certPool
	}

	cfg, _ := pgx.ParseConfig("")
	if address != "" {
		cfg.Host = address
	}
	if port > 0 {
		cfg.Port = uint16(port)
	}
	if username != "" {
		cfg.User = username
	}
	if password != "" {
		cfg.Password = password
	}
	if tlsConfig != nil {
		cfg.TLSConfig = tlsConfig
	}
	cfg.Database = database
	cfg.PreferSimpleProtocol = true
	cfg.RuntimeParams["application_name"] = viper.GetString(keys.ApplicationName)

	return cfg, nil
}

// https://bun.uptrace.dev/postgres/running-bun-in-production.html#database-sql
func tweakConnectionValues(sqldb *sql.DB) {
	maxOpenConns := 4 * runtime.GOMAXPROCS(0)
	sqldb.SetMaxOpenConns(maxOpenConns)
	sqldb.SetMaxIdleConns(maxOpenConns)
}

/*
	CONVERSION FUNCTIONS
*/

func (ps *bunDBService) TagStringsToTags(ctx context.Context, tags []string, originAccountID string) ([]*gtsmodel.Tag, error) {
	protocol := viper.GetString(config.Keys.Protocol)
	host := viper.GetString(config.Keys.Host)

	newTags := []*gtsmodel.Tag{}
	for _, t := range tags {
		tag := &gtsmodel.Tag{}
		// we can use selectorinsert here to create the new tag if it doesn't exist already
		// inserted will be true if this is a new tag we just created
		if err := ps.conn.NewSelect().Model(tag).Where("LOWER(?) = LOWER(?)", bun.Ident("name"), t).Scan(ctx); err != nil {
			if err == sql.ErrNoRows {
				// tag doesn't exist yet so populate it
				newID, err := id.NewRandomULID()
				if err != nil {
					return nil, err
				}
				tag.ID = newID
				tag.URL = fmt.Sprintf("%s://%s/tags/%s", protocol, host, t)
				tag.Name = t
				tag.FirstSeenFromAccountID = originAccountID
				tag.CreatedAt = time.Now()
				tag.UpdatedAt = time.Now()
				tag.Useable = true
				tag.Listable = true
			} else {
				return nil, fmt.Errorf("error getting tag with name %s: %s", t, err)
			}
		}

		// bail already if the tag isn't useable
		if !tag.Useable {
			continue
		}
		tag.LastStatusAt = time.Now()
		newTags = append(newTags, tag)
	}
	return newTags, nil
}

func (ps *bunDBService) EmojiStringsToEmojis(ctx context.Context, emojis []string) ([]*gtsmodel.Emoji, error) {
	newEmojis := []*gtsmodel.Emoji{}
	for _, e := range emojis {
		emoji := &gtsmodel.Emoji{}
		err := ps.conn.NewSelect().Model(emoji).Where("shortcode = ?", e).Where("visible_in_picker = true").Where("disabled = false").Scan(ctx)
		if err != nil {
			if err == sql.ErrNoRows {
				// no result found for this username/domain so just don't include it as an emoji and carry on about our business
				logrus.Debugf("no emoji found with shortcode %s, skipping it", e)
				continue
			}
			// a serious error has happened so bail
			return nil, fmt.Errorf("error getting emoji with shortcode %s: %s", e, err)
		}
		newEmojis = append(newEmojis, emoji)
	}
	return newEmojis, nil
}

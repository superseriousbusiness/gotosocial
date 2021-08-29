/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"strings"
	"time"

	"github.com/ReneKroon/ttlcache"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "modernc.org/sqlite"
)

const (
	dbTypePostgres = "postgres"
	dbTypeSqlite   = "sqlite"
)

var registerTables []interface{} = []interface{}{
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
	config *config.Config
	conn   *DBConn
}

// NewBunDBService returns a bunDB derived from the provided config, which implements the go-fed DB interface.
// Under the hood, it uses https://github.com/uptrace/bun to create and maintain a database connection.
func NewBunDBService(ctx context.Context, c *config.Config, log *logrus.Logger) (db.DB, error) {
	var sqldb *sql.DB
	var conn *DBConn

	// depending on the database type we're trying to create, we need to use a different driver...
	switch strings.ToLower(c.DBConfig.Type) {
	case dbTypePostgres:
		// POSTGRES
		opts, err := deriveBunDBPGOptions(c)
		if err != nil {
			return nil, fmt.Errorf("could not create bundb postgres options: %s", err)
		}
		sqldb = stdlib.OpenDB(*opts)
		conn = WrapDBConn(bun.NewDB(sqldb, pgdialect.New()), log)
	case dbTypeSqlite:
		// SQLITE
		var err error
		sqldb, err = sql.Open("sqlite", c.DBConfig.Address)
		if err != nil {
			return nil, fmt.Errorf("could not open sqlite db: %s", err)
		}
		conn = WrapDBConn(bun.NewDB(sqldb, sqlitedialect.New()), log)

		if strings.HasPrefix(strings.TrimPrefix(c.DBConfig.Address, "file:"), ":memory:") {
			log.Warn("sqlite in-memory database should only be used for debugging")

			// don't close connections on disconnect -- otherwise
			// the SQLite database will be deleted when there
			// are no active connections
			sqldb.SetConnMaxLifetime(0)
		}
	default:
		return nil, fmt.Errorf("database type %s not supported for bundb", strings.ToLower(c.DBConfig.Type))
	}

	// actually *begin* the connection so that we can tell if the db is there and listening
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("db connection error: %s", err)
	}
	log.Info("connected to database")

	for _, t := range registerTables {
		// https://bun.uptrace.dev/orm/many-to-many-relation/
		conn.RegisterModel(t)
	}

	ps := &bunDBService{
		Account: &accountDB{
			config: c,
			conn:   conn,
		},
		Admin: &adminDB{
			config: c,
			conn:   conn,
		},
		Basic: &basicDB{
			config: c,
			conn:   conn,
		},
		Domain: &domainDB{
			config: c,
			conn:   conn,
		},
		Instance: &instanceDB{
			config: c,
			conn:   conn,
		},
		Media: &mediaDB{
			config: c,
			conn:   conn,
		},
		Mention: &mentionDB{
			config: c,
			conn:   conn,
			cache:  ttlcache.NewCache(),
		},
		Notification: &notificationDB{
			config: c,
			conn:   conn,
			cache:  ttlcache.NewCache(),
		},
		Relationship: &relationshipDB{
			config: c,
			conn:   conn,
		},
		Session: &sessionDB{
			config: c,
			conn:   conn,
		},
		Status: &statusDB{
			config: c,
			conn:   conn,
			cache:  cache.NewStatusCache(),
		},
		Timeline: &timelineDB{
			config: c,
			conn:   conn,
		},
		config: c,
		conn:   conn,
	}

	// we can confidently return this useable service now
	return ps, nil
}

/*
	HANDY STUFF
*/

// deriveBunDBPGOptions takes an application config and returns either a ready-to-use set of options
// with sensible defaults, or an error if it's not satisfied by the provided config.
func deriveBunDBPGOptions(c *config.Config) (*pgx.ConnConfig, error) {
	if strings.ToUpper(c.DBConfig.Type) != db.DBTypePostgres {
		return nil, fmt.Errorf("expected db type of %s but got %s", db.DBTypePostgres, c.DBConfig.Type)
	}

	// validate port
	if c.DBConfig.Port == 0 {
		return nil, errors.New("no port set")
	}

	// validate address
	if c.DBConfig.Address == "" {
		return nil, errors.New("no address set")
	}

	// validate username
	if c.DBConfig.User == "" {
		return nil, errors.New("no user set")
	}

	// validate that there's a password
	if c.DBConfig.Password == "" {
		return nil, errors.New("no password set")
	}

	// validate database
	if c.DBConfig.Database == "" {
		return nil, errors.New("no database set")
	}

	var tlsConfig *tls.Config
	switch c.DBConfig.TLSMode {
	case config.DBTLSModeDisable, config.DBTLSModeUnset:
		break // nothing to do
	case config.DBTLSModeEnable:
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	case config.DBTLSModeRequire:
		tlsConfig = &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         c.DBConfig.Address,
		}
	}

	if tlsConfig != nil && c.DBConfig.TLSCACert != "" {
		// load the system cert pool first -- we'll append the given CA cert to this
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("error fetching system CA cert pool: %s", err)
		}

		// open the file itself and make sure there's something in it
		caCertBytes, err := os.ReadFile(c.DBConfig.TLSCACert)
		if err != nil {
			return nil, fmt.Errorf("error opening CA certificate at %s: %s", c.DBConfig.TLSCACert, err)
		}
		if len(caCertBytes) == 0 {
			return nil, fmt.Errorf("ca cert at %s was empty", c.DBConfig.TLSCACert)
		}

		// make sure we have a PEM block
		caPem, _ := pem.Decode(caCertBytes)
		if caPem == nil {
			return nil, fmt.Errorf("could not parse cert at %s into PEM", c.DBConfig.TLSCACert)
		}

		// parse the PEM block into the certificate
		caCert, err := x509.ParseCertificate(caPem.Bytes)
		if err != nil {
			return nil, fmt.Errorf("could not parse cert at %s into x509 certificate: %s", c.DBConfig.TLSCACert, err)
		}

		// we're happy, add it to the existing pool and then use this pool in our tls config
		certPool.AddCert(caCert)
		tlsConfig.RootCAs = certPool
	}

	cfg, _ := pgx.ParseConfig("")
	cfg.Host = c.DBConfig.Address
	cfg.Port = uint16(c.DBConfig.Port)
	cfg.User = c.DBConfig.User
	cfg.Password = c.DBConfig.Password
	cfg.TLSConfig = tlsConfig
	cfg.Database = c.DBConfig.Database
	cfg.PreferSimpleProtocol = true

	return cfg, nil
}

/*
	CONVERSION FUNCTIONS
*/

// TODO: move these to the type converter, it's bananas that they're here and not there

func (ps *bunDBService) MentionStringsToMentions(ctx context.Context, targetAccounts []string, originAccountID string, statusID string) ([]*gtsmodel.Mention, error) {
	ogAccount := &gtsmodel.Account{}
	if err := ps.conn.NewSelect().Model(ogAccount).Where("id = ?", originAccountID).Scan(ctx); err != nil {
		return nil, err
	}

	menchies := []*gtsmodel.Mention{}
	for _, a := range targetAccounts {
		// A mentioned account looks like "@test@example.org" or just "@test" for a local account
		// -- we can guarantee this from the regex that targetAccounts should have been derived from.
		// But we still need to do a bit of fiddling to get what we need here -- the username and domain (if given).

		// 1.  trim off the first @
		t := strings.TrimPrefix(a, "@")

		// 2. split the username and domain
		s := strings.Split(t, "@")

		// 3. if it's length 1 it's a local account, length 2 means remote, anything else means something is wrong
		var local bool
		switch len(s) {
		case 1:
			local = true
		case 2:
			local = false
		default:
			return nil, fmt.Errorf("mentioned account format '%s' was not valid", a)
		}

		var username, domain string
		username = s[0]
		if !local {
			domain = s[1]
		}

		// 4. check we now have a proper username and domain
		if username == "" || (!local && domain == "") {
			return nil, fmt.Errorf("username or domain for '%s' was nil", a)
		}

		// okay we're good now, we can start pulling accounts out of the database
		mentionedAccount := &gtsmodel.Account{}
		var err error

		// match username + account, case insensitive
		if local {
			// local user -- should have a null domain
			err = ps.conn.NewSelect().Model(mentionedAccount).Where("LOWER(?) = LOWER(?)", bun.Ident("username"), username).Where("? IS NULL", bun.Ident("domain")).Scan(ctx)
		} else {
			// remote user -- should have domain defined
			err = ps.conn.NewSelect().Model(mentionedAccount).Where("LOWER(?) = LOWER(?)", bun.Ident("username"), username).Where("LOWER(?) = LOWER(?)", bun.Ident("domain"), domain).Scan(ctx)
		}

		if err != nil {
			if err == sql.ErrNoRows {
				// no result found for this username/domain so just don't include it as a mencho and carry on about our business
				ps.conn.log.Debugf("no account found with username '%s' and domain '%s', skipping it", username, domain)
				continue
			}
			// a serious error has happened so bail
			return nil, fmt.Errorf("error getting account with username '%s' and domain '%s': %s", username, domain, err)
		}

		// id, createdAt and updatedAt will be populated by the db, so we have everything we need!
		menchies = append(menchies, &gtsmodel.Mention{
			StatusID:         statusID,
			OriginAccountID:  ogAccount.ID,
			OriginAccountURI: ogAccount.URI,
			TargetAccountID:  mentionedAccount.ID,
			NameString:       a,
			TargetAccountURI: mentionedAccount.URI,
			TargetAccountURL: mentionedAccount.URL,
			OriginAccount:    mentionedAccount,
		})
	}
	return menchies, nil
}

func (ps *bunDBService) TagStringsToTags(ctx context.Context, tags []string, originAccountID string, statusID string) ([]*gtsmodel.Tag, error) {
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
				tag.URL = fmt.Sprintf("%s://%s/tags/%s", ps.config.Protocol, ps.config.Host, t)
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

func (ps *bunDBService) EmojiStringsToEmojis(ctx context.Context, emojis []string, originAccountID string, statusID string) ([]*gtsmodel.Emoji, error) {
	newEmojis := []*gtsmodel.Emoji{}
	for _, e := range emojis {
		emoji := &gtsmodel.Emoji{}
		err := ps.conn.NewSelect().Model(emoji).Where("shortcode = ?", e).Where("visible_in_picker = true").Where("disabled = false").Scan(ctx)
		if err != nil {
			if err == sql.ErrNoRows {
				// no result found for this username/domain so just don't include it as an emoji and carry on about our business
				ps.conn.log.Debugf("no emoji found with shortcode %s, skipping it", e)
				continue
			}
			// a serious error has happened so bail
			return nil, fmt.Errorf("error getting emoji with shortcode %s: %s", e, err)
		}
		newEmojis = append(newEmojis, emoji)
	}
	return newEmojis, nil
}

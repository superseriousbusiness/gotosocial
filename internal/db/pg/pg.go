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

package pg

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-pg/pg/extra/pgdebug"
	"github.com/go-pg/pg/v10"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

// postgresService satisfies the DB interface
type postgresService struct {
	db.Account
	db.Admin
	db.Basic
	db.Instance
	db.Notification
	db.Relationship
	db.Status
	db.Timeline
	config *config.Config
	conn   *pg.DB
	log    *logrus.Logger
	cancel context.CancelFunc
}

// NewPostgresService returns a postgresService derived from the provided config, which implements the go-fed DB interface.
// Under the hood, it uses https://github.com/go-pg/pg to create and maintain a database connection.
func NewPostgresService(ctx context.Context, c *config.Config, log *logrus.Logger) (db.DB, error) {
	opts, err := derivePGOptions(c)
	if err != nil {
		return nil, fmt.Errorf("could not create postgres service: %s", err)
	}
	log.Debugf("using pg options: %+v", opts)

	// create a connection
	pgCtx, cancel := context.WithCancel(ctx)
	conn := pg.Connect(opts).WithContext(pgCtx)

	// this will break the logfmt format we normally log in,
	// since we can't choose where pg outputs to and it defaults to
	// stdout. So use this option with care!
	if log.GetLevel() >= logrus.TraceLevel {
		conn.AddQueryHook(pgdebug.DebugHook{
			// Print all queries.
			Verbose: true,
		})
	}

	// actually *begin* the connection so that we can tell if the db is there and listening
	if err := conn.Ping(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("db connection error: %s", err)
	}

	// print out discovered postgres version
	var version string
	if _, err = conn.QueryOneContext(ctx, pg.Scan(&version), "SELECT version()"); err != nil {
		cancel()
		return nil, fmt.Errorf("db connection error: %s", err)
	}
	log.Infof("connected to postgres version: %s", version)

	ps := &postgresService{
		Account: &accountDB{
			config: c,
			conn:   conn,
			log:    log,
			cancel: cancel,
		},
		Admin: &adminDB{
			config: c,
			conn:   conn,
			log:    log,
			cancel: cancel,
		},
		Basic: &basicDB{
			config: c,
			conn:   conn,
			log:    log,
			cancel: cancel,
		},
		Instance: &instanceDB{
			config: c,
			conn:   conn,
			log:    log,
			cancel: cancel,
		},
		Relationship: &relationshipDB{
			config: c,
			conn:   conn,
			log:    log,
			cancel: cancel,
		},
		Status: &statusDB{
			config: c,
			conn:   conn,
			log:    log,
			cancel: cancel,
		},
		Timeline: &timelineDB{
			config: c,
			conn:   conn,
			log:    log,
			cancel: cancel,
		},
		config: c,
		conn:   conn,
		log:    log,
		cancel: cancel,
	}

	// we can confidently return this useable postgres service now
	return ps, nil
}

/*
	HANDY STUFF
*/

// derivePGOptions takes an application config and returns either a ready-to-use *pg.Options
// with sensible defaults, or an error if it's not satisfied by the provided config.
func derivePGOptions(c *config.Config) (*pg.Options, error) {
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

	// We can rely on the pg library we're using to set
	// sensible defaults for everything we don't set here.
	options := &pg.Options{
		Addr:            fmt.Sprintf("%s:%d", c.DBConfig.Address, c.DBConfig.Port),
		User:            c.DBConfig.User,
		Password:        c.DBConfig.Password,
		Database:        c.DBConfig.Database,
		ApplicationName: c.ApplicationName,
		TLSConfig:       tlsConfig,
	}

	return options, nil
}

/*
	CONVERSION FUNCTIONS
*/

// TODO: move these to the type converter, it's bananas that they're here and not there

func (ps *postgresService) MentionStringsToMentions(targetAccounts []string, originAccountID string, statusID string) ([]*gtsmodel.Mention, error) {
	ogAccount := &gtsmodel.Account{}
	if err := ps.conn.Model(ogAccount).Where("id = ?", originAccountID).Select(); err != nil {
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
			err = ps.conn.Model(mentionedAccount).Where("LOWER(?) = LOWER(?)", pg.Ident("username"), username).Where("? IS NULL", pg.Ident("domain")).Select()
		} else {
			// remote user -- should have domain defined
			err = ps.conn.Model(mentionedAccount).Where("LOWER(?) = LOWER(?)", pg.Ident("username"), username).Where("LOWER(?) = LOWER(?)", pg.Ident("domain"), domain).Select()
		}

		if err != nil {
			if err == pg.ErrNoRows {
				// no result found for this username/domain so just don't include it as a mencho and carry on about our business
				ps.log.Debugf("no account found with username '%s' and domain '%s', skipping it", username, domain)
				continue
			}
			// a serious error has happened so bail
			return nil, fmt.Errorf("error getting account with username '%s' and domain '%s': %s", username, domain, err)
		}

		// id, createdAt and updatedAt will be populated by the db, so we have everything we need!
		menchies = append(menchies, &gtsmodel.Mention{
			StatusID:            statusID,
			OriginAccountID:     ogAccount.ID,
			OriginAccountURI:    ogAccount.URI,
			TargetAccountID:     mentionedAccount.ID,
			NameString:          a,
			MentionedAccountURI: mentionedAccount.URI,
			MentionedAccountURL: mentionedAccount.URL,
			GTSAccount:          mentionedAccount,
		})
	}
	return menchies, nil
}

func (ps *postgresService) TagStringsToTags(tags []string, originAccountID string, statusID string) ([]*gtsmodel.Tag, error) {
	newTags := []*gtsmodel.Tag{}
	for _, t := range tags {
		tag := &gtsmodel.Tag{}
		// we can use selectorinsert here to create the new tag if it doesn't exist already
		// inserted will be true if this is a new tag we just created
		if err := ps.conn.Model(tag).Where("LOWER(?) = LOWER(?)", pg.Ident("name"), t).Select(); err != nil {
			if err == pg.ErrNoRows {
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

func (ps *postgresService) EmojiStringsToEmojis(emojis []string, originAccountID string, statusID string) ([]*gtsmodel.Emoji, error) {
	newEmojis := []*gtsmodel.Emoji{}
	for _, e := range emojis {
		emoji := &gtsmodel.Emoji{}
		err := ps.conn.Model(emoji).Where("shortcode = ?", e).Where("visible_in_picker = true").Where("disabled = false").Select()
		if err != nil {
			if err == pg.ErrNoRows {
				// no result found for this username/domain so just don't include it as an emoji and carry on about our business
				ps.log.Debugf("no emoji found with shortcode %s, skipping it", e)
				continue
			}
			// a serious error has happened so bail
			return nil, fmt.Errorf("error getting emoji with shortcode %s: %s", e, err)
		}
		newEmojis = append(newEmojis, emoji)
	}
	return newEmojis, nil
}

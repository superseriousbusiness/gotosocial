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

package db

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-pg/pg"
	"github.com/gotosocial/gotosocial/internal/config"
	"github.com/sirupsen/logrus"
)

type postgresService struct {
	config *config.DBConfig
	conn   *pg.DB
	log    *logrus.Entry
	cancel context.CancelFunc
}

// newPostgresService returns a postgresService derived from the provided config, which implements the go-fed DB interface.
// Under the hood, it uses https://github.com/go-pg/pg to create and maintain a database connection.
func newPostgresService(ctx context.Context, c *config.Config, log *logrus.Entry) (*postgresService, error) {
	opts, err := derivePGOptions(c)
	if err != nil {
		return nil, fmt.Errorf("could not create postgres service: %s", err)
	}

	readyChan := make(chan interface{})
	opts.OnConnect = func(c *pg.Conn) error {
		close(readyChan)
		return nil
	}

	// create a connection
	pgCtx, cancel := context.WithCancel(ctx)
	conn := pg.Connect(opts).WithContext(pgCtx)

	// actually *begin* the connection so that we can tell if the db is there
	// and listening, and also trigger the opts.OnConnect function passed in above
	tx, err := conn.Begin()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("db connection error: %s", err)
	}

	// close the transaction we just started so it doesn't hang around
	if err := tx.Rollback(); err != nil {
		cancel()
		return nil, fmt.Errorf("db connection error: %s", err)
	}

	// make sure the opts.OnConnect function has been triggered
	// and closed the ready channel
	select {
	case <-readyChan:
		log.Infof("postgres connection ready")
	case <-time.After(5 * time.Second):
		cancel()
		return nil, errors.New("db connection timeout")
	}

	// we can confidently return this useable postgres service now
	return &postgresService{
		config: c.DBConfig,
		conn:   conn,
		log:    log,
		cancel: cancel,
	}, nil
}

/*
	HANDY STUFF
*/

// derivePGOptions takes an application config and returns either a ready-to-use *pg.Options
// with sensible defaults, or an error if it's not satisfied by the provided config.
func derivePGOptions(c *config.Config) (*pg.Options, error) {
	if strings.ToUpper(c.DBConfig.Type) != dbTypePostgres {
		return nil, fmt.Errorf("expected db type of %s but got %s", dbTypePostgres, c.DBConfig.Type)
	}

	// validate port
	if c.DBConfig.Port == 0 {
		return nil, errors.New("no port set")
	}

	// validate address
	if c.DBConfig.Address == "" {
		return nil, errors.New("no address set")
	}

	ipv4Regex := regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	hostnameRegex := regexp.MustCompile(`^(?:[a-z0-9]+(?:-[a-z0-9]+)*\.)+[a-z]{2,}$`)
	if !hostnameRegex.MatchString(c.DBConfig.Address) && !ipv4Regex.MatchString(c.DBConfig.Address) && c.DBConfig.Address != "localhost" {
		return nil, fmt.Errorf("address %s was neither an ipv4 address nor a valid hostname", c.DBConfig.Address)
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

	// We can rely on the pg library we're using to set
	// sensible defaults for everything we don't set here.
	options := &pg.Options{
		Addr:            fmt.Sprintf("%s:%d", c.DBConfig.Address, c.DBConfig.Port),
		User:            c.DBConfig.User,
		Password:        c.DBConfig.Password,
		Database:        c.DBConfig.Database,
		ApplicationName: c.ApplicationName,
	}

	return options, nil
}

/*
   GO-FED DB INTERFACE-IMPLEMENTING FUNCTIONS
*/
func (ps *postgresService) Lock(ctx context.Context, id *url.URL) error {
	return nil
}

func (ps *postgresService) Unlock(ctx context.Context, id *url.URL) error {
	return nil
}

func (ps *postgresService) InboxContains(ctx context.Context, inbox *url.URL, id *url.URL) (bool, error) {
	return false, nil
}

func (ps *postgresService) GetInbox(ctx context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	return nil, nil
}

func (ps *postgresService) SetInbox(ctx context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	return nil
}

func (ps *postgresService) Owns(ctx context.Context, id *url.URL) (owns bool, err error) {
	return false, nil
}

func (ps *postgresService) ActorForOutbox(ctx context.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return nil, nil
}

func (ps *postgresService) ActorForInbox(ctx context.Context, inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return nil, nil
}

func (ps *postgresService) OutboxForInbox(ctx context.Context, inboxIRI *url.URL) (outboxIRI *url.URL, err error) {
	return nil, nil
}

func (ps *postgresService) Exists(ctx context.Context, id *url.URL) (exists bool, err error) {
	return false, nil
}

func (ps *postgresService) Get(ctx context.Context, id *url.URL) (value vocab.Type, err error) {
	return nil, nil
}

func (ps *postgresService) Create(ctx context.Context, asType vocab.Type) error {
	return nil
}

func (ps *postgresService) Update(ctx context.Context, asType vocab.Type) error {
	return nil
}

func (ps *postgresService) Delete(ctx context.Context, id *url.URL) error {
	return nil
}

func (ps *postgresService) GetOutbox(ctx context.Context, outboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	return nil, nil
}

func (ps *postgresService) SetOutbox(ctx context.Context, outbox vocab.ActivityStreamsOrderedCollectionPage) error {
	return nil
}

func (ps *postgresService) NewID(ctx context.Context, t vocab.Type) (id *url.URL, err error) {
	return nil, nil
}

func (ps *postgresService) Followers(ctx context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return nil, nil
}

func (ps *postgresService) Following(ctx context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return nil, nil
}

func (ps *postgresService) Liked(ctx context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return nil, nil
}

/*
	EXTRA FUNCTIONS
*/

func (ps *postgresService) Stop(ctx context.Context) error {
	ps.log.Info("closing db connection")
	if err := ps.conn.Close(); err != nil {
		// only cancel if there's a problem closing the db
		ps.cancel()
		return err
	}
	return nil
}

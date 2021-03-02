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

	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-pg/pg"
)

type postgresService struct {
	config *Config
	conn   *pg.DB
	ready  bool
}

// newPostgresService returns a postgresService derived from the provided config, which implements the go-fed DB interface.
// Under the hood, it uses https://github.com/go-pg/pg to create and maintain a database connection.
func newPostgresService(config *Config) (*postgresService, error) {
	opts, err := derivePGOptions(config)
	if err != nil {
		return nil, fmt.Errorf("could not create postgres service: %s", err)
	}
	conn := pg.Connect(opts)
	return &postgresService{
		config,
		conn,
		false,
	}, nil

}

/*
	HANDY STUFF
*/

// derivePGOptions takes an application config and returns either a ready-to-use *pg.Options
// with sensible defaults, or an error if it's not satisfied by the provided config.
func derivePGOptions(config *Config) (*pg.Options, error) {
	if config.Type != dbTypePostgres {
		return nil, fmt.Errorf("expected db type of %s but got %s", dbTypePostgres, config.Type)
	}

	// use sensible default port
	var port int = config.Port
	if port == 0 {
		port = postgresDefaultPort
	}

	// validate address
	address := config.Address
	if address == "" {
		return nil, errors.New("address not provided")
	}
	if !hostnameRegex.MatchString(address) && !ipv4Regex.MatchString(address) {
		return nil, fmt.Errorf("address %s was neither an ipv4 address nor a valid hostname", address)
	}

	options := &pg.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Address, config.Port),
		User:     config.User,
		Password: config.Password,
		Database: config.Database,
		OnConnect: func(c *pg.Conn) error {
			return nil
		},
	}

	return options, nil
}

/*
   GO-FED DB INTERFACE-IMPLEMENTING FUNCTIONS
*/
func (ps *postgresService) Lock(c context.Context, id *url.URL) error {
	return nil
}

func (ps *postgresService) Unlock(c context.Context, id *url.URL) error {
	return nil
}

func (ps *postgresService) InboxContains(c context.Context, inbox *url.URL, id *url.URL) (bool, error) {
	return false, nil
}

func (ps *postgresService) GetInbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	return nil, nil
}

func (ps *postgresService) SetInbox(c context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	return nil
}

func (ps *postgresService) Owns(c context.Context, id *url.URL) (owns bool, err error) {
	return false, nil
}

func (ps *postgresService) ActorForOutbox(c context.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return nil, nil
}

func (ps *postgresService) ActorForInbox(c context.Context, inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return nil, nil
}

func (ps *postgresService) OutboxForInbox(c context.Context, inboxIRI *url.URL) (outboxIRI *url.URL, err error) {
	return nil, nil
}

func (ps *postgresService) Exists(c context.Context, id *url.URL) (exists bool, err error) {
	return false, nil
}

func (ps *postgresService) Get(c context.Context, id *url.URL) (value vocab.Type, err error) {
	return nil, nil
}

func (ps *postgresService) Create(c context.Context, asType vocab.Type) error {
	return nil
}

func (ps *postgresService) Update(c context.Context, asType vocab.Type) error {
	return nil
}

func (ps *postgresService) Delete(c context.Context, id *url.URL) error {
	return nil
}

func (ps *postgresService) GetOutbox(c context.Context, outboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	return nil, nil
}

func (ps *postgresService) SetOutbox(c context.Context, outbox vocab.ActivityStreamsOrderedCollectionPage) error {
	return nil
}

func (ps *postgresService) NewID(c context.Context, t vocab.Type) (id *url.URL, err error) {
	return nil, nil
}

func (ps *postgresService) Followers(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return nil, nil
}

func (ps *postgresService) Following(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return nil, nil
}

func (ps *postgresService) Liked(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return nil, nil
}

/*
	EXTRA FUNCTIONS
*/

func (ps *postgresService) Ready() bool {
	return false
}

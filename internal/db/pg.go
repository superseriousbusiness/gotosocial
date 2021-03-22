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
	"regexp"
	"strings"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/go-pg/pg/extra/pgdebug"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/gotosocial/gotosocial/internal/config"
	"github.com/gotosocial/gotosocial/internal/gtsmodel"
	"github.com/gotosocial/oauth2/v4"
	"github.com/sirupsen/logrus"
)

// postgresService satisfies the DB interface
type postgresService struct {
	config       *config.DBConfig
	conn         *pg.DB
	log          *logrus.Entry
	cancel       context.CancelFunc
	tokenStore   oauth2.TokenStore
	federationDB pub.Database
}

// newPostgresService returns a postgresService derived from the provided config, which implements the go-fed DB interface.
// Under the hood, it uses https://github.com/go-pg/pg to create and maintain a database connection.
func newPostgresService(ctx context.Context, c *config.Config, log *logrus.Entry) (*postgresService, error) {
	opts, err := derivePGOptions(c)
	if err != nil {
		return nil, fmt.Errorf("could not create postgres service: %s", err)
	}
	log.Debugf("using pg options: %+v", opts)

	readyChan := make(chan interface{})
	opts.OnConnect = func(ctx context.Context, c *pg.Conn) error {
		close(readyChan)
		return nil
	}

	// create a connection
	pgCtx, cancel := context.WithCancel(ctx)
	conn := pg.Connect(opts).WithContext(pgCtx)

	// this will break the logfmt format we normally log in,
	// since we can't choose where pg outputs to and it defaults to
	// stdout. So use this option with care!
	if log.Logger.GetLevel() >= logrus.TraceLevel {
		conn.AddQueryHook(pgdebug.DebugHook{
			// Print all queries.
			Verbose: true,
		})
	}

	// actually *begin* the connection so that we can tell if the db is there
	// and listening, and also trigger the opts.OnConnect function passed in above
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
		config:       c.DBConfig,
		conn:         conn,
		log:          log,
		cancel:       cancel,
		federationDB: newPostgresFederation(conn),
	}, nil
}

func (ps *postgresService) Federation() pub.Database {
	return ps.federationDB
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

func (ps *postgresService) CreateSchema(ctx context.Context) error {
	models := []interface{}{
		(*gtsmodel.Account)(nil),
		(*gtsmodel.Status)(nil),
		(*gtsmodel.User)(nil),
	}
	ps.log.Info("creating db schema")

	for _, model := range models {
		err := ps.conn.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}

	ps.log.Info("db schema created")
	return nil
}

func (ps *postgresService) IsHealthy(ctx context.Context) error {
	return ps.conn.Ping(ctx)
}

func (ps *postgresService) CreateTable(i interface{}) error {
	return ps.conn.Model(i).CreateTable(&orm.CreateTableOptions{
		IfNotExists: true,
	})
}

func (ps *postgresService) DropTable(i interface{}) error {
	return ps.conn.Model(i).DropTable(&orm.DropTableOptions{
		IfExists: true,
	})
}

func (ps *postgresService) GetByID(id string, i interface{}) error {
	return ps.conn.Model(i).Where("id = ?", id).Select()
}

func (ps *postgresService) GetWhere(key string, value interface{}, i interface{}) error {
	return ps.conn.Model(i).Where(fmt.Sprintf("%s = ?", key), value).Select()
}

func (ps *postgresService) GetAll(i interface{}) error {
	return ps.conn.Model(i).Select()
}

func (ps *postgresService) Put(i interface{}) error {
	_, err := ps.conn.Model(i).Insert(i)
	return err
}

func (ps *postgresService) UpdateByID(id string, i interface{}) error {
	_, err := ps.conn.Model(i).OnConflict("(id) DO UPDATE").Insert()
	return err
}

func (ps *postgresService) DeleteByID(id string, i interface{}) error {
	_, err := ps.conn.Model(i).Where("id = ?", id).Delete()
	return err
}

func (ps *postgresService) DeleteWhere(key string, value interface{}, i interface{}) error {
	_, err := ps.conn.Model(i).Where(fmt.Sprintf("%s = ?", key), value).Delete()
	return err
}

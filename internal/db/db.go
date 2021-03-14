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
	"fmt"
	"strings"

	"github.com/go-fed/activity/pub"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/gotosocial/gotosocial/internal/config"
	"github.com/sirupsen/logrus"
)

const dbTypePostgres string = "POSTGRES"

// DB provides methods for interacting with an underlying database (for now, just postgres).
// The function mapping lines up with the DB interface described in go-fed.
// See here: https://github.com/go-fed/activity/blob/master/pub/database.go
type DB interface {
	/*
		GO-FED DATABASE FUNCTIONS
	*/
	pub.Database

	/*
		OAUTH2 DATABASE FUNCTIONS
	*/
	TokenStore() oauth2.TokenStore

	/*
		ANY ADDITIONAL DESIRED FUNCTIONS
	*/

	// CreateSchema should populate the database with the required tables
	CreateSchema(context.Context) error

	// Stop should stop and close the database connection cleanly, returning an error if this is not possible
	Stop(context.Context) error

	// IsHealthy should return nil if the database connection is healthy, or an error if not
	IsHealthy(context.Context) error
}

// New returns a new database service that satisfies the Service interface and, by extension,
// the go-fed database interface described here: https://github.com/go-fed/activity/blob/master/pub/database.go
func New(ctx context.Context, c *config.Config, log *logrus.Logger) (DB, error) {
	switch strings.ToUpper(c.DBConfig.Type) {
	case dbTypePostgres:
		return newPostgresService(ctx, c, log.WithField("service", "db"))
	default:
		return nil, fmt.Errorf("database type %s not supported", c.DBConfig.Type)
	}
}

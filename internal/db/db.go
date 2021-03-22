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
	"github.com/gotosocial/gotosocial/internal/config"
	"github.com/sirupsen/logrus"
)

const dbTypePostgres string = "POSTGRES"

// DB provides methods for interacting with an underlying database or other storage mechanism (for now, just postgres).
type DB interface {
	// Federation returns an interface that's compatible with go-fed, for performing federation storage/retrieval functions.
	// See: https://pkg.go.dev/github.com/go-fed/activity@v1.0.0/pub?utm_source=gopls#Database
	Federation() pub.Database

	// CreateTable creates a table for the given interface
	CreateTable(i interface{}) error

	// DropTable drops the table for the given interface
	DropTable(i interface{}) error

	// Stop should stop and close the database connection cleanly, returning an error if this is not possible
	Stop(ctx context.Context) error

	// IsHealthy should return nil if the database connection is healthy, or an error if not
	IsHealthy(ctx context.Context) error

	// GetByID gets one entry by its id.
	GetByID(id string, i interface{}) error

	// GetWhere gets one entry where key = value
	GetWhere(key string, value interface{}, i interface{}) error

	// GetAll gets all entries of interface type i
	GetAll(i interface{}) error

	// Put stores i
	Put(i interface{}) error

	// Update by id updates i with id id
	UpdateByID(id string, i interface{}) error

	// Delete by id removes i with id id
	DeleteByID(id string, i interface{}) error

	// Delete where deletes i where key = value
	DeleteWhere(key string, value interface{}, i interface{}) error
}

// New returns a new database service that satisfies the DB interface and, by extension,
// the go-fed database interface described here: https://github.com/go-fed/activity/blob/master/pub/database.go
func New(ctx context.Context, c *config.Config, log *logrus.Logger) (DB, error) {
	switch strings.ToUpper(c.DBConfig.Type) {
	case dbTypePostgres:
		return newPostgresService(ctx, c, log.WithField("service", "db"))
	default:
		return nil, fmt.Errorf("database type %s not supported", c.DBConfig.Type)
	}
}

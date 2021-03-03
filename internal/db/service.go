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
	"github.com/sirupsen/logrus"
)

const dbTypePostgres string = "POSTGRES"

// Service provides methods for interacting with an underlying database (for now, just postgres).
// The function mapping lines up with the Database interface described in go-fed.
// See here: https://github.com/go-fed/activity/blob/master/pub/database.go
type Service interface {
	/*
		GO-FED DATABASE FUNCTIONS
	*/
	pub.Database

	/*
		ANY ADDITIONAL DESIRED FUNCTIONS
	*/
	Stop(context.Context) error
}

// Config provides configuration options for the database connection
type Config struct {
	Type            string `json:"type,omitempty"`
	Address         string `json:"address,omitempty"`
	Port            int    `json:"port,omitempty"`
	User            string `json:"user,omitempty"`
	Password        string `json:"password,omitempty"`
	PasswordFile    string `json:"passwordFile,omitempty"`
	Database        string `json:"database,omitempty"`
	ApplicationName string `json:"applicationName,omitempty"`
}

// NewService returns a new database service that satisfies the Service interface and, by extension,
// the go-fed database interface described here: https://github.com/go-fed/activity/blob/master/pub/database.go
func NewService(context context.Context, config *Config, log *logrus.Logger) (Service, error) {
	switch strings.ToUpper(config.Type) {
	case dbTypePostgres:
		return newPostgresService(context, config, log.WithField("service", "db"))
	default:
		return nil, fmt.Errorf("database type %s not supported", config.Type)
	}
}

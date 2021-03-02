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
	"fmt"
	"strings"

	"github.com/go-fed/activity/pub"
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

	// Ready indicates whether the database is ready to handle queries and whatnot.
	Ready() bool
}

// Config provides configuration options for the database connection
type Config struct {
	Type            string
	Address         string
	Port            int
	User            string
	Password        string
	Database        string
	ApplicationName string
}

// NewService returns a new database service that satisfies the Service interface and, by extension,
// the go-fed database interface described here: https://github.com/go-fed/activity/blob/master/pub/database.go
func NewService(config *Config) (Service, error) {
	switch strings.ToUpper(config.Type) {
	case dbTypePostgres:
		return newPostgresService(config)
	default:
		return nil, fmt.Errorf("database type %s not supported", config.Type)
	}
}

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

package config

// DBConfig provides configuration options for the database connection
type DBConfig struct {
	Type            string    `yaml:"type"`
	Address         string    `yaml:"address"`
	Port            int       `yaml:"port"`
	User            string    `yaml:"user"`
	Password        string    `yaml:"password"`
	Database        string    `yaml:"database"`
	ApplicationName string    `yaml:"applicationName"`
	TLSMode         DBTLSMode `yaml:"tlsMode"`
	TLSCACert       string    `yaml:"tlsCACert"`
}

// DBTLSMode describes a mode of connecting to a database with or without TLS.
type DBTLSMode string

// DBTLSModeDisable does not attempt to make a TLS connection to the database.
var DBTLSModeDisable DBTLSMode = "disable"

// DBTLSModeEnable attempts to make a TLS connection to the database, but doesn't fail if
// the certificate passed by the database isn't verified.
var DBTLSModeEnable DBTLSMode = "enable"

// DBTLSModeRequire attempts to make a TLS connection to the database, and requires
// that the certificate presented by the database is valid.
var DBTLSModeRequire DBTLSMode = "require"

// DBTLSModeUnset means that the TLS mode has not been set.
var DBTLSModeUnset DBTLSMode = ""

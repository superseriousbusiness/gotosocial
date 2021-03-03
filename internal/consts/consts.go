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

// Package consts is where we shove any consts that don't really belong anywhere else in the code.
// Don't judge me.
package consts

// FlagNames is used for storing the names of the various flags used for
// initializing and storing urfavecli flag variables.
type FlagNames struct {
	LogLevel        string
	ApplicationName string
	DbType          string
	DbAddress       string
	DbPort          string
	DbUser          string
	DbPassword      string
	DbDatabase      string
}

// GetFlagNames returns a struct containing the names of the various flags used for
// initializing and storing urfavecli flag variables.
func GetFlagNames() FlagNames {
	return FlagNames{
		LogLevel:        "log-level",
		ApplicationName: "application-name",
		DbType:          "db-type",
		DbAddress:       "db-address",
		DbPort:          "db-port",
		DbUser:          "db-users",
		DbPassword:      "db-password",
		DbDatabase:      "db-database",
	}
}

// EnvNames is used for storing the environment variable keys used for
// initializing and storing urfavecli flag variables.
type EnvNames struct {
	LogLevel        string
	ApplicationName string
	DbType          string
	DbAddress       string
	DbPort          string
	DbUser          string
	DbPassword      string
	DbDatabase      string
}

// GetEnvNames returns a struct containing the names of the environment variable keys used for
// initializing and storing urfavecli flag variables.
func GetEnvNames() FlagNames {
	return FlagNames{
		LogLevel:        "GTS_LOG_LEVEL",
		ApplicationName: "GTS_APPLICATION_NAME",
		DbType:          "GTS_DB_TYPE",
		DbAddress:       "GTS_DB_ADDRESS",
		DbPort:          "GTS_DB_PORT",
		DbUser:          "GTS_DB_USER",
		DbPassword:      "GTS_DB_PASSWORD",
		DbDatabase:      "GTS_DB_DATABASE",
	}
}

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

package main

import (
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/urfave/cli/v2"
)

func databaseFlags(flagNames, envNames config.Flags, defaults config.Defaults) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    flagNames.DbType,
			Usage:   "Database type: eg., postgres",
			Value:   defaults.DbType,
			EnvVars: []string{envNames.DbType},
		},
		&cli.StringFlag{
			Name:    flagNames.DbAddress,
			Usage:   "Database ipv4 address or hostname",
			Value:   defaults.DbAddress,
			EnvVars: []string{envNames.DbAddress},
		},
		&cli.IntFlag{
			Name:    flagNames.DbPort,
			Usage:   "Database port",
			Value:   defaults.DbPort,
			EnvVars: []string{envNames.DbPort},
		},
		&cli.StringFlag{
			Name:    flagNames.DbUser,
			Usage:   "Database username",
			Value:   defaults.DbUser,
			EnvVars: []string{envNames.DbUser},
		},
		&cli.StringFlag{
			Name:    flagNames.DbPassword,
			Usage:   "Database password",
			Value:   defaults.DbPassword,
			EnvVars: []string{envNames.DbPassword},
		},
		&cli.StringFlag{
			Name:    flagNames.DbDatabase,
			Usage:   "Database name",
			Value:   defaults.DbDatabase,
			EnvVars: []string{envNames.DbDatabase},
		},
		&cli.StringFlag{
			Name:    flagNames.DbTLSMode,
			Usage:   "Database tls mode",
			Value:   defaults.DBTlsMode,
			EnvVars: []string{envNames.DbTLSMode},
		},
		&cli.StringFlag{
			Name:    flagNames.DbTLSCACert,
			Usage:   "Path to CA cert for db tls connection",
			Value:   defaults.DBTlsCACert,
			EnvVars: []string{envNames.DbTLSCACert},
		},
	}
}

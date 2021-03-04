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
	"os"

	"github.com/gotosocial/gotosocial/internal/server"
	"github.com/gotosocial/gotosocial/internal/consts"
	"github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"
)

func main() {
	flagNames := consts.GetFlagNames()
	envNames := consts.GetEnvNames()
	app := &cli.App{
		Usage: "a fediverse social media server",
		Flags: []cli.Flag{
			// GENERAL FLAGS
			&cli.StringFlag{
				Name:    flagNames.LogLevel,
				Usage:   "Log level to run at: debug, info, warn, fatal",
				Value:   "info",
				EnvVars: []string{"GTS_LOG_LEVEL"},
			},
			&cli.StringFlag{
				Name:    flagNames.ApplicationName,
				Usage:   "Name of the application, used in various places internally",
				Value:   "gotosocial",
				EnvVars: []string{envNames.ApplicationName},
				Hidden:  true,
			},
			&cli.StringFlag{
				Name:    flagNames.ConfigPath,
				Usage:   "Path to a yaml file containing gotosocial configuration. Values set in this file will be overwritten by values set as env vars or arguments",
				Value:   "",
				EnvVars: []string{envNames.ConfigPath},
			},

			// DATABASE FLAGS
			&cli.StringFlag{
				Name:    flagNames.DbType,
				Usage:   "Database type: eg., postgres",
				Value:   "postgres",
				EnvVars: []string{envNames.DbType},
			},
			&cli.StringFlag{
				Name:    flagNames.DbAddress,
				Usage:   "Database ipv4 address or hostname",
				Value:   "localhost",
				EnvVars: []string{envNames.DbAddress},
			},
			&cli.IntFlag{
				Name:    flagNames.DbPort,
				Usage:   "Database port",
				Value:   5432,
				EnvVars: []string{envNames.DbPort},
			},
			&cli.StringFlag{
				Name:    flagNames.DbUser,
				Usage:   "Database username",
				Value:   "postgres",
				EnvVars: []string{envNames.DbUser},
			},
			&cli.StringFlag{
				Name:     flagNames.DbPassword,
				Usage:    "Database password",
				EnvVars:  []string{envNames.DbPassword},
			},
			&cli.StringFlag{
				Name:    flagNames.DbDatabase,
				Usage:   "Database name",
				Value:   "postgres",
				EnvVars: []string{envNames.DbDatabase},
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "server",
				Usage: "gotosocial server-related tasks",
				Subcommands: []*cli.Command{
					{
						Name:   "start",
						Usage:  "start the gotosocial server",
						Action: server.Run,
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}

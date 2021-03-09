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
	"fmt"
	"os"

	"github.com/gotosocial/gotosocial/internal/action"
	"github.com/gotosocial/gotosocial/internal/config"
	"github.com/gotosocial/gotosocial/internal/db"
	"github.com/gotosocial/gotosocial/internal/gotosocial"
	"github.com/gotosocial/gotosocial/internal/log"
	"github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"
)

func main() {
	flagNames := config.GetFlagNames()
	envNames := config.GetEnvNames()
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
				Name:    flagNames.DbPassword,
				Usage:   "Database password",
				EnvVars: []string{envNames.DbPassword},
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
						Name:  "start",
						Usage: "start the gotosocial server",
						Action: func(c *cli.Context) error {
							return runAction(c, gotosocial.Run)
						},
					},
				},
			},
			{
				Name:  "db",
				Usage: "database-related tasks and utils",
				Subcommands: []*cli.Command{
					{
						Name:  "init",
						Usage: "initialize a database with the required schema for gotosocial; has no effect & is safe to run on an already-initialized db",
						Action: func(c *cli.Context) error {
							return runAction(c, db.Initialize)
						},
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

// runAction builds up the config and logger necessary for any
// gotosocial action, and then executes the action.
func runAction(c *cli.Context, a action.GTSAction) error {

	// create a new *config.Config based on the config path provided...
	conf, err := config.New(c.String(config.GetFlagNames().ConfigPath))
	if err != nil {
		return fmt.Errorf("error creating config: %s", err)
	}
	// ... and the flags set on the *cli.Context by urfave
	conf.ParseFlags(c)

	// create a logger with the log level, formatting, and output splitter already set
	log, err := log.New(conf.LogLevel)
	if err != nil {
		return fmt.Errorf("error creating logger: %s", err)
	}

	return a(c.Context, conf, log)
}

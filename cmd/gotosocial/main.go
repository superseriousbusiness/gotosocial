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

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/action"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gotosocial"
	"github.com/superseriousbusiness/gotosocial/internal/log"

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
			&cli.StringFlag{
				Name:    flagNames.Host,
				Usage:   "Hostname to use for the server (eg., example.org, gotosocial.whatever.com)",
				Value:   "localhost",
				EnvVars: []string{envNames.Host},
			},
			&cli.StringFlag{
				Name:    flagNames.Protocol,
				Usage:   "Protocol to use for the REST api of the server (only use http for debugging and tests!)",
				Value:   "https",
				EnvVars: []string{envNames.Protocol},
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

			// TEMPLATE FLAGS
			&cli.StringFlag{
				Name:    flagNames.TemplateBaseDir,
				Usage:   "Basedir for html templating files for rendering pages and composing emails.",
				Value:   "./web/template/",
				EnvVars: []string{envNames.TemplateBaseDir},
			},

			// ACCOUNTS FLAGS
			&cli.BoolFlag{
				Name:    flagNames.AccountsOpenRegistration,
				Usage:   "Allow anyone to submit an account signup request. If false, server will be invite-only.",
				Value:   true,
				EnvVars: []string{envNames.AccountsOpenRegistration},
			},
			&cli.BoolFlag{
				Name:    flagNames.AccountsRequireApproval,
				Usage:   "Do account signups require approval by an admin or moderator before user can log in? If false, new registrations will be automatically approved.",
				Value:   true,
				EnvVars: []string{envNames.AccountsRequireApproval},
			},

			// MEDIA FLAGS
			&cli.IntFlag{
				Name:    flagNames.MediaMaxImageSize,
				Usage:   "Max size of accepted images in bytes",
				Value:   1048576, // 1mb
				EnvVars: []string{envNames.MediaMaxImageSize},
			},
			&cli.IntFlag{
				Name:    flagNames.MediaMaxVideoSize,
				Usage:   "Max size of accepted videos in bytes",
				Value:   5242880, // 5mb
				EnvVars: []string{envNames.MediaMaxVideoSize},
			},
			&cli.IntFlag{
				Name:    flagNames.MediaMinDescriptionChars,
				Usage:   "Min required chars for an image description",
				Value:   0,
				EnvVars: []string{envNames.MediaMinDescriptionChars},
			},
			&cli.IntFlag{
				Name:    flagNames.MediaMaxDescriptionChars,
				Usage:   "Max permitted chars for an image description",
				Value:   500,
				EnvVars: []string{envNames.MediaMaxDescriptionChars},
			},

			// STORAGE FLAGS
			&cli.StringFlag{
				Name:    flagNames.StorageBackend,
				Usage:   "Storage backend to use for media attachments",
				Value:   "local",
				EnvVars: []string{envNames.StorageBackend},
			},
			&cli.StringFlag{
				Name:    flagNames.StorageBasePath,
				Usage:   "Full path to an already-created directory where gts should store/retrieve media files. Subfolders will be created within this dir.",
				Value:   "/gotosocial/storage/media",
				EnvVars: []string{envNames.StorageBasePath},
			},
			&cli.StringFlag{
				Name:    flagNames.StorageServeProtocol,
				Usage:   "Protocol to use for serving media attachments (use https if storage is local)",
				Value:   "https",
				EnvVars: []string{envNames.StorageServeProtocol},
			},
			&cli.StringFlag{
				Name:    flagNames.StorageServeHost,
				Usage:   "Hostname to serve media attachments from (use the same value as host if storage is local)",
				Value:   "localhost",
				EnvVars: []string{envNames.StorageServeHost},
			},
			&cli.StringFlag{
				Name:    flagNames.StorageServeBasePath,
				Usage:   "Path to append to protocol and hostname to create the base path from which media files will be served (default will mostly be fine)",
				Value:   "/fileserver/media",
				EnvVars: []string{envNames.StorageServeBasePath},
			},

			// STATUSES FLAGS
			&cli.IntFlag{
				Name:    flagNames.StatusesMaxChars,
				Usage:   "Max permitted characters for posted statuses",
				Value:   5000,
				EnvVars: []string{envNames.StatusesMaxChars},
			},
			&cli.IntFlag{
				Name:    flagNames.StatusesCWMaxChars,
				Usage:   "Max permitted characters for content/spoiler warnings on statuses",
				Value:   100,
				EnvVars: []string{envNames.StatusesCWMaxChars},
			},
			&cli.IntFlag{
				Name:    flagNames.StatusesPollMaxOptions,
				Usage:   "Max amount of options permitted on a poll",
				Value:   6,
				EnvVars: []string{envNames.StatusesPollMaxOptions},
			},
			&cli.IntFlag{
				Name:    flagNames.StatusesPollOptionMaxChars,
				Usage:   "Max amount of characters for a poll option",
				Value:   50,
				EnvVars: []string{envNames.StatusesPollOptionMaxChars},
			},
			&cli.IntFlag{
				Name:    flagNames.StatusesMaxMediaFiles,
				Usage:   "Maximum number of media files/attachments per status",
				Value:   6,
				EnvVars: []string{envNames.StatusesMaxMediaFiles},
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
	conf, err := config.FromFile(c.String(config.GetFlagNames().ConfigPath))
	if err != nil {
		return fmt.Errorf("error creating config: %s", err)
	}
	// ... and the flags set on the *cli.Context by urfave
	conf.ParseCLIFlags(c)

	// create a logger with the log level, formatting, and output splitter already set
	log, err := log.New(conf.LogLevel)
	if err != nil {
		return fmt.Errorf("error creating logger: %s", err)
	}

	return a(c.Context, conf, log)
}

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
	"github.com/superseriousbusiness/gotosocial/internal/cliactions"
	"github.com/superseriousbusiness/gotosocial/internal/cliactions/admin/account"
	"github.com/superseriousbusiness/gotosocial/internal/cliactions/server"
	"github.com/superseriousbusiness/gotosocial/internal/cliactions/testrig"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"

	"github.com/urfave/cli/v2"
)

// Version is the software version of GtS being used
var Version string

// Commit is the git commit of GtS being used
var Commit string

func main() {
	flagNames := config.GetFlagNames()
	envNames := config.GetEnvNames()
	defaults := config.GetDefaults()
	app := &cli.App{
		Version: Version + " " + Commit[:7],
		Usage:   "a fediverse social media server",
		Flags: []cli.Flag{
			// GENERAL FLAGS
			&cli.StringFlag{
				Name:    flagNames.LogLevel,
				Usage:   "Log level to run at: debug, info, warn, fatal",
				Value:   defaults.LogLevel,
				EnvVars: []string{envNames.LogLevel},
			},
			&cli.StringFlag{
				Name:    flagNames.ApplicationName,
				Usage:   "Name of the application, used in various places internally",
				Value:   defaults.ApplicationName,
				EnvVars: []string{envNames.ApplicationName},
				Hidden:  true,
			},
			&cli.StringFlag{
				Name:    flagNames.ConfigPath,
				Usage:   "Path to a yaml file containing gotosocial configuration. Values set in this file will be overwritten by values set as env vars or arguments",
				Value:   defaults.ConfigPath,
				EnvVars: []string{envNames.ConfigPath},
			},
			&cli.StringFlag{
				Name:    flagNames.Host,
				Usage:   "Hostname to use for the server (eg., example.org, gotosocial.whatever.com)",
				Value:   defaults.Host,
				EnvVars: []string{envNames.Host},
			},
			&cli.StringFlag{
				Name:    flagNames.Protocol,
				Usage:   "Protocol to use for the REST api of the server (only use http for debugging and tests!)",
				Value:   defaults.Protocol,
				EnvVars: []string{envNames.Protocol},
			},

			// DATABASE FLAGS
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

			// TEMPLATE FLAGS
			&cli.StringFlag{
				Name:    flagNames.TemplateBaseDir,
				Usage:   "Basedir for html templating files for rendering pages and composing emails.",
				Value:   defaults.TemplateBaseDir,
				EnvVars: []string{envNames.TemplateBaseDir},
			},
			&cli.StringFlag{
				Name:    flagNames.AssetBaseDir,
				Usage:   "Directory to serve static assets from, accessible at example.com/assets/",
				Value:   defaults.AssetBaseDir,
				EnvVars: []string{envNames.AssetBaseDir},
			},

			// ACCOUNTS FLAGS
			&cli.BoolFlag{
				Name:    flagNames.AccountsOpenRegistration,
				Usage:   "Allow anyone to submit an account signup request. If false, server will be invite-only.",
				Value:   defaults.AccountsOpenRegistration,
				EnvVars: []string{envNames.AccountsOpenRegistration},
			},
			&cli.BoolFlag{
				Name:    flagNames.AccountsApprovalRequired,
				Usage:   "Do account signups require approval by an admin or moderator before user can log in? If false, new registrations will be automatically approved.",
				Value:   defaults.AccountsRequireApproval,
				EnvVars: []string{envNames.AccountsApprovalRequired},
			},
			&cli.BoolFlag{
				Name:    flagNames.AccountsReasonRequired,
				Usage:   "Do new account signups require a reason to be submitted on registration?",
				Value:   defaults.AccountsReasonRequired,
				EnvVars: []string{envNames.AccountsReasonRequired},
			},

			// MEDIA FLAGS
			&cli.IntFlag{
				Name:    flagNames.MediaMaxImageSize,
				Usage:   "Max size of accepted images in bytes",
				Value:   defaults.MediaMaxImageSize,
				EnvVars: []string{envNames.MediaMaxImageSize},
			},
			&cli.IntFlag{
				Name:    flagNames.MediaMaxVideoSize,
				Usage:   "Max size of accepted videos in bytes",
				Value:   defaults.MediaMaxVideoSize,
				EnvVars: []string{envNames.MediaMaxVideoSize},
			},
			&cli.IntFlag{
				Name:    flagNames.MediaMinDescriptionChars,
				Usage:   "Min required chars for an image description",
				Value:   defaults.MediaMinDescriptionChars,
				EnvVars: []string{envNames.MediaMinDescriptionChars},
			},
			&cli.IntFlag{
				Name:    flagNames.MediaMaxDescriptionChars,
				Usage:   "Max permitted chars for an image description",
				Value:   defaults.MediaMaxDescriptionChars,
				EnvVars: []string{envNames.MediaMaxDescriptionChars},
			},

			// STORAGE FLAGS
			&cli.StringFlag{
				Name:    flagNames.StorageBackend,
				Usage:   "Storage backend to use for media attachments",
				Value:   defaults.StorageBackend,
				EnvVars: []string{envNames.StorageBackend},
			},
			&cli.StringFlag{
				Name:    flagNames.StorageBasePath,
				Usage:   "Full path to an already-created directory where gts should store/retrieve media files. Subfolders will be created within this dir.",
				Value:   defaults.StorageBasePath,
				EnvVars: []string{envNames.StorageBasePath},
			},
			&cli.StringFlag{
				Name:    flagNames.StorageServeProtocol,
				Usage:   "Protocol to use for serving media attachments (use https if storage is local)",
				Value:   defaults.StorageServeProtocol,
				EnvVars: []string{envNames.StorageServeProtocol},
			},
			&cli.StringFlag{
				Name:    flagNames.StorageServeHost,
				Usage:   "Hostname to serve media attachments from (use the same value as host if storage is local)",
				Value:   defaults.StorageServeHost,
				EnvVars: []string{envNames.StorageServeHost},
			},
			&cli.StringFlag{
				Name:    flagNames.StorageServeBasePath,
				Usage:   "Path to append to protocol and hostname to create the base path from which media files will be served (default will mostly be fine)",
				Value:   defaults.StorageServeBasePath,
				EnvVars: []string{envNames.StorageServeBasePath},
			},

			// STATUSES FLAGS
			&cli.IntFlag{
				Name:    flagNames.StatusesMaxChars,
				Usage:   "Max permitted characters for posted statuses",
				Value:   defaults.StatusesMaxChars,
				EnvVars: []string{envNames.StatusesMaxChars},
			},
			&cli.IntFlag{
				Name:    flagNames.StatusesCWMaxChars,
				Usage:   "Max permitted characters for content/spoiler warnings on statuses",
				Value:   defaults.StatusesCWMaxChars,
				EnvVars: []string{envNames.StatusesCWMaxChars},
			},
			&cli.IntFlag{
				Name:    flagNames.StatusesPollMaxOptions,
				Usage:   "Max amount of options permitted on a poll",
				Value:   defaults.StatusesPollMaxOptions,
				EnvVars: []string{envNames.StatusesPollMaxOptions},
			},
			&cli.IntFlag{
				Name:    flagNames.StatusesPollOptionMaxChars,
				Usage:   "Max amount of characters for a poll option",
				Value:   defaults.StatusesPollOptionMaxChars,
				EnvVars: []string{envNames.StatusesPollOptionMaxChars},
			},
			&cli.IntFlag{
				Name:    flagNames.StatusesMaxMediaFiles,
				Usage:   "Maximum number of media files/attachments per status",
				Value:   defaults.StatusesMaxMediaFiles,
				EnvVars: []string{envNames.StatusesMaxMediaFiles},
			},

			// LETSENCRYPT FLAGS
			&cli.BoolFlag{
				Name:    flagNames.LetsEncryptEnabled,
				Usage:   "Enable letsencrypt TLS certs for this server. If set to true, then cert dir also needs to be set (or take the default).",
				Value:   defaults.LetsEncryptEnabled,
				EnvVars: []string{envNames.LetsEncryptEnabled},
			},
			&cli.StringFlag{
				Name:    flagNames.LetsEncryptCertDir,
				Usage:   "Directory to store acquired letsencrypt certificates.",
				Value:   defaults.LetsEncryptCertDir,
				EnvVars: []string{envNames.LetsEncryptCertDir},
			},
			&cli.StringFlag{
				Name:    flagNames.LetsEncryptEmailAddress,
				Usage:   "Email address to use when requesting letsencrypt certs. Will receive updates on cert expiry etc.",
				Value:   defaults.LetsEncryptEmailAddress,
				EnvVars: []string{envNames.LetsEncryptEmailAddress},
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
							return runAction(c, server.Start)
						},
					},
				},
			},
			{
				Name:  "admin",
				Usage: "gotosocial admin-related tasks",
				Subcommands: []*cli.Command{
					{
						Name:  "account",
						Usage: "admin commands related to accounts",
						Subcommands: []*cli.Command{
							{
								Name:  "create",
								Usage: "create a new account",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:  config.UsernameFlag,
										Usage: config.UsernameUsage,
									},
									&cli.StringFlag{
										Name:  config.EmailFlag,
										Usage: config.EmailUsage,
									},
									&cli.StringFlag{
										Name:  config.PasswordFlag,
										Usage: config.PasswordUsage,
									},
								},
								Action: func(c *cli.Context) error {
									return runAction(c, account.Create)
								},
							},
							{
								Name:  "confirm",
								Usage: "confirm an existing account manually, thereby skipping email confirmation",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:  config.UsernameFlag,
										Usage: config.UsernameUsage,
									},
								},
								Action: func(c *cli.Context) error {
									return runAction(c, account.Confirm)
								},
							},
							{
								Name:  "promote",
								Usage: "promote an account to admin",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:  config.UsernameFlag,
										Usage: config.UsernameUsage,
									},
								},
								Action: func(c *cli.Context) error {
									return runAction(c, account.Promote)
								},
							},
							{
								Name:  "demote",
								Usage: "demote an account from admin to normal user",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:  config.UsernameFlag,
										Usage: config.UsernameUsage,
									},
								},
								Action: func(c *cli.Context) error {
									return runAction(c, account.Demote)
								},
							},
							{
								Name:  "disable",
								Usage: "prevent an account from signing in or posting etc, but don't delete anything",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:  config.UsernameFlag,
										Usage: config.UsernameUsage,
									},
								},
								Action: func(c *cli.Context) error {
									return runAction(c, account.Disable)
								},
							},
							{
								Name:  "suspend",
								Usage: "completely remove an account and all of its posts, media, etc",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:  config.UsernameFlag,
										Usage: config.UsernameUsage,
									},
								},
								Action: func(c *cli.Context) error {
									return runAction(c, account.Suspend)
								},
							},
						},
					},
				},
			},
			{
				Name:  "testrig",
				Usage: "gotosocial testrig tasks",
				Subcommands: []*cli.Command{
					{
						Name:  "start",
						Usage: "start the gotosocial testrig",
						Action: func(c *cli.Context) error {
							return runAction(c, testrig.Start)
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
func runAction(c *cli.Context, a cliactions.GTSAction) error {

	// create a new *config.Config based on the config path provided...
	conf, err := config.FromFile(c.String(config.GetFlagNames().ConfigPath))
	if err != nil {
		return fmt.Errorf("error creating config: %s", err)
	}
	// ... and the flags set on the *cli.Context by urfave
	if err := conf.ParseCLIFlags(c, c.App.Version); err != nil {
		return fmt.Errorf("error parsing config: %s", err)
	}

	// create a logger with the log level, formatting, and output splitter already set
	log, err := log.New(conf.LogLevel)
	if err != nil {
		return fmt.Errorf("error creating logger: %s", err)
	}

	return a(c.Context, conf, log)
}

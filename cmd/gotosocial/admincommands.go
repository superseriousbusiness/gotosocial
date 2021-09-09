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
	"github.com/superseriousbusiness/gotosocial/internal/cliactions/admin/account"
	"github.com/superseriousbusiness/gotosocial/internal/cliactions/admin/trans"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/urfave/cli/v2"
)

func adminCommands() []*cli.Command {
	return []*cli.Command{
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
									Name:     config.UsernameFlag,
									Usage:    config.UsernameUsage,
									Required: true,
								},
								&cli.StringFlag{
									Name:     config.EmailFlag,
									Usage:    config.EmailUsage,
									Required: true,
								},
								&cli.StringFlag{
									Name:     config.PasswordFlag,
									Usage:    config.PasswordUsage,
									Required: true,
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
									Name:     config.UsernameFlag,
									Usage:    config.UsernameUsage,
									Required: true,
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
									Name:     config.UsernameFlag,
									Usage:    config.UsernameUsage,
									Required: true,
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
									Name:     config.UsernameFlag,
									Usage:    config.UsernameUsage,
									Required: true,
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
									Name:     config.UsernameFlag,
									Usage:    config.UsernameUsage,
									Required: true,
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
									Name:     config.UsernameFlag,
									Usage:    config.UsernameUsage,
									Required: true,
								},
							},
							Action: func(c *cli.Context) error {
								return runAction(c, account.Suspend)
							},
						},
						{
							Name:  "password",
							Usage: "set a new password for the given account",
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:     config.UsernameFlag,
									Usage:    config.UsernameUsage,
									Required: true,
								},
								&cli.StringFlag{
									Name:     config.PasswordFlag,
									Usage:    config.PasswordUsage,
									Required: true,
								},
							},
							Action: func(c *cli.Context) error {
								return runAction(c, account.Password)
							},
						},
					},
				},
				{
					Name:  "export",
					Usage: "export data from the database to file at the given path",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     config.TransPathFlag,
							Usage:    config.TransPathUsage,
							Required: true,
						},
					},
					Action: func(c *cli.Context) error {
						return runAction(c, trans.Export)
					},
				},
				{
					Name:  "import",
					Usage: "import data from a file into the database",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     config.TransPathFlag,
							Usage:    config.TransPathUsage,
							Required: true,
						},
					},
					Action: func(c *cli.Context) error {
						return runAction(c, trans.Import)
					},
				},
			},
		},
	}
}

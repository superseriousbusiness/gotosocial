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

func smtpFlags(flagNames, envNames config.Flags, defaults config.Defaults) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    flagNames.SMTPHost,
			Usage:   "Host of the smtp server. Eg., 'smtp.eu.mailgun.org'",
			Value:   defaults.SMTPHost,
			EnvVars: []string{envNames.SMTPHost},
		},
		&cli.IntFlag{
			Name:    flagNames.SMTPPort,
			Usage:   "Port of the smtp server. Eg., 587",
			Value:   defaults.SMTPPort,
			EnvVars: []string{envNames.SMTPPort},
		},
		&cli.StringFlag{
			Name:    flagNames.SMTPUsername,
			Usage:   "Username to authenticate with the smtp server as. Eg., 'postmaster@mail.example.org'",
			Value:   defaults.SMTPUsername,
			EnvVars: []string{envNames.SMTPUsername},
		},
		&cli.StringFlag{
			Name:    flagNames.SMTPPassword,
			Usage:   "Password to pass to the smtp server.",
			Value:   defaults.SMTPPassword,
			EnvVars: []string{envNames.SMTPPassword},
		},
		&cli.StringFlag{
			Name:    flagNames.SMTPFrom,
			Usage:   "Address to use as the 'from' field of the email. Eg., 'gotosocial@example.org'",
			Value:   defaults.SMTPFrom,
			EnvVars: []string{envNames.SMTPFrom},
		},
	}
}

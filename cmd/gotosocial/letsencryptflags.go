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

func letsEncryptFlags(flagNames, envNames config.Flags, defaults config.Defaults) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    flagNames.LetsEncryptEnabled,
			Usage:   "Enable letsencrypt TLS certs for this server. If set to true, then cert dir also needs to be set (or take the default).",
			Value:   defaults.LetsEncryptEnabled,
			EnvVars: []string{envNames.LetsEncryptEnabled},
		},
		&cli.IntFlag{
			Name:    flagNames.LetsEncryptPort,
			Usage:   "Port to listen on for letsencrypt certificate challenges. Must not be the same as the GtS webserver/API port.",
			Value:   defaults.LetsEncryptPort,
			EnvVars: []string{envNames.LetsEncryptPort},
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
	}
}

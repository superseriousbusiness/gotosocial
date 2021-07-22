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

func oidcFlags(flagNames, envNames config.Flags, defaults config.Defaults) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    flagNames.OIDCEnabled,
			Usage:   "Enabled OIDC authorization for this instance. If set to true, then the other OIDC flags must also be set.",
			Value:   defaults.OIDCEnabled,
			EnvVars: []string{envNames.OIDCEnabled},
		},
		&cli.StringFlag{
			Name:    flagNames.OIDCIdpName,
			Usage:   "Name of the OIDC identity provider. Will be shown to the user when logging in.",
			Value:   defaults.OIDCIdpName,
			EnvVars: []string{envNames.OIDCIdpName},
		},
		&cli.BoolFlag{
			Name:    flagNames.OIDCSkipVerification,
			Usage:   "Skip verification of tokens returned by the OIDC provider. Should only be set to 'true' for testing purposes, never in a production environment!",
			Value:   defaults.OIDCSkipVerification,
			EnvVars: []string{envNames.OIDCSkipVerification},
		},
		&cli.StringFlag{
			Name:    flagNames.OIDCIssuer,
			Usage:   "Address of the OIDC issuer. Should be the web address, including protocol, at which the issuer can be reached. Eg., 'https://example.org/auth'",
			Value:   defaults.OIDCIssuer,
			EnvVars: []string{envNames.OIDCIssuer},
		},
		&cli.StringFlag{
			Name:    flagNames.OIDCClientID,
			Usage:   "ClientID of GoToSocial, as registered with the OIDC provider.",
			Value:   defaults.OIDCClientID,
			EnvVars: []string{envNames.OIDCClientID},
		},
		&cli.StringFlag{
			Name:    flagNames.OIDCClientSecret,
			Usage:   "ClientSecret of GoToSocial, as registered with the OIDC provider.",
			Value:   defaults.OIDCClientSecret,
			EnvVars: []string{envNames.OIDCClientSecret},
		},
		&cli.StringSliceFlag{
			Name:    flagNames.OIDCScopes,
			Usage:   "ClientSecret of GoToSocial, as registered with the OIDC provider.",
			Value:   cli.NewStringSlice(defaults.OIDCScopes...),
			EnvVars: []string{envNames.OIDCScopes},
		},
	}
}

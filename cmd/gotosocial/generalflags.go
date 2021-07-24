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

func generalFlags(flagNames, envNames config.Flags, defaults config.Defaults) []cli.Flag {
	return []cli.Flag{
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
			Usage:   "Hostname to use for the server (eg., example.org, gotosocial.whatever.com). DO NOT change this on a server that's already run!",
			Value:   defaults.Host,
			EnvVars: []string{envNames.Host},
		},
		&cli.StringFlag{
			Name:    flagNames.AccountDomain,
			Usage:   "Domain to use in account names (eg., example.org, whatever.com). If not set, will default to the setting for host. DO NOT change this on a server that's already run!",
			Value:   defaults.AccountDomain,
			EnvVars: []string{envNames.AccountDomain},
		},
		&cli.StringFlag{
			Name:    flagNames.Protocol,
			Usage:   "Protocol to use for the REST api of the server (only use http for debugging and tests!)",
			Value:   defaults.Protocol,
			EnvVars: []string{envNames.Protocol},
		},
		&cli.IntFlag{
			Name:    flagNames.Port,
			Usage:   "Port to use for GoToSocial. Change this to 443 if you're running the binary directly on the host machine.",
			Value:   defaults.Port,
			EnvVars: []string{envNames.Port},
		},
		&cli.StringSliceFlag{
			Name:    flagNames.TrustedProxies,
			Usage:   "Proxies to trust when parsing x-forwarded headers into real IPs.",
			Value:   cli.NewStringSlice(defaults.TrustedProxies...),
			EnvVars: []string{envNames.TrustedProxies},
		},
	}
}

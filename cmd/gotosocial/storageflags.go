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

func storageFlags(flagNames, envNames config.Flags, defaults config.Defaults) []cli.Flag {
	return []cli.Flag{
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
	}
}

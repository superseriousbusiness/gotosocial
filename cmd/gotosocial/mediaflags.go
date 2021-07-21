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

func mediaFlags(flagNames, envNames config.Flags, defaults config.Defaults) []cli.Flag {
	return []cli.Flag{
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
	}
}

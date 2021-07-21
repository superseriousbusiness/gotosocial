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

func statusesFlags(flagNames, envNames config.Flags, defaults config.Defaults) []cli.Flag {
	return []cli.Flag{
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
	}
}

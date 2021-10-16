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

func getFlags() []cli.Flag {
	flagNames := config.GetFlagNames()
	envNames := config.GetEnvNames()
	defaults := config.GetDefaults()

	flags := []cli.Flag{}
	flagSets := [][]cli.Flag{
		generalFlags(flagNames, envNames, defaults),
		databaseFlags(flagNames, envNames, defaults),
		templateFlags(flagNames, envNames, defaults),
		accountsFlags(flagNames, envNames, defaults),
		mediaFlags(flagNames, envNames, defaults),
		storageFlags(flagNames, envNames, defaults),
		statusesFlags(flagNames, envNames, defaults),
		letsEncryptFlags(flagNames, envNames, defaults),
		oidcFlags(flagNames, envNames, defaults),
		smtpFlags(flagNames, envNames, defaults),
	}
	for _, fs := range flagSets {
		flags = append(flags, fs...)
	}

	return flags
}

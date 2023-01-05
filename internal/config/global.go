/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package config

import (
	"github.com/spf13/cobra"
)

var global *ConfigState

func init() {
	// init global state
	global = NewState()
}

// TODO: in the future we should move away from using globals in this config
// package, and instead pass the ConfigState round in a global gts state.

// Config provides you safe access to the global configuration.
func Config(fn func(cfg *Configuration)) {
	global.Config(fn)
}

// Reload will reload the current configuration values from file.
func Reload() error {
	return global.Reload()
}

// LoadEarlyFlags will bind specific flags from given Cobra command to global viper
// instance, and load the current configuration values. This is useful for flags like
// .ConfigPath which have to parsed first in order to perform early configuration load.
func LoadEarlyFlags(cmd *cobra.Command) error {
	return global.LoadEarlyFlags(cmd)
}

// BindFlags binds given command's pflags to the global viper instance.
func BindFlags(cmd *cobra.Command) error {
	return global.BindFlags(cmd)
}

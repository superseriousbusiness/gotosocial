// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
func Config(fn func(cfg *Configuration)) { global.Config(fn) }

// RegisterGlobalFlags ...
func RegisterGlobalFlags(root *cobra.Command) { global.RegisterGlobalFlags(root) }

// BindFlags binds given command's pflags to the global viper instance.
func BindFlags(cmd *cobra.Command) error { return global.BindFlags(cmd) }

// LoadConfigFile loads the currently set configuration file into the global viper instance.
func LoadConfigFile() error { return global.LoadConfigFile() }

// Reset will totally clear global
// ConfigState{}, loading defaults.
func Reset() { global.Reset() }

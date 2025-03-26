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
	"strings"
	"sync"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ConfigState manages safe concurrent access to Configuration{} values,
// and provides ease of linking them (including reloading) via viper to
// environment, CLI and configuration file variables.
type ConfigState struct { //nolint
	viper  *viper.Viper
	config Configuration
	mutex  sync.RWMutex
}

// NewState returns a new initialized ConfigState instance.
func NewState() *ConfigState {
	st := new(ConfigState)
	st.Reset()
	return st
}

// Config provides safe access to the ConfigState's contained Configuration,
// and will reload the current Configuration back into viper settings.
func (st *ConfigState) Config(fn func(*Configuration)) {
	st.mutex.Lock()
	defer func() {
		st.reloadToViper()
		st.mutex.Unlock()
	}()
	fn(&st.config)
}

// Viper provides safe access to the ConfigState's contained viper instance,
// and will reload the current viper setting state back into Configuration.
func (st *ConfigState) Viper(fn func(*viper.Viper)) {
	st.mutex.Lock()
	defer func() {
		st.reloadFromViper()
		st.mutex.Unlock()
	}()
	fn(st.viper)
}

// LoadEarlyFlags will bind specific flags from given Cobra command to ConfigState's viper
// instance, and load the current configuration values. This is useful for flags like
// .ConfigPath which have to parsed first in order to perform early configuration load.
func (st *ConfigState) LoadEarlyFlags(cmd *cobra.Command) (err error) {
	name := ConfigPathFlag()
	flag := cmd.Flags().Lookup(name)
	st.Viper(func(v *viper.Viper) {
		err = v.BindPFlag(name, flag)
	})
	return
}

// BindFlags will bind given Cobra command's pflags to this ConfigState's viper instance.
func (st *ConfigState) BindFlags(cmd *cobra.Command) (err error) {
	st.Viper(func(v *viper.Viper) {
		err = v.BindPFlags(cmd.Flags())
	})
	return
}

// Reload will reload the Configuration values from ConfigState's viper instance, and from file if set.
func (st *ConfigState) Reload() (err error) {
	st.Viper(func(v *viper.Viper) {
		if st.config.ConfigPath != "" {
			// Ensure configuration path is set
			v.SetConfigFile(st.config.ConfigPath)

			// Read in configuration from file
			if err = v.ReadInConfig(); err != nil {
				return
			}
		}
	})
	return
}

// Reset will totally clear
// ConfigState{}, loading defaults.
func (st *ConfigState) Reset() {
	// Do within lock.
	st.mutex.Lock()
	defer st.mutex.Unlock()

	// Create new viper.
	viper := viper.New()

	// Flag 'some-flag-name' becomes env var 'GTS_SOME_FLAG_NAME'
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("gts")

	// Load appropriate
	// named vals from env.
	viper.AutomaticEnv()

	// Reset variables.
	st.viper = viper
	st.config = Defaults

	// Load into viper.
	st.reloadToViper()
}

// reloadToViper will reload Configuration{} values into viper.
func (st *ConfigState) reloadToViper() {
	raw, err := st.config.MarshalMap()
	if err != nil {
		panic(err)
	}
	if err := st.viper.MergeConfigMap(raw); err != nil {
		panic(err)
	}
}

// reloadFromViper will reload Configuration{} values from viper.
func (st *ConfigState) reloadFromViper() {
	if err := st.viper.Unmarshal(&st.config, func(c *mapstructure.DecoderConfig) {
		c.TagName = "name"

		// empty config before marshaling
		c.ZeroFields = true

		oldhook := c.DecodeHook

		// Use the TextUnmarshaler interface when decoding.
		c.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.TextUnmarshallerHookFunc(),
			oldhook,
		)
	}); err != nil {
		panic(err)
	}
}

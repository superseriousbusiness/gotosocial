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
	"os"
	"path"
	"strings"
	"sync"

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
	defer st.mutex.Unlock()
	fn(&st.config)
	st.reloadToViper()
}

// Viper provides safe access to the ConfigState's contained viper instance,
// and will reload the current viper setting state back into Configuration.
func (st *ConfigState) Viper(fn func(*viper.Viper)) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	fn(st.viper)
	st.reloadFromViper()
}

// RegisterGlobalFlags ...
func (st *ConfigState) RegisterGlobalFlags(root *cobra.Command) {
	st.mutex.RLock()
	st.config.RegisterFlags(root.PersistentFlags())
	st.mutex.RUnlock()
}

// BindFlags will bind given Cobra command's pflags to this ConfigState's viper instance.
func (st *ConfigState) BindFlags(cmd *cobra.Command) (err error) {
	st.Viper(func(v *viper.Viper) {
		err = v.BindPFlags(cmd.Flags())
	})
	return
}

// LoadConfigFile loads the currently set configuration file into this ConfigState's viper instance.
func (st *ConfigState) LoadConfigFile() (err error) {
	st.Viper(func(v *viper.Viper) {
		if path := st.config.ConfigPath; path != "" {
			var cfgmap map[string]any

			// Read config map into memory.
			cfgmap, err := readConfigMap(path)
			if err != nil {
				return
			}

			// Merge the parsed config into viper.
			err = st.viper.MergeConfigMap(cfgmap)
			if err != nil {
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
	st.viper = viper.New()

	// Flag 'some-flag-name' becomes env var 'GTS_SOME_FLAG_NAME'
	st.viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	st.viper.SetEnvPrefix("gts")

	// Load appropriate
	// named vals from env.
	st.viper.AutomaticEnv()

	// Set default config.
	st.config = Defaults

	// Load into viper.
	st.reloadToViper()
}

// reloadToViper will reload Configuration{} values into viper.
func (st *ConfigState) reloadToViper() {
	if err := st.viper.MergeConfigMap(st.config.MarshalMap()); err != nil {
		panic(err)
	}
}

// reloadFromViper will reload Configuration{} values from viper.
func (st *ConfigState) reloadFromViper() {
	if err := st.config.UnmarshalMap(st.viper.AllSettings()); err != nil {
		panic(err)
	}
}

// readConfigMap reads given configuration file into memory,
// using viper's codec registry to handle decoding into a map,
// flattening the result for standardization, returning this.
// this ensures the stored config map in viper always has the
// same level of nesting, given we support varying levels.
func readConfigMap(file string) (map[string]any, error) {
	ext := path.Ext(file)
	ext = strings.TrimPrefix(ext, ".")

	registry := viper.NewCodecRegistry()
	dec, err := registry.Decoder(ext)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	cfgmap := make(map[string]any)

	if err := dec.Decode(data, cfgmap); err != nil {
		return nil, err
	}

	flattenConfigMap(cfgmap)

	return cfgmap, nil
}

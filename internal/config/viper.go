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

package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func InitViper(f *pflag.FlagSet, version string) error {
	// environment variable stuff
	// flag 'some-flag-name' becomes env var 'GTS_SOME_FLAG_NAME'
	viper.SetEnvPrefix("gts")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// register all aliases so that we can retrieve values set in the config file, using the standard keys
	for flagName, configKey := range Aliases {
		viper.RegisterAlias(flagName, configKey)
	}

	// bind this flag early so we know where to load the config file from
	viper.BindPFlag(FlagNames.ConfigPath, f.Lookup(FlagNames.ConfigPath))

	// config file stuff
	// check if we have a config path set (either by cli arg or env var)
	if configPath := viper.GetString(FlagNames.ConfigPath); configPath != "" {
		// we have a config path set; we need to juggle it so that viper can read it properly
		// see https://github.com/spf13/viper#reading-config-files
		dir, file := filepath.Split(configPath)        // return eg., /some/dir/ , config.yaml
		extension := filepath.Ext(file)                // return eg., .yaml
		fileName := strings.TrimRight(file, extension) // return eg., config

		viper.SetConfigName(fileName)                           // config
		viper.SetConfigType(strings.TrimPrefix(extension, ".")) // yaml
		viper.AddConfigPath(dir)                                // /some/dir/

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}

	// flag stuff
	// bind all of the flags in flag set to viper so that we can retrieve their values from the viper store
	if err := viper.BindPFlags(f); err != nil {
		return fmt.Errorf("error with viper: %s", err)
	}

	// override software version with whatever we've been passed
	viper.Set(FlagNames.SoftwareVersion, version)

	return nil
}

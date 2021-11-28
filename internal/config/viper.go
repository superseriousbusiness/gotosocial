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

	// flag stuff
	viper.BindPFlags(f)

	// override software version with whatever we've been passed
	viper.Set(FlagNames.SoftwareVersion, version)

	// config file stuff
	// first check if we have a config path set on the flag
	configPath, err := f.GetString(FlagNames.ConfigPath)
	if err != nil {
		return err
	}

	// no config path so nothing left to do here
	if configPath == "" {
		return nil
	}

	// we have a config path set; we need to juggle it so that viper can read it properly
	// see https://github.com/spf13/viper#reading-config-files
	dir, file := filepath.Split(configPath)        // return eg., /some/dir/ , config.yaml
	extension := filepath.Ext(file)                // return eg., yaml
	fileName := strings.TrimRight(file, extension) // return eg., config

	viper.SetConfigName(fileName)  // config
	viper.SetConfigType(extension) // yaml
	viper.AddConfigPath(dir)       // /some/dir/

	return viper.ReadInConfig()
}

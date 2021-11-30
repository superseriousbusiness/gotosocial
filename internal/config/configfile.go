package config

import (
	"strings"

	"github.com/spf13/viper"
)

func InitConfig() {
	// environment variable stuff
	// flag 'some-flag-name' becomes env var 'GTS_SOME_FLAG_NAME'
	viper.SetEnvPrefix("gts")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// config file stuff
	// check if we have a config path set (either by cli arg or env var)
	if configPath := viper.GetString(FlagNames.ConfigPath); configPath != "" {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}
	}
}

package config

import (
	"github.com/spf13/viper"
)

func InitConfig() error {
	// config file stuff
	// check if we have a config path set (either by cli arg or env var)
	if configPath := viper.GetString(FlagNames.ConfigPath); configPath != "" {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}

	return nil
}

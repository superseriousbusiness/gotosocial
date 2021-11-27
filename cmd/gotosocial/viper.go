package main

import (
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func initViper(flags *pflag.FlagSet) error {
	// flag stuff
	viper.BindPFlags(flags)
	
	// environment variable stuff
	viper.SetEnvPrefix("gts")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// config file stuff
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	return viper.ReadInConfig()
}

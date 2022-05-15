package config

import (
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var (
	global Configuration
	mutex  sync.Mutex
	gviper viper.Viper
)

func init() {
	// Flag 'some-flag-name' becomes env var 'GTS_SOME_FLAG_NAME'
	gviper.SetEnvPrefix("gts")
	gviper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Load appropriate named vals from env
	gviper.AutomaticEnv()

	// Set starting defaults
	global = Defaults
}

// Config provides you safe access to the global configuration.
func Config(fn func(cfg *Configuration)) {
	mutex.Lock()
	defer mutex.Unlock()
	fn(&global)
}

// Reload will reload the current configuration values from file.
func Reload() (err error) {
	Config(func(cfg *Configuration) {
		// Ensure configuration path is set
		gviper.SetConfigFile(cfg.ConfigPath)

		// Read in configuration from file
		if err = viper.ReadInConfig(); err != nil {
			return
		}

		// Unmarshal configuration to Configuration struct
		if err = viper.UnmarshalExact(cfg); err != nil {
			return
		}
	})
	return
}

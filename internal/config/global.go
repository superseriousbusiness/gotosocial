package config

import (
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var (
	global Configuration
	mutex  sync.Mutex
	gviper *viper.Viper
)

func init() {
	// init global viper
	gviper = viper.New()

	// Flag 'some-flag-name' becomes env var 'GTS_SOME_FLAG_NAME'
	gviper.SetEnvPrefix("gts")
	gviper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Load appropriate named vals from env
	gviper.AutomaticEnv()

	// Set starting defaults
	global = Defaults
}

// TODO: in the future we should move away from using globals in this config
// package, and instead wrap the functionality in Configuration{} and pass this
// round as a state to all of the gotosocial subsystems. The generator will still
// come in handy for that in generating getters/setters :)

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
		if err = gviper.ReadInConfig(); err != nil {
			return
		}

		// Unmarshal configuration to Configuration struct
		if err = cfg.unmarshal(gviper.AllSettings()); err != nil {
			return
		}
	})
	return
}

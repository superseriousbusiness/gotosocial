package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func Validate() error {
	developmentMode := viper.GetBool(Keys.EnableDevelopmentSettings)
	protocol := viper.GetString(Keys.Protocol)
	if developmentMode {
		if strings.ToLower(protocol) != "https" {
			return fmt.Errorf("protocol is set to %s. This isn't allowed unless EnableDevelopmentSettings is true", protocol)
		}
	}
	if viper.GetBool(Keys.LetsEncryptEnabled) {
		if strings.ToLower(protocol) != "https" {
			return fmt.Errorf("protocol is set to %s. This isn't allowed when LetsEncryptEnabled is true", protocol)
		}
	}
	return nil
}

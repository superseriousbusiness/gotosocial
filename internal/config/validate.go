/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Validate validates global config settings which don't have defaults, to make sure they are set sensibly.
func Validate() error {
	errs := []error{}

	// host
	if viper.GetString(Keys.Host) == "" {
		errs = append(errs, fmt.Errorf("%s must be set", Keys.Host))
	}

	// protocol
	protocol := viper.GetString(Keys.Protocol)
	switch protocol {
	case "https":
		// no problem
		break
	case "http":
		logrus.Warnf("%s was set to 'http'; this should *only* be used for debugging and tests!", Keys.Protocol)
	case "":
		errs = append(errs, fmt.Errorf("%s must be set", Keys.Protocol))
	default:
		errs = append(errs, fmt.Errorf("%s must be set to either http or https, provided value was %s", Keys.Protocol, protocol))
	}

	if len(errs) > 0 {
		errStrings := []string{}
		for _, err := range errs {
			errStrings = append(errStrings, err.Error())
		}
		return errors.New(strings.Join(errStrings, "; "))
	}

	return nil
}

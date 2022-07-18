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

package log

import (
	"fmt"
	"log/syslog"
	"strings"

	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// Initialize initializes the global Logrus logger, reading the desired
// log level from the viper store, or using a default if the level
// has not been set in viper.
//
// It also sets the output to log.SplitErrOutputs(...)
// so you get error logs on stderr and normal logs on stdout.
//
// If syslog settings are also in viper, then Syslog will be initialized as well.
func Initialize() error {
	// check if a desired log level has been set
	if lvlStr := config.GetLogLevel(); lvlStr != "" {
		switch strings.ToLower(lvlStr) {
		case "trace":
			SetLevel(level.TRACE)
		case "debug":
			SetLevel(level.DEBUG)
		case "", "info":
			SetLevel(level.INFO)
		case "warn":
			SetLevel(level.WARN)
		case "error":
			SetLevel(level.ERROR)
		case "fatal":
			SetLevel(level.FATAL)
		default:
			return fmt.Errorf("unknown log level: %q", lvlStr)
		}
	}

	// check if syslog has been enabled, and configure it if so
	if config.GetSyslogEnabled() {
		protocol := config.GetSyslogProtocol()
		address := config.GetSyslogAddress()

		// Dial a connection to the syslog daemon
		writer, err := syslog.Dial(protocol, address, 0, "gotosocial")
		if err != nil {
			return err
		}

		// Set the syslog writer
		sysout = writer
	}

	return nil
}

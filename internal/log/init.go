/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
)

// ParseLevel will parse the log level from given string and set to appropriate level.
func ParseLevel(str string) error {
	switch strings.ToLower(str) {
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
		return fmt.Errorf("unknown log level: %q", str)
	}
	return nil
}

// EnableSyslog will enabling logging to the syslog at given address.
func EnableSyslog(proto, addr string) error {
	// Dial a connection to the syslog daemon
	writer, err := syslog.Dial(proto, addr, 0, "gotosocial")
	if err != nil {
		return err
	}

	// Set the syslog writer
	sysout = writer

	return nil
}

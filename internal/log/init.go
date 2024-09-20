// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package log

import (
	"fmt"
	"log/syslog"
	"strings"
)

// ParseLevel will parse the log level from given string and set to appropriate LEVEL.
func ParseLevel(str string) error {
	switch strings.ToLower(str) {
	case "trace":
		SetLevel(TRACE)
	case "debug":
		SetLevel(DEBUG)
	case "", "info":
		SetLevel(INFO)
	case "warn":
		SetLevel(WARN)
	case "error":
		SetLevel(ERROR)
	case "fatal", "panic":
		SetLevel(PANIC)
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

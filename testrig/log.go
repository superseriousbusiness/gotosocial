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

package testrig

import (
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

// InitTestLog sets the global logger to trace level for logging
func InitTestLog() {
	log.SetTimeFormat(config.GetLogTimestampFormat())
	// Set the global log level from configuration
	if err := log.ParseLevel(config.GetLogLevel()); err != nil {
		log.Panicf(nil, "error parsing log level: %v", err)
	}

	if config.GetSyslogEnabled() {
		// Enable logging to syslog
		if err := log.EnableSyslog(
			config.GetSyslogProtocol(),
			config.GetSyslogAddress(),
		); err != nil {
			log.Panicf(nil, "error enabling syslogging: %v", err)
		}
	}
}

// InitTestSyslog returns a test syslog running on port 42069 and a channel for reading
// messages sent to the server, or an error if something goes wrong.
//
// Callers of this function should call Kill() on the server when they're finished with it!
func InitTestSyslog() (*syslog.Server, chan format.LogParts, error) {
	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(handler)

	if err := server.ListenUDP("127.0.0.1:42069"); err != nil {
		return nil, nil, err
	}

	if err := server.Boot(); err != nil {
		return nil, nil, err
	}

	return server, channel, nil
}

// InitTestSyslog returns a test syslog running on a unix socket, and a channel for reading
// messages sent to the server, or an error if something goes wrong.
//
// Callers of this function should call Kill() on the server when they're finished with it!
func InitTestSyslogUnixgram(address string) (*syslog.Server, chan format.LogParts, error) {
	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(handler)

	if err := server.ListenUnixgram(address); err != nil {
		return nil, nil, err
	}

	if err := server.Boot(); err != nil {
		return nil, nil, err
	}

	return server, channel, nil
}

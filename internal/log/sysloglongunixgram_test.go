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

//go:build !darwin

package log_test

import (
	"fmt"
	"os"
	"path"
	"regexp"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/google/uuid"
)

// TestSyslogLongMessageUnixgram is known to hang on macOS for messages longer than about 1500 bytes.
func (suite *SyslogTestSuite) TestSyslogLongMessageUnixgram() {
	socketPath := path.Join(os.TempDir(), uuid.NewString())
	defer func() {
		if err := os.Remove(socketPath); err != nil {
			panic(err)
		}
	}()

	server, channel, err := testrig.InitTestSyslogUnixgram(socketPath)
	if err != nil {
		panic(err)
	}
	syslogServer := server
	syslogChannel := channel

	config.SetSyslogEnabled(true)
	config.SetSyslogProtocol("unixgram")
	config.SetSyslogAddress(socketPath)

	testrig.InitTestLog()

	log.Error(nil, longMessage)

	funcName := log.Caller(2)
	prefix := fmt.Sprintf(`timestamp="02/01/2006 15:04:05.000" func=%s level=ERROR msg="`, funcName)

	entry := <-syslogChannel
	regex := fmt.Sprintf(`timestamp=.* func=.* level=ERROR msg="%s`, longMessage[:2048-len(prefix)])

	suite.Regexp(regexp.MustCompile(regex), entry["content"])

	if err := syslogServer.Kill(); err != nil {
		panic(err)
	}
}

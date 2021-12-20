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

package log_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/testrig"
	"gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

type SyslogTestSuite struct {
	suite.Suite
	syslogServer  *syslog.Server
	syslogChannel chan format.LogParts
}

func (suite *SyslogTestSuite) SetupTest() {
	testrig.InitTestConfig()

	viper.Set(config.Keys.SyslogEnabled, true)
	viper.Set(config.Keys.SyslogProtocol, "udp")
	viper.Set(config.Keys.SyslogAddress, "localhost:42069")
	server, channel, err := testrig.InitTestSyslog()
	if err != nil {
		panic(err)
	}
	suite.syslogServer = server
	suite.syslogChannel = channel

	testrig.InitTestLog()
}

func (suite *SyslogTestSuite) TearDownTest() {
	if err := suite.syslogServer.Kill(); err != nil {
		panic(err)
	}
}

func (suite *SyslogTestSuite) TestSyslog() {
	logrus.Warn("this is a test of the emergency broadcast system!")

	message := <-suite.syslogChannel
	suite.Contains(message["content"], "this is a test of the emergency broadcast system!")
}

func TestSyslogTestSuite(t *testing.T) {
	suite.Run(t, &SyslogTestSuite{})
}

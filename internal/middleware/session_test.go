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

package middleware_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type SessionTestSuite struct {
	suite.Suite
}

func (suite *SessionTestSuite) SetupTest() {
	testrig.InitTestConfig()
}

func (suite *SessionTestSuite) TestDeriveSessionNameLocalhostWithPort() {
	config.SetProtocol("http")
	config.SetHost("localhost:8080")

	sessionName, err := middleware.SessionName()
	suite.NoError(err)
	suite.Equal("gotosocial-localhost", sessionName)
}

func (suite *SessionTestSuite) TestDeriveSessionNameLocalhost() {
	config.SetProtocol("http")
	config.SetHost("localhost")

	sessionName, err := middleware.SessionName()
	suite.NoError(err)
	suite.Equal("gotosocial-localhost", sessionName)
}

func (suite *SessionTestSuite) TestDeriveSessionNoProtocol() {
	config.SetProtocol("")
	config.SetHost("localhost")

	sessionName, err := middleware.SessionName()
	suite.EqualError(err, "parse \"://localhost\": missing protocol scheme")
	suite.Equal("", sessionName)
}

func (suite *SessionTestSuite) TestDeriveSessionNoHost() {
	config.SetProtocol("https")
	config.SetHost("")
	config.SetPort(0)

	sessionName, err := middleware.SessionName()
	suite.EqualError(err, "could not derive hostname without port from https://")
	suite.Equal("", sessionName)
}

func (suite *SessionTestSuite) TestDeriveSessionOK() {
	config.SetProtocol("https")
	config.SetHost("example.org")

	sessionName, err := middleware.SessionName()
	suite.NoError(err)
	suite.Equal("gotosocial-example.org", sessionName)
}

func (suite *SessionTestSuite) TestDeriveSessionIDNOK() {
	config.SetProtocol("https")
	config.SetHost("fóid.org")

	sessionName, err := middleware.SessionName()
	suite.NoError(err)
	suite.Equal("gotosocial-xn--fid-gna.org", sessionName)
}

func TestSessionTestSuite(t *testing.T) {
	suite.Run(t, &SessionTestSuite{})
}

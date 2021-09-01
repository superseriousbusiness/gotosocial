/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package router

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

type SessionTestSuite struct {
	suite.Suite
}

func (suite *SessionTestSuite) TestDeriveSessionNameLocalhostWithPort() {
	cfg := &config.Config{
		Protocol: "http",
		Host:     "localhost:8080",
	}

	sessionName, err := sessionName(cfg)
	suite.NoError(err)
	suite.Equal("gotosocial-localhost", sessionName)
}

func (suite *SessionTestSuite) TestDeriveSessionNameLocalhost() {
	cfg := &config.Config{
		Protocol: "http",
		Host:     "localhost",
	}

	sessionName, err := sessionName(cfg)
	suite.NoError(err)
	suite.Equal("gotosocial-localhost", sessionName)
}

func (suite *SessionTestSuite) TestDeriveSessionNoProtocol() {
	cfg := &config.Config{
		Host: "localhost",
	}

	sessionName, err := sessionName(cfg)
	suite.EqualError(err, "parse \"://localhost\": missing protocol scheme")
	suite.Equal("", sessionName)
}

func (suite *SessionTestSuite) TestDeriveSessionNoHost() {
	cfg := &config.Config{
		Protocol: "https",
	}

	sessionName, err := sessionName(cfg)
	suite.EqualError(err, "could not derive hostname without port from https://")
	suite.Equal("", sessionName)
}

func (suite *SessionTestSuite) TestDeriveSessionOK() {
	cfg := &config.Config{
		Protocol: "https",
		Host:     "example.org",
	}

	sessionName, err := sessionName(cfg)
	suite.NoError(err)
	suite.Equal("gotosocial-example.org", sessionName)
}

func TestSessionTestSuite(t *testing.T) {
	suite.Run(t, &SessionTestSuite{})
}

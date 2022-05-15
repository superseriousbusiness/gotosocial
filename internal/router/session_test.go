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

package router_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type SessionTestSuite struct {
	suite.Suite
}

func (suite *SessionTestSuite) SetupTest() {
	testrig.InitTestConfig()
}

func (suite *SessionTestSuite) TestDeriveSessionNameLocalhostWithPort() {
	config.Config(func(cfg *config.Configuration) {
		cfg.Protocol = "http"
		cfg.Host = "localhost:8080"
	})

	sessionName, err := router.SessionName()
	suite.NoError(err)
	suite.Equal("gotosocial-localhost", sessionName)
}

func (suite *SessionTestSuite) TestDeriveSessionNameLocalhost() {
	config.Config(func(cfg *config.Configuration) {
		cfg.Protocol = "http"
		cfg.Host = "localhost"
	})

	sessionName, err := router.SessionName()
	suite.NoError(err)
	suite.Equal("gotosocial-localhost", sessionName)
}

func (suite *SessionTestSuite) TestDeriveSessionNoProtocol() {
	config.Config(func(cfg *config.Configuration) {
		cfg.Protocol = ""
		cfg.Host = "localhost"
	})

	sessionName, err := router.SessionName()
	suite.EqualError(err, "parse \"://localhost\": missing protocol scheme")
	suite.Equal("", sessionName)
}

func (suite *SessionTestSuite) TestDeriveSessionNoHost() {
	config.Config(func(cfg *config.Configuration) {
		cfg.Protocol = "https"
		cfg.Host = ""
		cfg.Port = 0
	})

	sessionName, err := router.SessionName()
	suite.EqualError(err, "could not derive hostname without port from https://")
	suite.Equal("", sessionName)
}

func (suite *SessionTestSuite) TestDeriveSessionOK() {
	config.Config(func(cfg *config.Configuration) {
		cfg.Protocol = "https"
		cfg.Host = "example.org"
	})

	sessionName, err := router.SessionName()
	suite.NoError(err)
	suite.Equal("gotosocial-example.org", sessionName)
}

func (suite *SessionTestSuite) TestDeriveSessionIDNOK() {
	config.Config(func(cfg *config.Configuration) {
		cfg.Protocol = "https"
		cfg.Host = "fóid.org"
	})

	sessionName, err := router.SessionName()
	suite.NoError(err)
	suite.Equal("gotosocial-xn--fid-gna.org", sessionName)
}

func TestSessionTestSuite(t *testing.T) {
	suite.Run(t, &SessionTestSuite{})
}

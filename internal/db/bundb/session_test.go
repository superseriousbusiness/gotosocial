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

package bundb_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SessionTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *SessionTestSuite) TestGetSession() {
	session, err := suite.db.GetSession(suite.T().Context())
	suite.NoError(err)
	suite.NotNil(session)
	suite.NotEmpty(session.Auth)
	suite.NotEmpty(session.Crypt)
	suite.NotEmpty(session.ID)

	// the same session should be returned with consecutive selects
	session2, err := suite.db.GetSession(suite.T().Context())
	suite.NoError(err)
	suite.NotNil(session2)
	suite.Equal(session.Auth, session2.Auth)
	suite.Equal(session.Crypt, session2.Crypt)
	suite.Equal(session.ID, session2.ID)
}

func TestSessionTestSuite(t *testing.T) {
	suite.Run(t, new(SessionTestSuite))
}

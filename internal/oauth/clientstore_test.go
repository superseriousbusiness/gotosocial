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

package oauth_test

import (
	"context"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/admin"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type ClientStoreTestSuite struct {
	suite.Suite
	db               db.DB
	state            state.State
	testApplications map[string]*gtsmodel.Application
}

func (suite *ClientStoreTestSuite) SetupSuite() {
	suite.testApplications = testrig.NewTestApplications()
}

func (suite *ClientStoreTestSuite) SetupTest() {
	suite.state.Caches.Init()
	testrig.InitTestConfig()
	testrig.InitTestLog()
	suite.db = testrig.NewTestDB(&suite.state)
	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)
	testrig.StandardDBSetup(suite.db, nil)
}

func (suite *ClientStoreTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *ClientStoreTestSuite) TestClientStoreGet() {
	testApp := suite.testApplications["application_1"]
	cs := oauth.NewClientStore(&suite.state)

	// Fetch clientInfo from the store.
	clientInfo, err := cs.GetByID(context.Background(), testApp.ClientID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Check expected values.
	suite.NotNil(clientInfo)
	suite.Equal(testApp.ClientID, clientInfo.GetID())
	suite.Equal(testApp.ClientSecret, clientInfo.GetSecret())
	suite.Equal(testApp.RedirectURIs[0], clientInfo.GetDomain())
	suite.Equal(testApp.ManagedByUserID, clientInfo.GetUserID())
}

func TestClientStoreTestSuite(t *testing.T) {
	suite.Run(t, new(ClientStoreTestSuite))
}

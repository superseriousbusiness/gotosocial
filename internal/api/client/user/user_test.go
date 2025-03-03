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

package user_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/admin"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/user"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type UserStandardTestSuite struct {
	suite.Suite
	db           db.DB
	tc           *typeutils.Converter
	mediaManager *media.Manager
	federator    *federation.Federator
	processor    *processing.Processor
	storage      *storage.Driver
	state        state.State

	testTokens       map[string]*gtsmodel.Token
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account

	userModule *user.Module
}

func (suite *UserStandardTestSuite) SetupTest() {
	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.testTokens = testrig.NewTestTokens()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()

	suite.db = testrig.NewTestDB(&suite.state)
	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)
	suite.storage = testrig.NewInMemoryStorage()
	suite.state.Storage = suite.storage

	suite.tc = typeutils.NewConverter(&suite.state)

	testrig.StartTimelines(
		&suite.state,
		visibility.NewFilter(&suite.state),
		suite.tc,
	)

	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)
	suite.federator = testrig.NewTestFederator(
		&suite.state,
		testrig.NewTestTransportController(
			&suite.state,
			testrig.NewMockHTTPClient(nil, "../../../../testrig/media"),
		),
		suite.mediaManager,
	)
	suite.processor = testrig.NewTestProcessor(
		&suite.state,
		suite.federator,
		testrig.NewEmailSender("../../../../web/template/", nil),
		testrig.NewNoopWebPushSender(),
		suite.mediaManager,
	)
	suite.userModule = user.New(suite.processor)
	testrig.StandardDBSetup(suite.db, suite.testAccounts)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")
}

func (suite *UserStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	testrig.StopWorkers(&suite.state)
}

func (suite *UserStandardTestSuite) POST(path string, formValues map[string][]string, handler gin.HandlerFunc) (*http.Response, int) {
	var (
		oauthToken = oauth.DBTokenToToken(suite.testTokens["local_account_1"])
		app        = suite.testApplications["application_1"]
		user       = suite.testUsers["local_account_1"]
		account    = suite.testAccounts["local_account_1"]
		target     = "http://localhost:8080" + path
	)

	// Prepare context.
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, app)
	ctx.Set(oauth.SessionAuthorizedToken, oauthToken)
	ctx.Set(oauth.SessionAuthorizedUser, user)
	ctx.Set(oauth.SessionAuthorizedAccount, account)

	// Prepare request.
	ctx.Request = httptest.NewRequest(http.MethodPost, target, nil)
	ctx.Request.Header.Set("accept", "application/json")
	ctx.Request.Form = url.Values(formValues)

	// Call the handler.
	handler(ctx)

	// Return response.
	return recorder.Result(), recorder.Code
}

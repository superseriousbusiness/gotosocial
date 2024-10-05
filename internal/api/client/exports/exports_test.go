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

package exports_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/exports"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ExportsTestSuite struct {
	// Suite interfaces
	suite.Suite
	state state.State

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account

	// module being tested
	exportsModule *exports.Module
}

func (suite *ExportsTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
}

func (suite *ExportsTestSuite) SetupTest() {
	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.state.DB = testrig.NewTestDB(&suite.state)
	suite.state.Storage = testrig.NewInMemoryStorage()

	testrig.StartTimelines(
		&suite.state,
		visibility.NewFilter(&suite.state),
		typeutils.NewConverter(&suite.state),
	)

	testrig.StandardDBSetup(suite.state.DB, nil)
	testrig.StandardStorageSetup(suite.state.Storage, "../../../../testrig/media")

	mediaManager := testrig.NewTestMediaManager(&suite.state)

	federator := testrig.NewTestFederator(
		&suite.state,
		testrig.NewTestTransportController(
			&suite.state,
			testrig.NewMockHTTPClient(nil, "../../../../testrig/media"),
		),
		mediaManager,
	)

	processor := testrig.NewTestProcessor(
		&suite.state,
		federator,
		testrig.NewEmailSender("../../../../web/template/", nil),
		mediaManager,
	)

	suite.exportsModule = exports.New(processor)
}

func (suite *ExportsTestSuite) TriggerHandler(
	handler gin.HandlerFunc,
	path string,
	contentType string,
	application *gtsmodel.Application,
	token *gtsmodel.Token,
	user *gtsmodel.User,
	account *gtsmodel.Account,
) *httptest.ResponseRecorder {
	// Set up request.
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)

	// Authorize the request ctx as though it
	// had passed through API auth handlers.
	ctx.Set(oauth.SessionAuthorizedApplication, application)
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(token))
	ctx.Set(oauth.SessionAuthorizedUser, user)
	ctx.Set(oauth.SessionAuthorizedAccount, account)

	// Create test request.
	target := "http://localhost:8080/api" + path
	ctx.Request = httptest.NewRequest(http.MethodGet, target, nil)
	ctx.Request.Header.Set("Accept", contentType)

	// Trigger handler.
	handler(ctx)

	return recorder
}

func (suite *ExportsTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.state.DB)
	testrig.StandardStorageTeardown(suite.state.Storage)
	testrig.StopWorkers(&suite.state)
}

func (suite *ExportsTestSuite) TestExports() {
	type testCase struct {
		handler     gin.HandlerFunc
		path        string
		contentType string
		application *gtsmodel.Application
		token       *gtsmodel.Token
		user        *gtsmodel.User
		account     *gtsmodel.Account
		expect      string
	}

	testCases := []testCase{
		// Export Following
		{
			handler:     suite.exportsModule.ExportFollowingGETHandler,
			path:        exports.FollowingPath,
			contentType: apiutil.TextCSV,
			application: suite.testApplications["application_1"],
			token:       suite.testTokens["local_account_1"],
			user:        suite.testUsers["local_account_1"],
			account:     suite.testAccounts["local_account_1"],
			expect: `Account address,Show boosts,Notify on new posts,Languages
1happyturtle@localhost:8080,true,false,
admin@localhost:8080,true,false,
`,
		},
		// Export Followers.
		{
			handler:     suite.exportsModule.ExportFollowersGETHandler,
			path:        exports.FollowingPath,
			contentType: apiutil.TextCSV,
			application: suite.testApplications["application_1"],
			token:       suite.testTokens["local_account_1"],
			user:        suite.testUsers["local_account_1"],
			account:     suite.testAccounts["local_account_1"],
			expect: `Account address
1happyturtle@localhost:8080
admin@localhost:8080
`,
		},
		// Export Lists.
		{
			handler:     suite.exportsModule.ExportListsGETHandler,
			path:        exports.ListsPath,
			contentType: apiutil.TextCSV,
			application: suite.testApplications["application_1"],
			token:       suite.testTokens["local_account_1"],
			user:        suite.testUsers["local_account_1"],
			account:     suite.testAccounts["local_account_1"],
			expect: `Cool Ass Posters From This Instance,1happyturtle@localhost:8080
Cool Ass Posters From This Instance,admin@localhost:8080
`,
		},
		// Export Mutes.
		{
			handler:     suite.exportsModule.ExportMutesGETHandler,
			path:        exports.MutesPath,
			contentType: apiutil.TextCSV,
			application: suite.testApplications["application_1"],
			token:       suite.testTokens["local_account_1"],
			user:        suite.testUsers["local_account_1"],
			account:     suite.testAccounts["local_account_1"],
			expect: `Account address,Hide notifications
`,
		},
		// Export Blocks.
		{
			handler:     suite.exportsModule.ExportBlocksGETHandler,
			path:        exports.BlocksPath,
			contentType: apiutil.TextCSV,
			application: suite.testApplications["application_1"],
			token:       suite.testTokens["local_account_2"],
			user:        suite.testUsers["local_account_2"],
			account:     suite.testAccounts["local_account_2"],
			expect: `foss_satan@fossbros-anonymous.io
`,
		},
		// Export Stats.
		{
			handler:     suite.exportsModule.ExportStatsGETHandler,
			path:        exports.StatsPath,
			contentType: apiutil.AppJSON,
			application: suite.testApplications["application_1"],
			token:       suite.testTokens["local_account_1"],
			user:        suite.testUsers["local_account_1"],
			account:     suite.testAccounts["local_account_1"],
			expect: `{
  "media_storage": "",
  "followers_count": 2,
  "following_count": 2,
  "statuses_count": 8,
  "lists_count": 1,
  "blocks_count": 0,
  "mutes_count": 0
}`,
		},
	}

	for _, test := range testCases {
		recorder := suite.TriggerHandler(
			test.handler,
			test.path,
			test.contentType,
			test.application,
			test.token,
			test.user,
			test.account,
		)

		// Check response code.
		suite.EqualValues(http.StatusOK, recorder.Code)

		// Check response body.
		b, err := io.ReadAll(recorder.Body)
		if err != nil {
			suite.FailNow(err.Error())
		}

		// If json response, indent it nicely.
		if recorder.Result().Header.Get("Content-Type") == "application/json" {
			dst := &bytes.Buffer{}
			if err := json.Indent(dst, b, "", "  "); err != nil {
				suite.FailNow(err.Error())
			}
			b = dst.Bytes()
		}

		suite.Equal(test.expect, string(b))
	}
}

func TestExportsTestSuite(t *testing.T) {
	suite.Run(t, new(ExportsTestSuite))
}

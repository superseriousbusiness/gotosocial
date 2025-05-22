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

package importdata_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	importdata "code.superseriousbusiness.org/gotosocial/internal/api/client/import"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/testrig"
)

type ImportTestSuite struct {
	// Suite interfaces
	suite.Suite
	state state.State

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account

	// module being tested
	importModule *importdata.Module
}

func (suite *ImportTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
}

func (suite *ImportTestSuite) SetupTest() {
	suite.state.Caches.Init()

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.state.DB = testrig.NewTestDB(&suite.state)
	suite.state.Storage = testrig.NewInMemoryStorage()

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
		testrig.NewNoopWebPushSender(),
		mediaManager,
	)
	testrig.StartWorkers(&suite.state, processor.Workers())

	suite.importModule = importdata.New(processor)
}

func (suite *ImportTestSuite) TriggerHandler(
	importData string,
	importType string,
	importMode string,
) {
	// Set up request.
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)

	// Authorize the request ctx as though it
	// had passed through API auth handlers.
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// Create test request.
	b, w, err := testrig.CreateMultipartFormData(
		testrig.StringToDataF("data", "data.csv", importData),
		map[string][]string{
			"type": {importType},
			"mode": {importMode},
		},
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	target := "http://localhost:8080/api/v1/import"
	ctx.Request = httptest.NewRequest(http.MethodPost, target, bytes.NewReader(b.Bytes()))
	ctx.Request.Header.Set("Accept", "application/json")
	ctx.Request.Header.Set("Content-Type", w.FormDataContentType())

	// Trigger handler.
	suite.importModule.ImportPOSTHandler(ctx)

	if code := recorder.Code; code != http.StatusAccepted {
		b, err := io.ReadAll(recorder.Body)
		if err != nil {
			panic(err)
		}
		suite.FailNow("", "expected 202, got %d: %s", code, string(b))
	}
}

func (suite *ImportTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.state.DB)
	testrig.StandardStorageTeardown(suite.state.Storage)
	testrig.StopWorkers(&suite.state)
}

func (suite *ImportTestSuite) TestImportFollows() {
	var (
		ctx         = suite.T().Context()
		testAccount = suite.testAccounts["local_account_1"]
	)

	// Clear existing follows from Zork.
	if err := suite.state.DB.DeleteAccountFollows(ctx, testAccount.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Have zork refollow turtle and admin.
	data := `Account address,Show boosts
admin@localhost:8080,true
1happyturtle@localhost:8080,true
`

	// Trigger the import handler.
	suite.TriggerHandler(data, "following", "merge")

	// Wait for zork to be
	// following admin.
	if !testrig.WaitFor(func() bool {
		f, err := suite.state.DB.IsFollowing(
			ctx,
			testAccount.ID,
			suite.testAccounts["admin_account"].ID,
		)
		if err != nil {
			suite.FailNow(err.Error())
		}

		return f
	}) {
		suite.FailNow("timed out waiting for zork to follow admin")
	}

	// Wait for zork to be
	// follow req'ing turtle.
	if !testrig.WaitFor(func() bool {
		f, err := suite.state.DB.IsFollowRequested(
			ctx,
			testAccount.ID,
			suite.testAccounts["local_account_2"].ID,
		)
		if err != nil {
			suite.FailNow(err.Error())
		}

		return f
	}) {
		suite.FailNow("timed out waiting for zork to follow req turtle")
	}
}

func (suite *ImportTestSuite) TestImportMutes() {
	var (
		ctx         = suite.T().Context()
		testAccount = suite.testAccounts["local_account_1"]
	)

	// Clear existing mutes from Zork.
	if err := suite.state.DB.DeleteAccountMutes(ctx, testAccount.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Have zork mute turtle, admin and remote fossbro.
	data := `Account address,Hide notifications
admin@localhost:8080,true
unknown@localhost:8080,true
1happyturtle@localhost:8080,false
foss_satan@fossbros-anonymous.io,true
`

	// Trigger the import handler.
	suite.TriggerHandler(data, "mutes", "merge")

	// Wait for mutes to be applied
	if !testrig.WaitFor(func() bool {
		mutes, err := suite.state.DB.GetAccountMutes(ctx, testAccount.ID, nil)
		if err != nil {
			suite.FailNow(err.Error())
		}
		for _, m := range mutes {
			switch m.TargetAccount.ID {
			case suite.testAccounts["remote_account_1"].ID:
				if *m.Notifications != true {
					suite.FailNow("expected notifications from fossbro to be muted")
				}
			case suite.testAccounts["admin_account"].ID:
				if *m.Notifications != true {
					suite.FailNow("expected notifications from admin to be muted")
				}
			case suite.testAccounts["local_account_2"].ID:
				if *m.Notifications != false {
					suite.FailNow("expected notifications from turtle to NOT be muted")
				}
			default:
				suite.FailNow("unexpected muted account", m.TargetAccount)
			}
		}
		return len(mutes) == 3
	}) {
		suite.FailNow("timed out waiting for mutes to apply")
	}

	// Import again in overwrite mode:
	//   - remote fossbro is unmuted, admin and turtle are kept
	//   - Notification hiding is reversed to confirm mutes are modified
	data = `Account address,Hide notifications
admin@localhost:8080,false
1happyturtle@localhost:8080,true
`

	// Trigger the import handler.
	suite.TriggerHandler(data, "mutes", "overwrite")

	// Wait for mutes to be applied
	if !testrig.WaitFor(func() bool {
		mutes, err := suite.state.DB.GetAccountMutes(ctx, testAccount.ID, nil)
		if err != nil {
			suite.FailNow(err.Error())
		}
		for _, m := range mutes {
			switch m.TargetAccount.ID {
			case suite.testAccounts["remote_account_1"].ID:
				suite.FailNow("fossbro is still muted")
			case suite.testAccounts["admin_account"].ID:
				if *m.Notifications != false {
					suite.FailNow("expected notifications from admin to be NOT muted")
				}
			case suite.testAccounts["local_account_2"].ID:
				if *m.Notifications != true {
					suite.FailNow("expected notifications from turtle to be muted")
				}
			default:
				suite.FailNow("unexpected muted account", m.TargetAccount)
			}
		}
		return len(mutes) == 2
	}) {
		suite.FailNow("timed out waiting for import to apply")
	}

}

func TestImportTestSuite(t *testing.T) {
	suite.Run(t, new(ImportTestSuite))
}

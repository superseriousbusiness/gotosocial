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

package bookmarks_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/admin"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/bookmarks"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/statuses"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
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

type BookmarkTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db           db.DB
	tc           *typeutils.Converter
	mediaManager *media.Manager
	federator    *federation.Federator
	emailSender  email.Sender
	processor    *processing.Processor
	storage      *storage.Driver
	state        state.State

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status
	testFollows      map[string]*gtsmodel.Follow
	testBookmarks    map[string]*gtsmodel.StatusBookmark

	// module being tested
	statusModule   *statuses.Module
	bookmarkModule *bookmarks.Module
}

func (suite *BookmarkTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testFollows = testrig.NewTestFollows()
	suite.testBookmarks = testrig.NewTestBookmarks()
}

func (suite *BookmarkTestSuite) SetupTest() {
	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	testrig.InitTestConfig()
	testrig.InitTestLog()

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

	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")

	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)
	suite.federator = testrig.NewTestFederator(&suite.state, testrig.NewTestTransportController(&suite.state, testrig.NewMockHTTPClient(nil, "../../../../testrig/media")), suite.mediaManager)
	suite.emailSender = testrig.NewEmailSender("../../../../web/template/", nil)
	suite.processor = testrig.NewTestProcessor(
		&suite.state,
		suite.federator,
		suite.emailSender,
		testrig.NewNoopWebPushSender(),
		suite.mediaManager,
	)
	suite.statusModule = statuses.New(suite.processor)
	suite.bookmarkModule = bookmarks.New(suite.processor)
}

func (suite *BookmarkTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	testrig.StopWorkers(&suite.state)
}

func (suite *BookmarkTestSuite) getBookmarks(
	account *gtsmodel.Account,
	token *gtsmodel.Token,
	user *gtsmodel.User,
	expectedHTTPStatus int,
	maxID string,
	minID string,
	limit int,
) ([]*apimodel.Status, string, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, account)
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(token))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, user)

	// create the request URI
	requestPath := bookmarks.BasePath + "?" + bookmarks.LimitKey + "=" + strconv.Itoa(limit)
	if maxID != "" {
		requestPath = requestPath + "&" + bookmarks.MaxIDKey + "=" + maxID
	}
	if minID != "" {
		requestPath = requestPath + "&" + bookmarks.MinIDKey + "=" + minID
	}
	baseURI := config.GetProtocol() + "://" + config.GetHost()
	requestURI := baseURI + "/api/" + requestPath

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodGet, requestURI, nil)
	ctx.Request.Header.Set("accept", "application/json")

	// trigger the handler
	suite.bookmarkModule.BookmarksGETHandler(ctx)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		return nil, "", fmt.Errorf("expected %d got %d", expectedHTTPStatus, resultCode)
	}

	b, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, "", err
	}

	resp := []*apimodel.Status{}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, "", err
	}

	return resp, result.Header.Get("Link"), nil
}

func (suite *BookmarkTestSuite) TestGetBookmarksSingle() {
	testAccount := suite.testAccounts["local_account_1"]
	testToken := suite.testTokens["local_account_1"]
	testUser := suite.testUsers["local_account_1"]

	statuses, linkHeader, err := suite.getBookmarks(testAccount, testToken, testUser, http.StatusOK, "", "", 10)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(statuses, 1)
	suite.Equal(`<http://localhost:8080/api/v1/bookmarks?limit=10&max_id=01F8MHD2QCZSZ6WQS2ATVPEYJ9>; rel="next", <http://localhost:8080/api/v1/bookmarks?limit=10&min_id=01F8MHD2QCZSZ6WQS2ATVPEYJ9>; rel="prev"`, linkHeader)
}

func (suite *BookmarkTestSuite) TestGetBookmarksMultiple() {
	testAccount := suite.testAccounts["local_account_1"]
	testToken := suite.testTokens["local_account_1"]
	testUser := suite.testUsers["local_account_1"]

	// Add a few extra bookmarks for this account.
	ctx := context.Background()
	for _, b := range []*gtsmodel.StatusBookmark{
		{
			ID:              "01GSZPDQYE9WZ26T501KMM876V", // oldest
			AccountID:       testAccount.ID,
			StatusID:        suite.testStatuses["admin_account_status_2"].ID,
			TargetAccountID: suite.testAccounts["admin_account"].ID,
		},
		{
			ID:              "01GSZPGHY3ACEN11D512V6MR0M",
			AccountID:       testAccount.ID,
			StatusID:        suite.testStatuses["admin_account_status_3"].ID,
			TargetAccountID: suite.testAccounts["admin_account"].ID,
		},
		{
			ID:              "01GSZPGY4ZSHNV0PR3HSBB1DDV", // newest
			AccountID:       testAccount.ID,
			StatusID:        suite.testStatuses["admin_account_status_4"].ID,
			TargetAccountID: suite.testAccounts["admin_account"].ID,
		},
	} {
		if err := suite.db.Put(ctx, b); err != nil {
			suite.FailNow(err.Error())
		}
	}

	statuses, linkHeader, err := suite.getBookmarks(testAccount, testToken, testUser, http.StatusOK, "", "", 10)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(statuses, 4)
	suite.Equal(`<http://localhost:8080/api/v1/bookmarks?limit=10&max_id=01F8MHD2QCZSZ6WQS2ATVPEYJ9>; rel="next", <http://localhost:8080/api/v1/bookmarks?limit=10&min_id=01GSZPGY4ZSHNV0PR3HSBB1DDV>; rel="prev"`, linkHeader)
}

func (suite *BookmarkTestSuite) TestGetBookmarksMultiplePaging() {
	testAccount := suite.testAccounts["local_account_1"]
	testToken := suite.testTokens["local_account_1"]
	testUser := suite.testUsers["local_account_1"]

	// Add a few extra bookmarks for this account.
	ctx := context.Background()
	for _, b := range []*gtsmodel.StatusBookmark{
		{
			ID:              "01GSZPDQYE9WZ26T501KMM876V", // oldest
			AccountID:       testAccount.ID,
			StatusID:        suite.testStatuses["admin_account_status_2"].ID,
			TargetAccountID: suite.testAccounts["admin_account"].ID,
		},
		{
			ID:              "01GSZPGHY3ACEN11D512V6MR0M",
			AccountID:       testAccount.ID,
			StatusID:        suite.testStatuses["admin_account_status_3"].ID,
			TargetAccountID: suite.testAccounts["admin_account"].ID,
		},
		{
			ID:              "01GSZPGY4ZSHNV0PR3HSBB1DDV", // newest
			AccountID:       testAccount.ID,
			StatusID:        suite.testStatuses["admin_account_status_4"].ID,
			TargetAccountID: suite.testAccounts["admin_account"].ID,
		},
	} {
		if err := suite.db.Put(ctx, b); err != nil {
			suite.FailNow(err.Error())
		}
	}

	statuses, linkHeader, err := suite.getBookmarks(testAccount, testToken, testUser, http.StatusOK, "01GSZPGY4ZSHNV0PR3HSBB1DDV", "", 10)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(statuses, 3)
	suite.Equal(`<http://localhost:8080/api/v1/bookmarks?limit=10&max_id=01F8MHD2QCZSZ6WQS2ATVPEYJ9>; rel="next", <http://localhost:8080/api/v1/bookmarks?limit=10&min_id=01GSZPGHY3ACEN11D512V6MR0M>; rel="prev"`, linkHeader)
}

func (suite *BookmarkTestSuite) TestGetBookmarksNone() {
	testAccount := suite.testAccounts["local_account_1"]
	testToken := suite.testTokens["local_account_1"]
	testUser := suite.testUsers["local_account_1"]

	// Remove all bookmarks for this account.
	if err := suite.db.DeleteStatusBookmarks(context.Background(), "", testAccount.ID); err != nil {
		suite.FailNow(err.Error())
	}

	statuses, linkHeader, err := suite.getBookmarks(testAccount, testToken, testUser, http.StatusOK, "", "", 10)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Empty(statuses)
	suite.Empty(linkHeader)
}

func (suite *BookmarkTestSuite) TestGetBookmarksNonexistentStatus() {
	testAccount := suite.testAccounts["local_account_1"]
	testToken := suite.testTokens["local_account_1"]
	testUser := suite.testUsers["local_account_1"]

	// Add a few extra bookmarks for this account.
	ctx := context.Background()
	for _, b := range []*gtsmodel.StatusBookmark{
		{
			ID:              "01GSZPDQYE9WZ26T501KMM876V", // oldest
			AccountID:       testAccount.ID,
			StatusID:        suite.testStatuses["admin_account_status_2"].ID,
			TargetAccountID: suite.testAccounts["admin_account"].ID,
		},
		{
			ID:              "01GSZPGHY3ACEN11D512V6MR0M",
			AccountID:       testAccount.ID,
			StatusID:        suite.testStatuses["admin_account_status_3"].ID,
			TargetAccountID: suite.testAccounts["admin_account"].ID,
		},
		{
			ID:              "01GSZPGY4ZSHNV0PR3HSBB1DDV", // newest
			AccountID:       testAccount.ID,
			StatusID:        "01GSZQCRX4CXPECWA5M37QNV9F", // <-- THIS ONE DOESN'T EXIST
			TargetAccountID: suite.testAccounts["admin_account"].ID,
		},
	} {
		if err := suite.db.Put(ctx, b); err != nil {
			suite.FailNow(err.Error())
		}
	}

	statuses, linkHeader, err := suite.getBookmarks(testAccount, testToken, testUser, http.StatusOK, "", "", 10)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(statuses, 3)
	suite.Equal(`<http://localhost:8080/api/v1/bookmarks?limit=10&max_id=01F8MHD2QCZSZ6WQS2ATVPEYJ9>; rel="next", <http://localhost:8080/api/v1/bookmarks?limit=10&min_id=01GSZPGHY3ACEN11D512V6MR0M>; rel="prev"`, linkHeader)
}

func TestBookmarkTestSuite(t *testing.T) {
	suite.Run(t, new(BookmarkTestSuite))
}

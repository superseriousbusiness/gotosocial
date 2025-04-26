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

package notifications_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/notifications"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

func (suite *NotificationsTestSuite) getNotifications(
	account *gtsmodel.Account,
	token *gtsmodel.Token,
	user *gtsmodel.User,
	maxID string,
	minID string,
	limit int,
	types []string,
	excludeTypes []string,
	expectedHTTPStatus int,
	expectedBody string,
) ([]*apimodel.Notification, string, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, account)
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(token))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, user)

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodGet, config.GetProtocol()+"://"+config.GetHost()+"/api/"+notifications.BasePath, nil)
	ctx.Request.Header.Set("accept", "application/json")
	query := url.Values{}
	if maxID != "" {
		query.Set(notifications.MaxIDKey, maxID)
	}
	if minID != "" {
		query.Set(notifications.MinIDKey, maxID)
	}
	if limit != 0 {
		query.Set(notifications.LimitKey, strconv.Itoa(limit))
	}
	if len(types) > 0 {
		query[notifications.TypesKey] = types
	}
	if len(excludeTypes) > 0 {
		query[notifications.ExcludeTypesKey] = excludeTypes
	}
	ctx.Request.URL.RawQuery = query.Encode()

	// trigger the handler
	suite.notificationsModule.NotificationsGETHandler(ctx)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, "", err
	}

	errs := gtserror.NewMultiError(2)

	// check code
	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs.Appendf("expected %d got %d", expectedHTTPStatus, resultCode)
	}

	// if we got an expected body, return early
	if expectedBody != "" {
		if string(b) != expectedBody {
			errs.Appendf("expected %s got %s", expectedBody, string(b))
		}
		return nil, "", errs.Combine()
	}

	resp := make([]*apimodel.Notification, 0)
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, "", err
	}

	return resp, result.Header.Get("Link"), nil
}

// Test that we can retrieve at least one notification and the expected Link header.
func (suite *NotificationsTestSuite) TestGetNotificationsSingle() {
	testAccount := suite.testAccounts["local_account_1"]
	testToken := suite.testTokens["local_account_1"]
	testUser := suite.testUsers["local_account_1"]

	maxID := ""
	minID := ""
	limit := 10
	types := []string(nil)
	excludeTypes := []string(nil)
	expectedHTTPStatus := http.StatusOK
	expectedBody := ""

	notifications, linkHeader, err := suite.getNotifications(
		testAccount,
		testToken,
		testUser,
		maxID,
		minID,
		limit,
		types,
		excludeTypes,
		expectedHTTPStatus,
		expectedBody,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(notifications, 1)
	suite.Equal(`<http://localhost:8080/api/v1/notifications?limit=10&max_id=01F8Q0ANPTWW10DAKTX7BRPBJP>; rel="next", <http://localhost:8080/api/v1/notifications?limit=10&min_id=01F8Q0ANPTWW10DAKTX7BRPBJP>; rel="prev"`, linkHeader)
}

// Add some extra notifications of different types than the fixture's single fav notification per account.
func (suite *NotificationsTestSuite) addMoreNotifications(testAccount *gtsmodel.Account) {
	for _, b := range []*gtsmodel.Notification{
		{
			ID:               id.NewULID(),
			NotificationType: gtsmodel.NotificationFollowRequest,
			TargetAccountID:  testAccount.ID,
			OriginAccountID:  suite.testAccounts["local_account_2"].ID,
		},
		{
			ID:               id.NewULID(),
			NotificationType: gtsmodel.NotificationFollow,
			TargetAccountID:  testAccount.ID,
			OriginAccountID:  suite.testAccounts["remote_account_2"].ID,
		},
	} {
		if err := suite.db.Put(context.Background(), b); err != nil {
			suite.FailNow(err.Error())
		}
	}
}

// Test that we can exclude a notification type.
func (suite *NotificationsTestSuite) TestGetNotificationsExcludeOneType() {
	testAccount := suite.testAccounts["local_account_1"]
	testToken := suite.testTokens["local_account_1"]
	testUser := suite.testUsers["local_account_1"]

	suite.addMoreNotifications(testAccount)

	maxID := ""
	minID := ""
	limit := 10
	types := []string(nil)
	excludeTypes := []string{"follow_request"}
	expectedHTTPStatus := http.StatusOK
	expectedBody := ""

	notifications, _, err := suite.getNotifications(
		testAccount,
		testToken,
		testUser,
		maxID,
		minID,
		limit,
		types,
		excludeTypes,
		expectedHTTPStatus,
		expectedBody,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// This should not include the follow request notification.
	suite.Len(notifications, 2)
	for _, notification := range notifications {
		suite.NotEqual("follow_request", notification.Type)
	}
}

// Test that we can fetch only a single notification type.
func (suite *NotificationsTestSuite) TestGetNotificationsIncludeOneType() {
	testAccount := suite.testAccounts["local_account_1"]
	testToken := suite.testTokens["local_account_1"]
	testUser := suite.testUsers["local_account_1"]

	suite.addMoreNotifications(testAccount)

	maxID := ""
	minID := ""
	limit := 10
	types := []string{"favourite"}
	excludeTypes := []string(nil)
	expectedHTTPStatus := http.StatusOK
	expectedBody := ""

	notifications, _, err := suite.getNotifications(
		testAccount,
		testToken,
		testUser,
		maxID,
		minID,
		limit,
		types,
		excludeTypes,
		expectedHTTPStatus,
		expectedBody,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// This should only include the fav notification.
	suite.Len(notifications, 1)
	for _, notification := range notifications {
		suite.Equal("favourite", notification.Type)
	}
}

// Test including an unknown notification type, it should be ignored.
func (suite *NotificationsTestSuite) TestGetNotificationsIncludeUnknownType() {
	testAccount := suite.testAccounts["local_account_1"]
	testToken := suite.testTokens["local_account_1"]
	testUser := suite.testUsers["local_account_1"]

	suite.addMoreNotifications(testAccount)

	maxID := ""
	minID := ""
	limit := 10
	types := []string{"favourite", "something.weird"}
	excludeTypes := []string(nil)
	expectedHTTPStatus := http.StatusOK
	expectedBody := ""

	notifications, _, err := suite.getNotifications(
		testAccount,
		testToken,
		testUser,
		maxID,
		minID,
		limit,
		types,
		excludeTypes,
		expectedHTTPStatus,
		expectedBody,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// This should only include the fav notification.
	suite.Len(notifications, 1)
	for _, notification := range notifications {
		suite.Equal("favourite", notification.Type)
	}
}

func TestBookmarkTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationsTestSuite))
}

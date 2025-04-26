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

package push_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/push"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
)

// getSubscription gets the push subscription for the named account and token.
func (suite *PushTestSuite) getSubscription(
	accountFixtureName string,
	tokenFixtureName string,
	expectedHTTPStatus int,
) (*apimodel.WebPushSubscription, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts[accountFixtureName])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens[tokenFixtureName]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers[accountFixtureName])

	// create the request
	requestUrl := config.GetProtocol() + "://" + config.GetHost() + "/api" + push.SubscriptionPath
	ctx.Request = httptest.NewRequest(http.MethodGet, requestUrl, nil)
	ctx.Request.Header.Set("accept", "application/json")

	// trigger the handler
	suite.pushModule.PushSubscriptionGETHandler(ctx)

	// read the response
	result := recorder.Result()
	defer func() {
		_ = result.Body.Close()
	}()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		return nil, fmt.Errorf("expected %d got %d", expectedHTTPStatus, resultCode)
	}

	resp := &apimodel.WebPushSubscription{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// Get a subscription that should exist.
func (suite *PushTestSuite) TestGetSubscription() {
	accountFixtureName := "local_account_1"
	// This token should have a subscription associated with it already, with all event types turned on.
	tokenFixtureName := "local_account_1"

	subscription, err := suite.getSubscription(accountFixtureName, tokenFixtureName, 200)
	if suite.NoError(err) {
		suite.NotEmpty(subscription.ID)
		suite.NotEmpty(subscription.Endpoint)
		suite.NotEmpty(subscription.ServerKey)
		suite.True(subscription.Alerts.Mention)
	}
}

// Get a subscription that should not exist, which should fail.
func (suite *PushTestSuite) TestGetMissingSubscription() {
	accountFixtureName := "local_account_1"
	// This token should not have a subscription.
	tokenFixtureName := "local_account_1_push_only"

	_, err := suite.getSubscription(accountFixtureName, tokenFixtureName, 404)
	suite.NoError(err)
}

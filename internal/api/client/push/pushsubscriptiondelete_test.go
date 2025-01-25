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
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/superseriousbusiness/gotosocial/internal/api/client/push"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

// deleteSubscription deletes the push subscription for the named account and token.
func (suite *PushTestSuite) deleteSubscription(
	accountFixtureName string,
	tokenFixtureName string,
	expectedHTTPStatus int,
) error {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts[accountFixtureName])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens[tokenFixtureName]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers[accountFixtureName])

	// create the request
	requestUrl := config.GetProtocol() + "://" + config.GetHost() + "/api" + push.SubscriptionPath
	ctx.Request = httptest.NewRequest(http.MethodDelete, requestUrl, nil)

	// trigger the handler
	suite.pushModule.PushSubscriptionDELETEHandler(ctx)

	// read the response
	result := recorder.Result()
	defer func() {
		_ = result.Body.Close()
	}()

	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		return fmt.Errorf("expected %d got %d", expectedHTTPStatus, resultCode)
	}

	return nil
}

// Delete a subscription that should exist.
func (suite *PushTestSuite) TestDeleteSubscription() {
	accountFixtureName := "local_account_1"
	// This token should have a subscription associated with it already.
	tokenFixtureName := "local_account_1"

	err := suite.deleteSubscription(accountFixtureName, tokenFixtureName, 200)
	suite.NoError(err)
}

// Delete a subscription that should not exist, which should succeed anyway.
func (suite *PushTestSuite) TestDeleteMissingSubscription() {
	accountFixtureName := "local_account_1"
	// This token should not have a subscription.
	tokenFixtureName := "local_account_1_user_authorization_token"

	err := suite.deleteSubscription(accountFixtureName, tokenFixtureName, 200)
	suite.NoError(err)
}

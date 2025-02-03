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
	"net/url"
	"strconv"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/api/client/push"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

// putSubscription updates the push subscription for the named account and token.
// It only allows updating two event types if using the form API. Add more if you need them.
func (suite *PushTestSuite) putSubscription(
	accountFixtureName string,
	tokenFixtureName string,
	alertsMention *bool,
	alertsStatus *bool,
	policy *string,
	requestJson *string,
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
	ctx.Request = httptest.NewRequest(http.MethodPut, requestUrl, nil)
	ctx.Request.Header.Set("accept", "application/json")

	if requestJson != nil {
		ctx.Request.Header.Set("content-type", "application/json")
		ctx.Request.Body = io.NopCloser(strings.NewReader(*requestJson))
	} else {
		ctx.Request.Form = make(url.Values)
		if alertsMention != nil {
			ctx.Request.Form["data[alerts][mention]"] = []string{strconv.FormatBool(*alertsMention)}
		}
		if alertsStatus != nil {
			ctx.Request.Form["data[alerts][status]"] = []string{strconv.FormatBool(*alertsStatus)}
		}
		if policy != nil {
			ctx.Request.Form["data[policy]"] = []string{*policy}
		}
	}

	// trigger the handler
	suite.pushModule.PushSubscriptionPUTHandler(ctx)

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

// Update a subscription that already exists.
func (suite *PushTestSuite) TestPutSubscription() {
	accountFixtureName := "local_account_1"
	// This token should have a subscription associated with it already, with all event types turned on.
	tokenFixtureName := "local_account_1"

	alertsMention := true
	alertsStatus := false
	policy := "followed"
	subscription, err := suite.putSubscription(
		accountFixtureName,
		tokenFixtureName,
		&alertsMention,
		&alertsStatus,
		&policy,
		nil,
		200,
	)
	if suite.NoError(err) {
		suite.NotEmpty(subscription.ID)
		suite.NotEmpty(subscription.Endpoint)
		suite.NotEmpty(subscription.ServerKey)
		suite.True(subscription.Alerts.Mention)
		suite.False(subscription.Alerts.Status)
		// Omitted event types should default to off.
		suite.False(subscription.Alerts.Favourite)
		suite.Equal(apimodel.WebPushNotificationPolicyFollowed, subscription.Policy)
	}
}

// Update a subscription that already exists, using the JSON format.
func (suite *PushTestSuite) TestPutSubscriptionJSON() {
	accountFixtureName := "local_account_1"
	// This token should have a subscription associated with it already, with all event types turned on.
	tokenFixtureName := "local_account_1"

	requestJson := `{
		"data": {
			"alerts": {
				"mention": true,
				"status": false
			},
			"policy": "followed"
		}
	}`
	subscription, err := suite.putSubscription(
		accountFixtureName,
		tokenFixtureName,
		nil,
		nil,
		nil,
		&requestJson,
		200,
	)
	if suite.NoError(err) {
		suite.NotEmpty(subscription.ID)
		suite.NotEmpty(subscription.Endpoint)
		suite.NotEmpty(subscription.ServerKey)
		suite.True(subscription.Alerts.Mention)
		suite.False(subscription.Alerts.Status)
		// Omitted event types should default to off.
		suite.False(subscription.Alerts.Favourite)
		suite.Equal(apimodel.WebPushNotificationPolicyFollowed, subscription.Policy)
	}
}

// Update a subscription that does not exist, which should fail.
func (suite *PushTestSuite) TestPutMissingSubscription() {
	accountFixtureName := "local_account_1"
	// This token should not have a subscription.
	tokenFixtureName := "local_account_1_user_authorization_token"

	alertsMention := true
	alertsStatus := false
	_, err := suite.putSubscription(
		accountFixtureName,
		tokenFixtureName,
		&alertsMention,
		&alertsStatus,
		nil,
		nil,
		404,
	)
	suite.NoError(err)
}

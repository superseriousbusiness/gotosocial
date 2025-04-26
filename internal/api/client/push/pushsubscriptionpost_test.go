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

	"code.superseriousbusiness.org/gotosocial/internal/api/client/push"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
)

// postSubscription creates or replaces the push subscription for the named account and token.
// It only allows updating two event types if using the form API. Add more if you need them.
func (suite *PushTestSuite) postSubscription(
	accountFixtureName string,
	tokenFixtureName string,
	endpoint *string,
	auth *string,
	p256dh *string,
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
	ctx.Request = httptest.NewRequest(http.MethodPost, requestUrl, nil)
	ctx.Request.Header.Set("accept", "application/json")

	if requestJson != nil {
		ctx.Request.Header.Set("content-type", "application/json")
		ctx.Request.Body = io.NopCloser(strings.NewReader(*requestJson))
	} else {
		ctx.Request.Form = make(url.Values)
		if endpoint != nil {
			ctx.Request.Form["subscription[endpoint]"] = []string{*endpoint}
		}
		if auth != nil {
			ctx.Request.Form["subscription[keys][auth]"] = []string{*auth}
		}
		if p256dh != nil {
			ctx.Request.Form["subscription[keys][p256dh]"] = []string{*p256dh}
		}
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
	suite.pushModule.PushSubscriptionPOSTHandler(ctx)

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

// Create a new subscription.
func (suite *PushTestSuite) TestPostSubscription() {
	accountFixtureName := "local_account_1"
	// This token should not have a subscription.
	tokenFixtureName := "local_account_1_push_only"

	endpoint := "https://example.test/push"
	auth := "cgna/fzrYLDQyPf5hD7IsA=="
	p256dh := "BMYVItYVOX+AHBdtA62Q0i6c+F7MV2Gia3aoDr8mvHkuPBNIOuTLDfmFcnBqoZcQk6BtLcIONbxhHpy2R+mYIUY="
	alertsMention := true
	alertsStatus := false
	policy := "followed"
	subscription, err := suite.postSubscription(
		accountFixtureName,
		tokenFixtureName,
		&endpoint,
		&auth,
		&p256dh,
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

// Create a new subscription with only required fields.
func (suite *PushTestSuite) TestPostSubscriptionMinimal() {
	accountFixtureName := "local_account_1"
	// This token should not have a subscription.
	tokenFixtureName := "local_account_1_push_only"

	endpoint := "https://example.test/push"
	auth := "cgna/fzrYLDQyPf5hD7IsA=="
	p256dh := "BMYVItYVOX+AHBdtA62Q0i6c+F7MV2Gia3aoDr8mvHkuPBNIOuTLDfmFcnBqoZcQk6BtLcIONbxhHpy2R+mYIUY="
	subscription, err := suite.postSubscription(
		accountFixtureName,
		tokenFixtureName,
		&endpoint,
		&auth,
		&p256dh,
		nil,
		nil,
		nil,
		nil,
		200,
	)
	if suite.NoError(err) {
		suite.NotEmpty(subscription.ID)
		suite.NotEmpty(subscription.Endpoint)
		suite.NotEmpty(subscription.ServerKey)
		// All event types should default to off.
		suite.False(subscription.Alerts.Mention)
		suite.False(subscription.Alerts.Status)
		suite.False(subscription.Alerts.Favourite)
		// Policy should default to all.
		suite.Equal(apimodel.WebPushNotificationPolicyAll, subscription.Policy)
	}
}

// Create a new subscription with a missing endpoint, which should fail.
func (suite *PushTestSuite) TestPostInvalidSubscription() {
	accountFixtureName := "local_account_1"
	// This token should not have a subscription.
	tokenFixtureName := "local_account_1_push_only"

	// No endpoint.
	auth := "cgna/fzrYLDQyPf5hD7IsA=="
	p256dh := "BMYVItYVOX+AHBdtA62Q0i6c+F7MV2Gia3aoDr8mvHkuPBNIOuTLDfmFcnBqoZcQk6BtLcIONbxhHpy2R+mYIUY="
	alertsMention := true
	alertsStatus := false
	_, err := suite.postSubscription(
		accountFixtureName,
		tokenFixtureName,
		nil,
		&auth,
		&p256dh,
		&alertsMention,
		&alertsStatus,
		nil,
		nil,
		422,
	)
	suite.NoError(err)
}

// Create a new subscription, using the JSON format.
func (suite *PushTestSuite) TestPostSubscriptionJSON() {
	accountFixtureName := "local_account_1"
	// This token should not have a subscription.
	tokenFixtureName := "local_account_1_push_only"

	requestJson := `{
		"subscription": {
			"endpoint": "https://example.test/push",
			"keys": {
				"auth": "cgna/fzrYLDQyPf5hD7IsA==",
				"p256dh": "BMYVItYVOX+AHBdtA62Q0i6c+F7MV2Gia3aoDr8mvHkuPBNIOuTLDfmFcnBqoZcQk6BtLcIONbxhHpy2R+mYIUY="
			}
		},
		"data": {
			"alerts": {
				"mention": true,
				"status": false
			},
			"policy": "followed"
		}
	}`
	subscription, err := suite.postSubscription(
		accountFixtureName,
		tokenFixtureName,
		nil,
		nil,
		nil,
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

// Create a new subscription, using the JSON format and only required fields.
func (suite *PushTestSuite) TestPostSubscriptionJSONMinimal() {
	accountFixtureName := "local_account_1"
	// This token should not have a subscription.
	tokenFixtureName := "local_account_1_push_only"

	requestJson := `{
		"subscription": {
			"endpoint": "https://example.test/push",
			"keys": {
				"auth": "cgna/fzrYLDQyPf5hD7IsA==",
				"p256dh": "BMYVItYVOX+AHBdtA62Q0i6c+F7MV2Gia3aoDr8mvHkuPBNIOuTLDfmFcnBqoZcQk6BtLcIONbxhHpy2R+mYIUY="
			}
		}
	}`
	subscription, err := suite.postSubscription(
		accountFixtureName,
		tokenFixtureName,
		nil,
		nil,
		nil,
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
		// All event types should default to off.
		suite.False(subscription.Alerts.Mention)
		suite.False(subscription.Alerts.Status)
		suite.False(subscription.Alerts.Favourite)
		// Policy should default to all.
		suite.Equal(apimodel.WebPushNotificationPolicyAll, subscription.Policy)
	}
}

// Create a new subscription with a missing endpoint, using the JSON format, which should fail.
func (suite *PushTestSuite) TestPostInvalidSubscriptionJSON() {
	accountFixtureName := "local_account_1"
	// This token should not have a subscription.
	tokenFixtureName := "local_account_1_push_only"

	// No endpoint.
	requestJson := `{
		"subscription": {
			"keys": {
				"auth": "cgna/fzrYLDQyPf5hD7IsA==",
				"p256dh": "BMYVItYVOX+AHBdtA62Q0i6c+F7MV2Gia3aoDr8mvHkuPBNIOuTLDfmFcnBqoZcQk6BtLcIONbxhHpy2R+mYIUY="
			}
		},
		"data": {
			"alerts": {
				"mention": true,
				"status": false
			}
		}
	}`
	_, err := suite.postSubscription(
		accountFixtureName,
		tokenFixtureName,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&requestJson,
		422,
	)
	suite.NoError(err)
}

// Replace a subscription that already exists.
func (suite *PushTestSuite) TestPostExistingSubscription() {
	accountFixtureName := "local_account_1"
	// This token should have a subscription associated with it already, with all event types turned on.
	tokenFixtureName := "local_account_1"

	endpoint := "https://example.test/push"
	auth := "JMFtMRgZaeHpwsDjBnhcmQ=="
	p256dh := "BMYVItYVOX+AHBdtA62Q0i6c+F7MV2Gia3aoDr8mvHkuPBNIOuTLDfmFcnBqoZcQk6BtLcIONbxhHpy2R+mYIUY="
	alertsMention := true
	alertsStatus := false
	policy := "followed"
	subscription, err := suite.postSubscription(
		accountFixtureName,
		tokenFixtureName,
		&endpoint,
		&auth,
		&p256dh,
		&alertsMention,
		&alertsStatus,
		&policy,
		nil,
		200,
	)
	if suite.NoError(err) {
		suite.NotEqual(suite.testWebPushSubscriptions["local_account_1_token_1"].ID, subscription.ID)
		suite.NotEmpty(subscription.Endpoint)
		suite.NotEmpty(subscription.ServerKey)
		suite.True(subscription.Alerts.Mention)
		suite.False(subscription.Alerts.Status)
		// Omitted event types should default to off.
		suite.False(subscription.Alerts.Favourite)
	}
}

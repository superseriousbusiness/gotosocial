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

package mutes_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/api/client/mutes"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

func (suite *MutesTestSuite) getMutedAccounts(
	expectedHTTPStatus int,
	expectedBody string,
) ([]*apimodel.MutedAccount, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodGet, config.GetProtocol()+"://"+config.GetHost()+"/api/"+mutes.BasePath, nil)
	ctx.Request.Header.Set("accept", "application/json")

	// trigger the handler
	suite.mutesModule.MutesGETHandler(ctx)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	errs := gtserror.NewMultiError(2)

	// check code + body
	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs.Appendf("expected %d got %d", expectedHTTPStatus, resultCode)
		if expectedBody == "" {
			return nil, errs.Combine()
		}
	}

	// if we got an expected body, return early
	if expectedBody != "" {
		if string(b) != expectedBody {
			errs.Appendf("expected %s got %s", expectedBody, string(b))
		}
		return nil, errs.Combine()
	}

	resp := make([]*apimodel.MutedAccount, 0)
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *MutesTestSuite) TestGetMutedAccounts() {
	// Mute a user with a finite duration.
	mute1 := &gtsmodel.UserMute{
		ID:              "01HZQ4K4MJTZ3RWVAEEJQDKK7M",
		ExpiresAt:       time.Now().Add(time.Duration(1) * time.Hour),
		AccountID:       suite.testAccounts["local_account_1"].ID,
		TargetAccountID: suite.testAccounts["local_account_2"].ID,
	}
	err := suite.db.PutMute(context.Background(), mute1)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Mute a user with an indefinite duration.
	mute2 := &gtsmodel.UserMute{
		ID:              "01HZQ4K641EMWBEJ9A99WST1GP",
		AccountID:       suite.testAccounts["local_account_1"].ID,
		TargetAccountID: suite.testAccounts["remote_account_1"].ID,
	}
	err = suite.db.PutMute(context.Background(), mute2)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Fetch all muted accounts for the logged-in account.
	mutedAccounts, err := suite.getMutedAccounts(http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotEmpty(mutedAccounts)

	// Check that we got the accounts we just muted, and that their mute expiration times are set correctly.
	// Note that the account list will be in *reverse* order by mute ID.
	if suite.Len(mutedAccounts, 2) {
		// This mute expiration should be a string.
		mutedAccount1 := mutedAccounts[1]
		suite.Equal(mute1.TargetAccountID, mutedAccount1.ID)
		suite.NotEmpty(mutedAccount1.MuteExpiresAt)

		// This mute expiration should be null.
		mutedAccount2 := mutedAccounts[0]
		suite.Equal(mute2.TargetAccountID, mutedAccount2.ID)
		suite.Nil(mutedAccount2.MuteExpiresAt)
	}
}

func (suite *MutesTestSuite) TestIndefinitelyMutedAccountSerializesMuteExpirationAsNull() {
	// Mute a user with an indefinite duration.
	mute := &gtsmodel.UserMute{
		ID:              "01HZQ4K641EMWBEJ9A99WST1GP",
		AccountID:       suite.testAccounts["local_account_1"].ID,
		TargetAccountID: suite.testAccounts["remote_account_1"].ID,
	}
	err := suite.db.PutMute(context.Background(), mute)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Fetch all muted accounts for the logged-in account.
	// The expected body contains `"mute_expires_at":null`.
	_, err = suite.getMutedAccounts(http.StatusOK, `[{"id":"01F8MH5ZK5VRH73AKHQM6Y9VNX","username":"foss_satan","acct":"foss_satan@fossbros-anonymous.io","display_name":"big gerald","locked":false,"discoverable":true,"bot":false,"created_at":"2021-09-26T10:52:36.000Z","note":"i post about like, i dunno, stuff, or whatever!!!!","url":"http://fossbros-anonymous.io/@foss_satan","avatar":"","avatar_static":"","header":"http://localhost:8080/assets/default_header.webp","header_static":"http://localhost:8080/assets/default_header.webp","followers_count":0,"following_count":0,"statuses_count":3,"last_status_at":"2021-09-11T09:40:37.000Z","emojis":[],"fields":[],"mute_expires_at":null}]`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

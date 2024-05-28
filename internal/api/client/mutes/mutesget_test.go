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
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

func (suite *MutesTestSuite) getMutedAccounts(
	expectedHTTPStatus int,
	expectedBody string,
) ([]*apimodel.Account, error) {
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

	resp := make([]*apimodel.Account, 0)
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *MutesTestSuite) TestGetMutedAccounts() {
	// Mute one user with a finite duration.
	mute := &gtsmodel.UserMute{
		ID:              id.NewULID(),
		ExpiresAt:       time.Now().Add(time.Duration(1) * time.Hour),
		AccountID:       suite.testAccounts["local_account_1"].ID,
		TargetAccountID: suite.testAccounts["local_account_2"].ID,
	}
	err := suite.db.PutMute(context.Background(), mute)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Fetch all muted accounts for the logged-in account.
	mutedAccounts, err := suite.getMutedAccounts(http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotEmpty(mutedAccounts)

	// Check that we got the account we just muted, and that it has a mute expiration time.
	if suite.Len(mutedAccounts, 1) {
		mutedAccount := mutedAccounts[0]
		suite.Equal(mute.TargetAccountID, mutedAccount.ID)
		suite.NotEmpty(mutedAccount.MuteExpiresAt)
	}
}

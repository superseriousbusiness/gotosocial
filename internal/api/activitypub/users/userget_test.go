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

package users_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/gotosocial/internal/api/activitypub/users"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type UserGetTestSuite struct {
	UserStandardTestSuite
}

func (suite *UserGetTestSuite) TestGetUser() {
	// the dereference we're gonna use
	derefRequests := testrig.NewTestDereferenceRequests(suite.testAccounts)
	signedRequest := derefRequests["foss_satan_dereference_zork"]
	targetAccount := suite.testAccounts["local_account_1"]

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetAccount.URI, nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/activity+json")
	ctx.Request.Header.Set("Signature", signedRequest.SignatureHeader)
	ctx.Request.Header.Set("Date", signedRequest.DateHeader)

	// we need to pass the context through signature check first to set appropriate values on it
	suite.signatureCheck(ctx)

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   users.UsernameKey,
			Value: targetAccount.Username,
		},
	}

	// trigger the function being tested
	suite.userModule.UsersGETHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	// should be a Person
	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	suite.NoError(err)

	t, err := streams.ToType(context.Background(), m)
	suite.NoError(err)

	person, ok := t.(vocab.ActivityStreamsPerson)
	suite.True(ok)

	// convert person to account
	a, err := suite.tc.ASRepresentationToAccount(context.Background(), person, "", "")
	suite.NoError(err)
	suite.EqualValues(targetAccount.Username, a.Username)
}

// TestGetUserPublicKeyDeleted checks whether the public key of a deleted account can still be dereferenced.
// This is needed by remote instances for authenticating delete requests and stuff like that.
func (suite *UserGetTestSuite) TestGetUserPublicKeyDeleted() {
	userModule := users.New(suite.processor)
	targetAccount := suite.testAccounts["local_account_1"]

	suite.processor.User().DeleteSelf(context.Background(), suite.testAccounts["local_account_1"])

	// wait for the account delete to be processed
	if !testrig.WaitFor(func() bool {
		a, _ := suite.db.GetAccountByID(context.Background(), targetAccount.ID)
		return !a.SuspendedAt.IsZero()
	}) {
		suite.FailNow("delete of account timed out")
	}

	// the dereference we're gonna use
	derefRequests := testrig.NewTestDereferenceRequests(suite.testAccounts)
	signedRequest := derefRequests["foss_satan_dereference_zork_public_key"]

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetAccount.PublicKeyURI, nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/activity+json")
	ctx.Request.Header.Set("Signature", signedRequest.SignatureHeader)
	ctx.Request.Header.Set("Date", signedRequest.DateHeader)

	// we need to pass the context through signature check first to set appropriate values on it
	suite.signatureCheck(ctx)

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   users.UsernameKey,
			Value: targetAccount.Username,
		},
	}

	// trigger the function being tested
	userModule.UsersGETHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	// should be a Person
	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	suite.NoError(err)

	t, err := streams.ToType(context.Background(), m)
	suite.NoError(err)

	person, ok := t.(vocab.ActivityStreamsPerson)
	suite.True(ok)

	// convert person to account
	a, err := suite.tc.ASRepresentationToAccount(context.Background(), person, "", "")
	suite.NoError(err)
	suite.EqualValues(targetAccount.Username, a.Username)
}

func TestUserGetTestSuite(t *testing.T) {
	suite.Run(t, new(UserGetTestSuite))
}

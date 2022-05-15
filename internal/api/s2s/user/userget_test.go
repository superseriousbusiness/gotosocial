/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package user_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/api/s2s/user"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type UserGetTestSuite struct {
	UserStandardTestSuite
}

func (suite *UserGetTestSuite) TestGetUser() {
	// the dereference we're gonna use
	derefRequests := testrig.NewTestDereferenceRequests(suite.testAccounts)
	signedRequest := derefRequests["foss_satan_dereference_zork"]
	targetAccount := suite.testAccounts["local_account_1"]

	clientWorker := concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)
	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db, fedWorker)
	federator := testrig.NewTestFederator(suite.db, tc, suite.storage, suite.mediaManager, fedWorker)
	emailSender := testrig.NewEmailSender("../../../../web/template/", nil)
	processor := testrig.NewTestProcessor(suite.db, suite.storage, federator, emailSender, suite.mediaManager, clientWorker, fedWorker)
	userModule := user.New(processor).(*user.Module)

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetAccount.URI, nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/activity+json")
	ctx.Request.Header.Set("Signature", signedRequest.SignatureHeader)
	ctx.Request.Header.Set("Date", signedRequest.DateHeader)

	// we need to pass the context through signature check first to set appropriate values on it
	suite.securityModule.SignatureCheck(ctx)

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   user.UsernameKey,
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
	// since this account is already known, we should get a pretty full model of it from the conversion
	a, err := suite.tc.ASRepresentationToAccount(context.Background(), person, false)
	suite.NoError(err)
	suite.EqualValues(targetAccount.Username, a.Username)
}

// TestGetUserPublicKeyDeleted checks whether the public key of a deleted account can still be dereferenced.
// This is needed by remote instances for authenticating delete requests and stuff like that.
func (suite *UserGetTestSuite) TestGetUserPublicKeyDeleted() {
	targetAccount := suite.testAccounts["local_account_1"]

	// first delete the account, as though zork had deleted himself
	authed := &oauth.Auth{
		Application: suite.testApplications["local_account_1"],
		User:        suite.testUsers["local_account_1"],
		Account:     suite.testAccounts["local_account_1"],
	}
	suite.processor.AccountDeleteLocal(context.Background(), authed, &apimodel.AccountDeleteRequest{
		Password:       "password",
		DeleteOriginID: targetAccount.ID,
	})

	// now wait just a sec for it to go through....
	time.Sleep(1 * time.Second)

	// the dereference we're gonna use
	derefRequests := testrig.NewTestDereferenceRequests(suite.testAccounts)
	signedRequest := derefRequests["foss_satan_dereference_zork_public_key"]

	clientWorker := concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)
	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db, fedWorker)
	federator := testrig.NewTestFederator(suite.db, tc, suite.storage, suite.mediaManager, fedWorker)
	emailSender := testrig.NewEmailSender("../../../../web/template/", nil)
	processor := testrig.NewTestProcessor(suite.db, suite.storage, federator, emailSender, suite.mediaManager, clientWorker, fedWorker)
	userModule := user.New(processor).(*user.Module)

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetAccount.PublicKeyURI, nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/activity+json")
	ctx.Request.Header.Set("Signature", signedRequest.SignatureHeader)
	ctx.Request.Header.Set("Date", signedRequest.DateHeader)

	// we need to pass the context through signature check first to set appropriate values on it
	suite.securityModule.SignatureCheck(ctx)

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   user.UsernameKey,
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
	a, err := suite.tc.ASRepresentationToAccount(context.Background(), person, false)
	suite.NoError(err)
	suite.EqualValues(targetAccount.Username, a.Username)
}

func TestUserGetTestSuite(t *testing.T) {
	suite.Run(t, new(UserGetTestSuite))
}

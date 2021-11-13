/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/api/s2s/user"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type OutboxGetTestSuite struct {
	UserStandardTestSuite
}

func (suite *OutboxGetTestSuite) TestGetOutbox() {
	// the dereference we're gonna use
	derefRequests := testrig.NewTestDereferenceRequests(suite.testAccounts)
	signedRequest := derefRequests["foss_satan_dereference_zork_outbox"]
	targetAccount := suite.testAccounts["local_account_1"]

	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db)
	federator := testrig.NewTestFederator(suite.db, tc, suite.storage)
	emailSender := testrig.NewEmailSender("../../../../web/template/", nil)
	processor := testrig.NewTestProcessor(suite.db, suite.storage, federator, emailSender)
	userModule := user.New(suite.config, processor).(*user.Module)

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetAccount.OutboxURI, nil) // the endpoint we're hitting
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
	userModule.OutboxGETHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.Equal(`{"@context":"https://www.w3.org/ns/activitystreams","first":"http://localhost:8080/users/the_mighty_zork/outbox?page=true","id":"http://localhost:8080/users/the_mighty_zork/outbox","type":"OrderedCollection"}`, string(b))

	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	suite.NoError(err)

	t, err := streams.ToType(context.Background(), m)
	suite.NoError(err)

	_, ok := t.(vocab.ActivityStreamsOrderedCollection)
	suite.True(ok)
}

func (suite *OutboxGetTestSuite) TestGetOutboxFirstPage() {
	// the dereference we're gonna use
	derefRequests := testrig.NewTestDereferenceRequests(suite.testAccounts)
	signedRequest := derefRequests["foss_satan_dereference_zork_outbox_first"]
	targetAccount := suite.testAccounts["local_account_1"]

	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db)
	federator := testrig.NewTestFederator(suite.db, tc, suite.storage)
	emailSender := testrig.NewEmailSender("../../../../web/template/", nil)
	processor := testrig.NewTestProcessor(suite.db, suite.storage, federator, emailSender)
	userModule := user.New(suite.config, processor).(*user.Module)

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetAccount.OutboxURI+"?page=true", nil) // the endpoint we're hitting
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
	userModule.OutboxGETHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.Equal(`{"@context":"https://www.w3.org/ns/activitystreams","id":"http://localhost:8080/users/the_mighty_zork/outbox?page=true","next":"http://localhost:8080/users/the_mighty_zork/outbox?page=true\u0026max_id=01F8MHAMCHF6Y650WCRSCP4WMY","orderedItems":{"actor":"http://localhost:8080/users/the_mighty_zork","cc":"http://localhost:8080/users/the_mighty_zork/followers","id":"http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/activity","object":"http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY","published":"2021-10-20T10:40:37Z","to":"https://www.w3.org/ns/activitystreams#Public","type":"Create"},"partOf":"http://localhost:8080/users/the_mighty_zork/outbox","prev":"http://localhost:8080/users/the_mighty_zork/outbox?page=true\u0026min_id=01F8MHAMCHF6Y650WCRSCP4WMY","type":"OrderedCollectionPage"}`, string(b))

	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	suite.NoError(err)

	t, err := streams.ToType(context.Background(), m)
	suite.NoError(err)

	_, ok := t.(vocab.ActivityStreamsOrderedCollectionPage)
	suite.True(ok)
}

func (suite *OutboxGetTestSuite) TestGetOutboxNextPage() {
	// the dereference we're gonna use
	derefRequests := testrig.NewTestDereferenceRequests(suite.testAccounts)
	signedRequest := derefRequests["foss_satan_dereference_zork_outbox_next"]
	targetAccount := suite.testAccounts["local_account_1"]

	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db)
	federator := testrig.NewTestFederator(suite.db, tc, suite.storage)
	emailSender := testrig.NewEmailSender("../../../../web/template/", nil)
	processor := testrig.NewTestProcessor(suite.db, suite.storage, federator, emailSender)
	userModule := user.New(suite.config, processor).(*user.Module)

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetAccount.OutboxURI+"?page=true&max_id=01F8MHAMCHF6Y650WCRSCP4WMY", nil) // the endpoint we're hitting
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
		gin.Param{
			Key:   user.MaxIDKey,
			Value: "01F8MHAMCHF6Y650WCRSCP4WMY",
		},
	}

	// trigger the function being tested
	userModule.OutboxGETHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.Equal(`{"@context":"https://www.w3.org/ns/activitystreams","id":"http://localhost:8080/users/the_mighty_zork/outbox?page=true\u0026maxID=01F8MHAMCHF6Y650WCRSCP4WMY","orderedItems":[],"partOf":"http://localhost:8080/users/the_mighty_zork/outbox","type":"OrderedCollectionPage"}`, string(b))

	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	suite.NoError(err)

	t, err := streams.ToType(context.Background(), m)
	suite.NoError(err)

	_, ok := t.(vocab.ActivityStreamsOrderedCollectionPage)
	suite.True(ok)
}

func TestOutboxGetTestSuite(t *testing.T) {
	suite.Run(t, new(OutboxGetTestSuite))
}

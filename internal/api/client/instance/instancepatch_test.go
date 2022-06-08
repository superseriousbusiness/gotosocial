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

package instance_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/instance"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type InstancePatchTestSuite struct {
	InstanceStandardTestSuite
}

func (suite *InstancePatchTestSuite) TestInstancePatch1() {
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"title":            "Example Instance",
			"contact_username": "admin",
			"contact_email":    "someone@example.org",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, bodyBytes, instance.InstanceInformationPath, w.FormDataContentType())

	// call the handler
	suite.instanceModule.InstanceUpdatePATCHHandler(ctx)

	// we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"uri":"http://localhost:8080","title":"Example Instance","description":"","short_description":"","email":"someone@example.org","version":"","registrations":true,"approval_required":true,"invites_enabled":false,"urls":{"streaming_api":"wss://localhost:8080"},"stats":{"domain_count":0,"status_count":16,"user_count":4},"thumbnail":"","contact_account":{"id":"01F8MH17FWEB39HZJ76B6VXSKF","username":"admin","acct":"admin","display_name":"","locked":false,"bot":false,"created_at":"2022-05-17T13:10:59.000Z","note":"","url":"http://localhost:8080/@admin","avatar":"","avatar_static":"","header":"","header_static":"","followers_count":1,"following_count":1,"statuses_count":4,"last_status_at":"2021-10-20T10:41:37.000Z","emojis":[],"fields":[]},"max_toot_chars":5000}`, string(b))
}

func (suite *InstancePatchTestSuite) TestInstancePatch2() {
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"title": "<p>Geoff's Instance</p>",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, bodyBytes, instance.InstanceInformationPath, w.FormDataContentType())

	// call the handler
	suite.instanceModule.InstanceUpdatePATCHHandler(ctx)

	// we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"uri":"http://localhost:8080","title":"Geoff's Instance","description":"","short_description":"","email":"","version":"","registrations":true,"approval_required":true,"invites_enabled":false,"urls":{"streaming_api":"wss://localhost:8080"},"stats":{"domain_count":0,"status_count":16,"user_count":4},"thumbnail":"","max_toot_chars":5000}`, string(b))
}

func (suite *InstancePatchTestSuite) TestInstancePatch3() {
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"short_description": "<p>This is some html, which is <em>allowed</em> in short descriptions.</p>",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, bodyBytes, instance.InstanceInformationPath, w.FormDataContentType())

	// call the handler
	suite.instanceModule.InstanceUpdatePATCHHandler(ctx)

	// we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"uri":"http://localhost:8080","title":"localhost:8080","description":"","short_description":"\u003cp\u003eThis is some html, which is \u003cem\u003eallowed\u003c/em\u003e in short descriptions.\u003c/p\u003e","email":"","version":"","registrations":true,"approval_required":true,"invites_enabled":false,"urls":{"streaming_api":"wss://localhost:8080"},"stats":{"domain_count":0,"status_count":16,"user_count":4},"thumbnail":"","max_toot_chars":5000}`, string(b))
}

func (suite *InstancePatchTestSuite) TestInstancePatch4() {
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, bodyBytes, instance.InstanceInformationPath, w.FormDataContentType())

	// call the handler
	suite.instanceModule.InstanceUpdatePATCHHandler(ctx)

	suite.Equal(http.StatusBadRequest, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Bad Request: empty form submitted"}`, string(b))
}

func (suite *InstancePatchTestSuite) TestInstancePatch5() {
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"short_description": "<p>This is some html, which is <em>allowed</em> in short descriptions.</p>",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPatch, bodyBytes, instance.InstanceInformationPath, w.FormDataContentType())

	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// call the handler
	suite.instanceModule.InstanceUpdatePATCHHandler(ctx)

	suite.Equal(http.StatusForbidden, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Forbidden: user is not an admin so cannot update instance settings"}`, string(b))
}

func TestInstancePatchTestSuite(t *testing.T) {
	suite.Run(t, &InstancePatchTestSuite{})
}

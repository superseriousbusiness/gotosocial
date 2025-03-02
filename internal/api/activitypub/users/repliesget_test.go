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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/superseriousbusiness/activity/streams"
	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/api/activitypub/users"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type RepliesGetTestSuite struct {
	UserStandardTestSuite
}

func (suite *RepliesGetTestSuite) TestGetReplies() {
	// the dereference we're gonna use
	derefRequests := testrig.NewTestDereferenceRequests(suite.testAccounts)
	signedRequest := derefRequests["foss_satan_dereference_local_account_1_status_1_replies"]
	targetAccount := suite.testAccounts["local_account_1"]
	targetStatus := suite.testStatuses["local_account_1_status_1"]

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetStatus.URI+"/replies?only_other_accounts=false", nil) // the endpoint we're hitting
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
		gin.Param{
			Key:   users.StatusIDKey,
			Value: targetStatus.ID,
		},
	}

	// trigger the function being tested
	suite.userModule.StatusRepliesGETHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	// Read response body.
	result := recorder.Result()
	defer result.Body.Close()
	b, err := io.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	// Indent JSON
	// for readability.
	b = indentJSON(b)

	// Create JSON string of expected output.
	expect := toJSON(map[string]any{
		"@context":   "https://www.w3.org/ns/activitystreams",
		"type":       "OrderedCollection",
		"id":         targetStatus.URI + "/replies?only_other_accounts=false",
		"first":      targetStatus.URI + "/replies?limit=20&only_other_accounts=false",
		"totalItems": 1,
	})
	assert.Equal(suite.T(), expect, string(b))

	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	assert.NoError(suite.T(), err)

	t, err := streams.ToType(context.Background(), m)
	assert.NoError(suite.T(), err)

	_, ok := t.(vocab.ActivityStreamsOrderedCollection)
	assert.True(suite.T(), ok)
}

func (suite *RepliesGetTestSuite) TestGetRepliesNext() {
	// the dereference we're gonna use
	derefRequests := testrig.NewTestDereferenceRequests(suite.testAccounts)
	signedRequest := derefRequests["foss_satan_dereference_local_account_1_status_1_replies_next"]
	targetAccount := suite.testAccounts["local_account_1"]
	targetStatus := suite.testStatuses["local_account_1_status_1"]

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetStatus.URI+"/replies?only_other_accounts=false&page=true", nil) // the endpoint we're hitting
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
		gin.Param{
			Key:   users.StatusIDKey,
			Value: targetStatus.ID,
		},
	}

	// trigger the function being tested
	suite.userModule.StatusRepliesGETHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	// Read response body.
	result := recorder.Result()
	defer result.Body.Close()
	b, err := io.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	// Indent JSON
	// for readability.
	b = indentJSON(b)

	// Create JSON string of expected output.
	expect := toJSON(map[string]any{
		"@context":     "https://www.w3.org/ns/activitystreams",
		"type":         "OrderedCollectionPage",
		"id":           targetStatus.URI + "/replies?limit=20&only_other_accounts=false",
		"partOf":       targetStatus.URI + "/replies?only_other_accounts=false",
		"next":         "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/replies?limit=20&max_id=01FF25D5Q0DH7CHD57CTRS6WK0&only_other_accounts=false",
		"prev":         "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/replies?limit=20&min_id=01FF25D5Q0DH7CHD57CTRS6WK0&only_other_accounts=false",
		"orderedItems": []string{"http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0"},
		"totalItems":   1,
	})
	assert.Equal(suite.T(), expect, string(b))

	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	assert.NoError(suite.T(), err)

	t, err := streams.ToType(context.Background(), m)
	assert.NoError(suite.T(), err)

	page, ok := t.(vocab.ActivityStreamsOrderedCollectionPage)
	assert.True(suite.T(), ok)

	assert.Equal(suite.T(), page.GetActivityStreamsOrderedItems().Len(), 1)
}

func (suite *RepliesGetTestSuite) TestGetRepliesLast() {
	// the dereference we're gonna use
	derefRequests := testrig.NewTestDereferenceRequests(suite.testAccounts)
	signedRequest := derefRequests["foss_satan_dereference_local_account_1_status_1_replies_last"]
	targetAccount := suite.testAccounts["local_account_1"]
	targetStatus := suite.testStatuses["local_account_1_status_1"]

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetStatus.URI+"/replies?min_id=01FF25D5Q0DH7CHD57CTRS6WK0&only_other_accounts=false", nil)
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
		gin.Param{
			Key:   users.StatusIDKey,
			Value: targetStatus.ID,
		},
	}

	// trigger the function being tested
	suite.userModule.StatusRepliesGETHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	// Read response body.
	result := recorder.Result()
	defer result.Body.Close()
	b, err := io.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	// Indent JSON
	// for readability.
	b = indentJSON(b)

	// Create JSON string of expected output.
	expect := toJSON(map[string]any{
		"@context":     "https://www.w3.org/ns/activitystreams",
		"type":         "OrderedCollectionPage",
		"id":           targetStatus.URI + "/replies?min_id=01FF25D5Q0DH7CHD57CTRS6WK0&only_other_accounts=false",
		"partOf":       targetStatus.URI + "/replies?only_other_accounts=false",
		"orderedItems": []any{}, // empty
		"totalItems":   1,
	})
	assert.Equal(suite.T(), expect, string(b))

	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	assert.NoError(suite.T(), err)

	t, err := streams.ToType(context.Background(), m)
	assert.NoError(suite.T(), err)

	page, ok := t.(vocab.ActivityStreamsOrderedCollectionPage)
	assert.True(suite.T(), ok)

	assert.Equal(suite.T(), page.GetActivityStreamsOrderedItems().Len(), 0)
}

func TestRepliesGetTestSuite(t *testing.T) {
	suite.Run(t, new(RepliesGetTestSuite))
}

// toJSON will return indented JSON serialized form of 'a'.
func toJSON(a any) string {
	v, ok := a.(vocab.Type)
	if ok {
		m, err := ap.Serialize(v)
		if err != nil {
			panic(err)
		}
		a = m
	}
	var dst bytes.Buffer
	enc := json.NewEncoder(&dst)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	err := enc.Encode(a)
	if err != nil {
		panic(err)
	}
	dst.Truncate(dst.Len() - 1) // drop new-line
	return dst.String()
}

// indentJSON will return indented JSON from raw provided JSON.
func indentJSON(b []byte) []byte {
	var dst bytes.Buffer
	err := json.Indent(&dst, b, "", "  ")
	if err != nil {
		panic(err)
	}
	return dst.Bytes()
}

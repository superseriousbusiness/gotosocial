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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"codeberg.org/superseriousbusiness/activity/streams"
	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/activitypub/users"
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

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetAccount.OutboxURI, nil) // the endpoint we're hitting
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
	suite.userModule.OutboxGETHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	suite.NoError(err)
	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "first": "http://localhost:8080/users/the_mighty_zork/outbox?limit=40",
  "id": "http://localhost:8080/users/the_mighty_zork/outbox",
  "totalItems": 9,
  "type": "OrderedCollection"
}`, dst.String())

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

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetAccount.OutboxURI+"?limit=40", nil) // the endpoint we're hitting
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
	suite.userModule.OutboxGETHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	b = checkDropPublished(suite.T(), b, "orderedItems")
	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	suite.NoError(err)
	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://localhost:8080/users/the_mighty_zork/outbox?limit=40",
  "next": "http://localhost:8080/users/the_mighty_zork/outbox?limit=40\u0026max_id=01F8MHAMCHF6Y650WCRSCP4WMY",
  "orderedItems": [
    {
      "actor": "http://localhost:8080/users/the_mighty_zork",
      "cc": "http://localhost:8080/users/the_mighty_zork/followers",
      "id": "http://localhost:8080/users/the_mighty_zork/statuses/01JDPZC707CKDN8N4QVWM4Z1NR/activity#Create",
      "object": "http://localhost:8080/users/the_mighty_zork/statuses/01JDPZC707CKDN8N4QVWM4Z1NR",
      "to": "https://www.w3.org/ns/activitystreams#Public",
      "type": "Create"
    },
    {
      "actor": "http://localhost:8080/users/the_mighty_zork",
      "cc": "http://localhost:8080/users/the_mighty_zork/followers",
      "id": "http://localhost:8080/users/the_mighty_zork/statuses/01HH9KYNQPA416TNJ53NSATP40/activity#Create",
      "object": "http://localhost:8080/users/the_mighty_zork/statuses/01HH9KYNQPA416TNJ53NSATP40",
      "to": "https://www.w3.org/ns/activitystreams#Public",
      "type": "Create"
    },
    {
      "actor": "http://localhost:8080/users/the_mighty_zork",
      "cc": "http://localhost:8080/users/the_mighty_zork/followers",
      "id": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/activity#Create",
      "object": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
      "to": "https://www.w3.org/ns/activitystreams#Public",
      "type": "Create"
    }
  ],
  "partOf": "http://localhost:8080/users/the_mighty_zork/outbox",
  "prev": "http://localhost:8080/users/the_mighty_zork/outbox?limit=40\u0026min_id=01JDPZC707CKDN8N4QVWM4Z1NR",
  "totalItems": 9,
  "type": "OrderedCollectionPage"
}`, dst.String())

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

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Request = httptest.NewRequest(http.MethodGet, targetAccount.OutboxURI+"?limit=40&max_id=01F8MHAMCHF6Y650WCRSCP4WMY", nil) // the endpoint we're hitting
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
			Key:   users.MaxIDKey,
			Value: "01F8MHAMCHF6Y650WCRSCP4WMY",
		},
	}

	// trigger the function being tested
	suite.userModule.OutboxGETHandler(ctx)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	suite.NoError(err)
	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://localhost:8080/users/the_mighty_zork/outbox?limit=40&max_id=01F8MHAMCHF6Y650WCRSCP4WMY",
  "orderedItems": [],
  "partOf": "http://localhost:8080/users/the_mighty_zork/outbox",
  "totalItems": 9,
  "type": "OrderedCollectionPage"
}`, dst.String())

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

// checkDropPublished checks the published field at given key position for formatting, and drops from the JSON.
// This is useful because the published property is usually set to the current time string (which is difficult to test).
func checkDropPublished(t *testing.T, b []byte, at ...string) []byte {
	m := make(map[string]any)
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("error unmarshaling json into map: %v", err)
	}

	entries := make([]map[string]any, 0)
	for _, key := range at {
		switch vt := m[key].(type) {
		case []interface{}:
			for _, t := range vt {
				if entry, ok := t.(map[string]any); ok {
					entries = append(entries, entry)
				}
			}
		}
	}

	for _, entry := range entries {
		if s, ok := entry["published"].(string); !ok {
			t.Fatal("missing published data on json")
		} else if _, err := time.Parse(time.RFC3339, s); err != nil {
			t.Fatalf("error parsing published time: %v", err)
		}
		delete(entry, "published")
	}

	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("error remarshaling json: %v", err)
	}

	return b
}

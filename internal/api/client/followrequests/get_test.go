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

package followrequests_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/oklog/ulid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/tomnomnom/linkheader"
)

// random reader according to current-time source seed.
var randRd = rand.New(rand.NewSource(time.Now().Unix()))

type GetTestSuite struct {
	FollowRequestStandardTestSuite
}

func (suite *GetTestSuite) TestGet() {
	requestingAccount := suite.testAccounts["remote_account_2"]
	targetAccount := suite.testAccounts["local_account_1"]

	// put a follow request in the database
	fr := &gtsmodel.FollowRequest{
		ID:              "01FJ1S8DX3STJJ6CEYPMZ1M0R3",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		URI:             fmt.Sprintf("%s/follow/01FJ1S8DX3STJJ6CEYPMZ1M0R3", requestingAccount.URI),
		AccountID:       requestingAccount.ID,
		TargetAccountID: targetAccount.ID,
	}

	err := suite.db.Put(context.Background(), fr)
	suite.NoError(err)

	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodGet, []byte{}, "/api/v1/follow_requests", "")

	// call the handler
	suite.followRequestModule.FollowRequestGETHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := io.ReadAll(result.Body)
	assert.NoError(suite.T(), err)
	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	suite.NoError(err)
	suite.Equal(`[
  {
    "id": "01FHMQX3GAABWSM0S2VZEC2SWC",
    "username": "Some_User",
    "acct": "Some_User@example.org",
    "display_name": "some user",
    "locked": true,
    "discoverable": true,
    "bot": false,
    "created_at": "2020-08-10T12:13:28.000Z",
    "note": "i'm a real son of a gun",
    "url": "http://example.org/@Some_User",
    "avatar": "",
    "avatar_static": "",
    "header": "http://localhost:8080/assets/default_header.png",
    "header_static": "http://localhost:8080/assets/default_header.png",
    "followers_count": 0,
    "following_count": 0,
    "statuses_count": 0,
    "last_status_at": null,
    "emojis": [],
    "fields": []
  }
]`, dst.String())
}

func (suite *GetTestSuite) TestGetPagedNextLimit2() { suite.testGetPaged(2, "next") }
func (suite *GetTestSuite) TestGetPagedNextLimit4() { suite.testGetPaged(4, "next") }
func (suite *GetTestSuite) TestGetPagedNextLimit6() { suite.testGetPaged(6, "next") }

func (suite *GetTestSuite) TestGetPagedPrevLimit2() { suite.testGetPaged(2, "prev") }
func (suite *GetTestSuite) TestGetPagedPrevLimit4() { suite.testGetPaged(4, "prev") }
func (suite *GetTestSuite) TestGetPagedPrevLimit6() { suite.testGetPaged(6, "prev") }

func (suite *GetTestSuite) testGetPaged(limit int, ref string) {
	ctx := context.Background()

	// The authed local account we are going to use for HTTP requests
	requestingAccount := suite.testAccounts["local_account_1"]

	// Esnure no account blocks exist between accounts.
	_ = suite.db.DeleteAccountBlocks(
		context.Background(),
		requestingAccount.ID,
	)

	// Ensure no account follows exist between accounts.
	_ = suite.db.DeleteAccountFollows(
		context.Background(),
		requestingAccount.ID,
	)

	// Ensure no account follow_requests exist between accounts.
	_ = suite.db.DeleteAccountFollowRequests(
		context.Background(),
		requestingAccount.ID,
	)

	// Use iterable now time as constant (no syscalls)
	now, _ := time.Parse("01/02/2006", "01/02/2006")

	for _, targetAccount := range suite.testAccounts {
		if targetAccount.ID == requestingAccount.ID {
			// we cannot be our own target...
			continue
		}

		// Convert now to timestamp.
		ts := ulid.Timestamp(now)

		// Create anew ulid for now.
		u := ulid.MustNew(ts, randRd)

		// put a follow request in the database
		err := suite.db.PutFollowRequest(ctx, &gtsmodel.FollowRequest{
			ID:              u.String(),
			CreatedAt:       now,
			UpdatedAt:       now,
			URI:             fmt.Sprintf("%s/follow/%s", targetAccount.URI, u.String()),
			AccountID:       targetAccount.ID,
			TargetAccountID: requestingAccount.ID,
		})
		suite.NoError(err)

		// Bump now by 1 second.
		now = now.Add(time.Second)
	}

	// Get _ALL_ follow requests we expect to see without any paging (this filters invisible).
	apiRsp, err := suite.processor.Account().FollowRequestsGet(ctx, requestingAccount, nil)
	suite.NoError(err)
	expectAccounts := apiRsp.Items // interfaced{} account slice

	// Set the start query, only need to set limit fow now.
	nextQuery := fmt.Sprintf("limit=%d", limit)

	for p := 0; ; p++ {
		// Prepare new request for endpoint
		recorder := httptest.NewRecorder()
		ctx := suite.newContext(recorder, http.MethodGet, []byte{}, "/api/v1/follow_requests", "")
		ctx.Request.URL.RawQuery = nextQuery // setting provided next query value

		// call the handler and check for valid response code.
		suite.followRequestModule.FollowRequestGETHandler(ctx)
		suite.Equal(http.StatusOK, recorder.Code)

		var accounts []*model.Account

		// Decode response body into API account models
		result := recorder.Result()
		dec := json.NewDecoder(result.Body)
		err := dec.Decode(&accounts)
		suite.NoError(err)
		_ = result.Body.Close()

		for i := 0; i < len(accounts); i++ {
			iface := expectAccounts[0]

			// Check that expected account matches received.
			expectAccID := iface.(*model.Account).ID
			receivdAccID := accounts[i].ID
			suite.Equal(expectAccID, receivdAccID, "unexpected account at position in response on page=%d", p)

			// Drop checked account from expected.
			expectAccounts = expectAccounts[1:]
		}

		if len(expectAccounts) == 0 {
			// Reached end.
			break
		}

		// Parse response link header values.
		links := linkheader.ParseMultiple(recorder.Header().Values("Link"))
		filteredLinks := links.FilterByRel("next")
		suite.NotEmpty(filteredLinks, "no next link provided with more remaining accounts on page=%d", p)

		// A ref link header was set.
		link := filteredLinks[0]

		// Parse URI from URI string.
		uri, err := url.Parse(link.URL)
		suite.NoError(err)

		// Set next raw query value.
		nextQuery = uri.RawQuery
	}
}

func TestGetTestSuite(t *testing.T) {
	suite.Run(t, &GetTestSuite{})
}

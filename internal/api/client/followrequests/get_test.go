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

func (suite *GetTestSuite) TestGetPaged() {
	var targetAccounts []*gtsmodel.Account

	requestingAccount := suite.testAccounts["local_account_1"]

	// Use iterable now time as constant (no syscalls)
	now, _ := time.Parse("01/02/2006", "01/02/2006")

	for _, targetAccount := range suite.testAccounts {
		if targetAccount.ID == requestingAccount.ID {
			// we cannot be our own target...
			continue
		}

		// Ensure no follow request already exists.
		_ = suite.db.DeleteFollowRequest(
			context.Background(),
			targetAccount.ID,
			requestingAccount.ID,
		)

		// Convert now to timestamp.
		t := ulid.Timestamp(now)

		// Create anew ulid for now.
		u := ulid.MustNew(t, randRd)

		// put a follow request in the database
		fr := &gtsmodel.FollowRequest{
			ID:              u.String(),
			CreatedAt:       now,
			UpdatedAt:       now,
			URI:             fmt.Sprintf("%s/follow/%s", targetAccount.URI, u.String()),
			AccountID:       targetAccount.ID,
			TargetAccountID: requestingAccount.ID,
		}
		err := suite.db.Put(context.Background(), fr)
		suite.NoError(err)

		// Add target to account slice
		targetAccounts = append(targetAccounts, targetAccount)

		// Bump now by 1 second.
		now = now.Add(time.Second)
	}

	const limit = 2

	// How many rounds of pages to check.
	rounds := len(targetAccounts) / limit

	// NOTE:
	// we order our follow request account IDs by the age of
	// the follow request, so the order of targetAccounts should
	// be the same order we get them from the API endpoint.
	//
	// Further NOTE:
	// we don't actually bother setting maxID in this test.
	nextQuery := fmt.Sprintf("limit=%d", limit) // starting value

	for i := 0; i < rounds; i++ {
		recorder := httptest.NewRecorder()
		ctx := suite.newContext(recorder, http.MethodGet, []byte{}, "/api/v1/follow_requests", "")

		// Update request query to add paging.
		ctx.Request.URL.RawQuery = nextQuery
		nextQuery = "" // reset

		// call the handler
		suite.followRequestModule.FollowRequestGETHandler(ctx)

		// 1. we should have OK because our request was valid
		suite.Equal(http.StatusOK, recorder.Code)

		// 2. we should have no error message in the result body
		result := recorder.Result()
		defer result.Body.Close()

		var accounts []model.Account

		// Decode response body into API account models
		dec := json.NewDecoder(result.Body)
		err := dec.Decode(&accounts)
		suite.NoError(err)
		_ = result.Body.Close()

		// Expected number of accounts returned.
		expectLen := limit
		if expectLen > len(targetAccounts) {
			expectLen = len(targetAccounts)
		}

		if len(accounts) != expectLen {
			// This indicates we've been served less accounts than 'limit'
			// but we haven't reached the end of our expected targetAccounts.
			suite.T().Errorf("incorrect number of returned accounts: page=%d accounts=%+v", i, accounts)
			expectLen = len(accounts) // ensures no panic and can at least check account order
		}

		// Take a slice of expected accounts,
		// drop these now from targetAccounts.
		expect := targetAccounts[:expectLen]
		targetAccounts = targetAccounts[expectLen:]

		for j := range expect {
			if expect[j].ID != accounts[j].ID {
				suite.T().Errorf("unexpected account at position in paged response: page=%d accounts=%+v", i, accounts)
				break
			}
		}

		// Parse response link header values.
		links := linkheader.ParseMultiple(recorder.Header().Values("Link"))
		nextLinks := links.FilterByRel("prev")
		if len(nextLinks) != 1 && len(targetAccounts) > 0 {
			suite.T().Error("no next link provided with more remaining accounts")
			break
		}

		// A next link header was set.
		nextLink := nextLinks[0]

		// Parse URI from next URI string.
		nextURI, err := url.Parse(nextLink.URL)
		if err != nil {
			suite.T().Errorf("invalid returned link header next uri: %v", err)
			break
		}

		// Set next raw query value.
		nextQuery = nextURI.RawQuery
	}
}

func TestGetTestSuite(t *testing.T) {
	suite.Run(t, &GetTestSuite{})
}

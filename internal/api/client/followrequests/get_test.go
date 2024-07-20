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
	"strconv"
	"testing"
	"time"

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
    "header": "http://localhost:8080/assets/default_header.webp",
    "header_static": "http://localhost:8080/assets/default_header.webp",
    "followers_count": 0,
    "following_count": 0,
    "statuses_count": 1,
    "last_status_at": "2023-11-02T10:44:25.000Z",
    "emojis": [],
    "fields": []
  }
]`, dst.String())
}

func (suite *GetTestSuite) TestGetPageNewestToOldestLimit2() {
	suite.testGetPage(2, "newestToOldest")
}

func (suite *GetTestSuite) TestGetPageNewestToOldestLimit4() {
	suite.testGetPage(4, "newestToOldest")
}

func (suite *GetTestSuite) TestGetPageNewestToOldestLimit6() {
	suite.testGetPage(6, "newestToOldest")
}

func (suite *GetTestSuite) TestGetPageOldestToNewestLimit2() {
	suite.testGetPage(2, "oldestToNewest")
}

func (suite *GetTestSuite) TestGetPageOldestToNewestLimit4() {
	suite.testGetPage(4, "oldestToNewest")
}

func (suite *GetTestSuite) TestGetPageOldestToNewestLimit6() {
	suite.testGetPage(6, "oldestToNewest")
}

func (suite *GetTestSuite) testGetPage(limit int, direction string) {
	ctx := context.Background()

	// The authed local account we are going to use for HTTP requests
	requestingAccount := suite.testAccounts["local_account_1"]
	suite.clearAccountRelations(requestingAccount.ID)

	// Get current time.
	now := time.Now()

	var i int

	// Have each account in the testrig follow req the
	// account requesting their followers from the API.
	for _, account := range suite.testAccounts {
		targetAccount := requestingAccount
		if account.ID == requestingAccount.ID {
			// we cannot be our own target...
			continue
		}

		// Get next simple ID.
		id := strconv.Itoa(i)
		i++

		// put a follow request in the database
		err := suite.db.PutFollowRequest(ctx, &gtsmodel.FollowRequest{
			ID:              id,
			CreatedAt:       now,
			UpdatedAt:       now,
			URI:             fmt.Sprintf("%s/follow/%s", account.URI, id),
			AccountID:       account.ID,
			TargetAccountID: targetAccount.ID,
		})
		suite.NoError(err)

		// Bump now by 1 second.
		now = now.Add(time.Second)
	}

	// Get _ALL_ follow requests we expect to see without any paging (this filters invisible).
	apiRsp, err := suite.processor.Account().FollowRequestsGet(ctx, requestingAccount, nil)
	suite.NoError(err)
	expectAccounts := apiRsp.Items // interfaced{} account slice

	// Iteratively set
	// link query string.
	var query string

	switch direction {
	case "newestToOldest":
		// Set the starting query to page from
		// newest (ie., first entry in slice).
		acc := expectAccounts[0].(*model.Account)
		newest, _ := suite.db.GetFollowRequest(ctx, acc.ID, requestingAccount.ID)
		expectAccounts = expectAccounts[1:]
		query = fmt.Sprintf("limit=%d&max_id=%s", limit, newest.ID)

	case "oldestToNewest":
		// Set the starting query to page from
		// oldest (ie., last entry in slice).
		acc := expectAccounts[len(expectAccounts)-1].(*model.Account)
		oldest, _ := suite.db.GetFollowRequest(ctx, acc.ID, requestingAccount.ID)
		expectAccounts = expectAccounts[:len(expectAccounts)-1]
		query = fmt.Sprintf("limit=%d&min_id=%s", limit, oldest.ID)
	}

	for p := 0; ; p++ {
		// Prepare new request for endpoint
		recorder := httptest.NewRecorder()
		ctx := suite.newContext(recorder, http.MethodGet, []byte{}, "/api/v1/follow_requests", "")
		ctx.Request.URL.RawQuery = query // setting provided next query value

		// call the handler and check for valid response code.
		suite.T().Logf("direction=%q page=%d query=%q", direction, p, query)
		suite.followRequestModule.FollowRequestGETHandler(ctx)
		suite.Equal(http.StatusOK, recorder.Code)

		var accounts []*model.Account

		// Decode response body into API account models
		result := recorder.Result()
		dec := json.NewDecoder(result.Body)
		err := dec.Decode(&accounts)
		suite.NoError(err)
		_ = result.Body.Close()

		var (

			// start provides the starting index for loop in accounts.
			start func([]*model.Account) int

			// iter performs the loop iter step with index.
			iter func(int) int

			// check performs the loop conditional check against index and accounts.
			check func(int, []*model.Account) bool

			// expect pulls the next account to check against from expectAccounts.
			expect func([]interface{}) interface{}

			// trunc drops the last checked account from expectAccounts.
			trunc func([]interface{}) []interface{}
		)

		switch direction {
		case "newestToOldest":
			// When paging newest to oldest (ie., first page to last page):
			// - iter from start of received accounts
			// - iterate backward through received accounts
			// - stop when we reach last index of received accounts
			// - compare each received with the first index of expected accounts
			// - after each compare, drop the first index of expected accounts
			start = func([]*model.Account) int { return 0 }
			iter = func(i int) int { return i + 1 }
			check = func(idx int, i []*model.Account) bool { return idx < len(i) }
			expect = func(i []interface{}) interface{} { return i[0] }
			trunc = func(i []interface{}) []interface{} { return i[1:] }

		case "oldestToNewest":
			// When paging oldest to newest (ie., last page to first page):
			// - iter from end of received accounts
			// - iterate backward through received accounts
			// - stop when we reach first index of received accounts
			// - compare each received with the last index of expected accounts
			// - after each compare, drop the last index of expected accounts
			start = func(i []*model.Account) int { return len(i) - 1 }
			iter = func(i int) int { return i - 1 }
			check = func(idx int, _ []*model.Account) bool { return idx >= 0 }
			expect = func(i []interface{}) interface{} { return i[len(i)-1] }
			trunc = func(i []interface{}) []interface{} { return i[:len(i)-1] }
		}

		for i := start(accounts); check(i, accounts); i = iter(i) {
			// Get next expected account.
			iface := expect(expectAccounts)

			// Check that expected account matches received.
			expectAccID := iface.(*model.Account).ID
			receivdAccID := accounts[i].ID
			suite.Equal(expectAccID, receivdAccID, "unexpected account at position in response on page=%d", p)

			// Drop checked from expected accounts.
			expectAccounts = trunc(expectAccounts)
		}

		if len(expectAccounts) == 0 {
			// Reached end.
			break
		}

		// Parse response link header values.
		values := result.Header.Values("Link")
		links := linkheader.ParseMultiple(values)

		var filteredLinks linkheader.Links
		if direction == "newestToOldest" {
			filteredLinks = links.FilterByRel("next")
		} else {
			filteredLinks = links.FilterByRel("prev")
		}

		suite.NotEmpty(filteredLinks, "no next link provided with more remaining accounts on page=%d", p)

		// A ref link header was set.
		link := filteredLinks[0]

		// Parse URI from URI string.
		uri, err := url.Parse(link.URL)
		suite.NoError(err)

		// Set next raw query value.
		query = uri.RawQuery
	}
}

func (suite *GetTestSuite) clearAccountRelations(id string) {
	// Esnure no account blocks exist between accounts.
	_ = suite.db.DeleteAccountBlocks(
		context.Background(),
		id,
	)

	// Ensure no account follows exist between accounts.
	_ = suite.db.DeleteAccountFollows(
		context.Background(),
		id,
	)

	// Ensure no account follow_requests exist between accounts.
	_ = suite.db.DeleteAccountFollowRequests(
		context.Background(),
		id,
	)
}

func TestGetTestSuite(t *testing.T) {
	suite.Run(t, &GetTestSuite{})
}

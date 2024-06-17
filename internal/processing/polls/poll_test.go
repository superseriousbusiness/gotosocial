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

package polls_test

import (
	"context"
	"math/rand"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/processing/common"
	"github.com/superseriousbusiness/gotosocial/internal/processing/polls"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type PollTestSuite struct {
	suite.Suite
	state  state.State
	filter *visibility.Filter
	polls  polls.Processor

	testAccounts map[string]*gtsmodel.Account
	testPolls    map[string]*gtsmodel.Poll
}

func (suite *PollTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()
	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)
	testrig.NewTestDB(&suite.state)
	converter := typeutils.NewConverter(&suite.state)
	controller := testrig.NewTestTransportController(&suite.state, nil)
	mediaMgr := media.NewManager(&suite.state)
	federator := testrig.NewTestFederator(&suite.state, controller, mediaMgr)
	suite.filter = visibility.NewFilter(&suite.state)
	common := common.New(&suite.state, mediaMgr, converter, federator, suite.filter)
	suite.polls = polls.New(&common, &suite.state, converter)
}

func (suite *PollTestSuite) TearDownTest() {
	testrig.StopWorkers(&suite.state)
	testrig.StandardDBTeardown(suite.state.DB)
}

func (suite *PollTestSuite) TestPollGet() {
	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Perform test for all requester + poll combos.
	for _, account := range suite.testAccounts {
		for _, poll := range suite.testPolls {
			suite.testPollGet(ctx, account, poll)
		}
	}
}

func (suite *PollTestSuite) testPollGet(ctx context.Context, requester *gtsmodel.Account, poll *gtsmodel.Poll) {
	// Ensure poll model is fully populated before anything.
	if err := suite.state.DB.PopulatePoll(ctx, poll); err != nil {
		suite.T().Fatalf("error populating poll: %v", err)
	}

	var check func(*apimodel.Poll, gtserror.WithCode) bool

	switch {
	case !pollIsVisible(suite.filter, ctx, requester, poll):
		// Poll should not be visible to requester, this should
		// return an error code 404 (to prevent info leak).
		check = func(poll *apimodel.Poll, err gtserror.WithCode) bool {
			return poll == nil && err.Code() == http.StatusNotFound
		}

	default:
		// All other cases should succeed! i.e. no error and poll returned.
		check = func(poll *apimodel.Poll, err gtserror.WithCode) bool {
			return poll != nil && err == nil
		}
	}

	// Perform the poll vote and check the expected response.
	if !check(suite.polls.PollGet(ctx, requester, poll.ID)) {
		suite.T().Errorf("unexpected response for poll get by %s", requester.DisplayName)
	}

}

func (suite *PollTestSuite) TestPollVote() {
	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// randomChoices generates random vote choices in poll.
	randomChoices := func(poll *gtsmodel.Poll) []int {
		var max int
		if *poll.Multiple {
			max = len(poll.Options)
		} else {
			max = 1
		}
		count := 1 + rand.Intn(max)
		choices := make([]int, count)
		for i := range choices {
			choices[i] = rand.Intn(len(poll.Options))
		}
		return choices
	}

	// Perform test for all requester + poll combos.
	for _, account := range suite.testAccounts {
		for _, poll := range suite.testPolls {
			// Generate some valid choices and test.
			choices := randomChoices(poll)
			suite.testPollVote(ctx,
				account,
				poll,
				choices,
			)

			// Test with empty choices.
			suite.testPollVote(ctx,
				account,
				poll,
				nil,
			)

			// Test with out of range choice.
			suite.testPollVote(ctx,
				account,
				poll,
				[]int{len(poll.Options)},
			)
		}
	}
}

func (suite *PollTestSuite) testPollVote(ctx context.Context, requester *gtsmodel.Account, poll *gtsmodel.Poll, choices []int) {
	// Ensure poll model is fully populated before anything.
	if err := suite.state.DB.PopulatePoll(ctx, poll); err != nil {
		suite.T().Fatalf("error populating poll: %v", err)
	}

	var check func(*apimodel.Poll, gtserror.WithCode) bool

	switch {
	case !poll.ClosedAt.IsZero():
		// Poll is already closed, i.e. no new votes allowed!
		// This should return an error 422 (unprocessable entity).
		check = func(poll *apimodel.Poll, err gtserror.WithCode) bool {
			return poll == nil && err.Code() == http.StatusUnprocessableEntity
		}

	case !voteChoicesAreValid(poll, choices):
		// These are invalid vote choices, this should return
		// an error code 400 to indicate invalid request data.
		check = func(poll *apimodel.Poll, err gtserror.WithCode) bool {
			return poll == nil && err.Code() == http.StatusBadRequest
		}

	case poll.Status.AccountID == requester.ID:
		// Immediately we know that poll owner cannot vote in
		// their own poll. this should return an error 422.
		check = func(poll *apimodel.Poll, err gtserror.WithCode) bool {
			return poll == nil && err.Code() == http.StatusUnprocessableEntity
		}

	case !pollIsVisible(suite.filter, ctx, requester, poll):
		// Poll should not be visible to requester, this should
		// return an error code 404 (to prevent info leak).
		check = func(poll *apimodel.Poll, err gtserror.WithCode) bool {
			return poll == nil && err.Code() == http.StatusNotFound
		}

	default:
		// All other cases should succeed! i.e. no error and poll returned.
		check = func(poll *apimodel.Poll, err gtserror.WithCode) bool {
			return poll != nil && err == nil
		}
	}

	// Perform the poll vote and check the expected response.
	if !check(suite.polls.PollVote(ctx, requester, poll.ID, choices)) {
		suite.T().Errorf("unexpected response for poll vote by %s with %v", requester.DisplayName, choices)
	}
}

// voteChoicesAreValid is a utility function to check whether choices are valid for poll.
func voteChoicesAreValid(poll *gtsmodel.Poll, choices []int) bool {
	if len(choices) == 0 || !*poll.Multiple && len(choices) > 1 {
		// Invalid number of vote choices.
		return false
	}
	for _, choice := range choices {
		if choice < 0 || choice >= len(poll.Options) {
			// Choice index out of range.
			return false
		}
	}
	return true
}

// pollIsVisible is a short-hand function to return only a single boolean value for a visibility check on poll source status to account.
func pollIsVisible(filter *visibility.Filter, ctx context.Context, to *gtsmodel.Account, poll *gtsmodel.Poll) bool {
	visible, _ := filter.StatusVisible(ctx, to, poll.Status)
	return visible
}

func TestPollTestSuite(t *testing.T) {
	suite.Run(t, new(PollTestSuite))
}

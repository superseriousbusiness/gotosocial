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

package bundb_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type PollTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *PollTestSuite) TestGetPollBy() {
	t := suite.T()

	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Sentinel error to mark avoiding a test case.
	sentinelErr := errors.New("sentinel")

	// isEqual checks if 2 poll models are equal.
	isEqual := func(p1, p2 gtsmodel.Poll) bool {
		// Clear populated sub-models.
		p1.Status = nil
		p2.Status = nil

		// Localize all of the time fields.
		p1.ExpiresAt = p1.ExpiresAt.Local()
		p2.ExpiresAt = p2.ExpiresAt.Local()
		p1.ClosedAt = p1.ClosedAt.Local()
		p2.ClosedAt = p2.ClosedAt.Local()

		// Perform the comparison.
		return suite.Equal(p1, p2)
	}

	for _, poll := range suite.testPolls {
		for lookup, dbfunc := range map[string]func() (*gtsmodel.Poll, error){
			"id": func() (*gtsmodel.Poll, error) {
				return suite.db.GetPollByID(ctx, poll.ID)
			},
		} {

			// Clear database caches.
			suite.state.Caches.Init()

			t.Logf("checking database lookup %q", lookup)

			// Perform database function.
			checkPoll, err := dbfunc()
			if err != nil {
				if err == sentinelErr {
					continue
				}

				t.Errorf("error encountered for database lookup %q: %v", lookup, err)
				continue
			}

			// Check received account data.
			if !isEqual(*checkPoll, *poll) {
				t.Errorf("poll does not contain expected data: %+v", checkPoll)
				continue
			}

			// Check that poll source status populated.
			if poll.StatusID != (*checkPoll).Status.ID {
				t.Errorf("poll source status not correctly populated for: %+v", poll)
				continue
			}
		}
	}
}

func (suite *PollTestSuite) TestGetPollVoteBy() {
	t := suite.T()

	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Sentinel error to mark avoiding a test case.
	sentinelErr := errors.New("sentinel")

	// isEqual checks if 2 poll vote models are equal.
	isEqual := func(v1, v2 gtsmodel.PollVote) bool {
		// Clear populated sub-models.
		v1.Poll = nil
		v2.Poll = nil
		v1.Account = nil
		v2.Account = nil

		// Localize all of the time fields.
		v1.CreatedAt = v1.CreatedAt.Local()
		v2.CreatedAt = v2.CreatedAt.Local()

		// Perform the comparison.
		return suite.Equal(v1, v2)
	}

	for _, vote := range suite.testPollVotes {
		for lookup, dbfunc := range map[string]func() (*gtsmodel.PollVote, error){
			"id": func() (*gtsmodel.PollVote, error) {
				return suite.db.GetPollVoteByID(ctx, vote.ID)
			},

			"poll_id_account_id": func() (*gtsmodel.PollVote, error) {
				return suite.db.GetPollVoteBy(ctx, vote.PollID, vote.AccountID)
			},
		} {

			// Clear database caches.
			suite.state.Caches.Init()

			t.Logf("checking database lookup %q", lookup)

			// Perform database function.
			checkVote, err := dbfunc()
			if err != nil {
				if err == sentinelErr {
					continue
				}

				t.Errorf("error encountered for database lookup %q: %v", lookup, err)
				continue
			}

			// Check received account data.
			if !isEqual(*checkVote, *vote) {
				t.Errorf("poll vote does not contain expected data: %+v", checkVote)
				continue
			}

			// Check that vote source poll populated.
			if checkVote.PollID != (*checkVote).Poll.ID {
				t.Errorf("vote source poll not correctly populated for: %+v", vote)
				continue
			}

			// Check that vote author account populated.
			if checkVote.AccountID != (*checkVote).Account.ID {
				t.Errorf("vote author account not correctly populated for: %+v", vote)
				continue
			}
		}
	}
}

func (suite *PollTestSuite) TestUpdatePoll() {
	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	for _, poll := range suite.testPolls {
		// Take copy of poll.
		poll := util.Ptr(*poll)

		// Update the poll closed field.
		poll.ClosedAt = time.Now()

		// Update poll model in the database.
		err := suite.db.UpdatePoll(ctx, poll)
		suite.NoError(err)

		// Refetch poll from database to get latest.
		latest, err := suite.db.GetPollByID(ctx, poll.ID)
		suite.NoError(err)

		// The latest poll should have updated closedAt.
		suite.Equal(poll.ClosedAt, latest.ClosedAt)
	}
}

func (suite *PollTestSuite) TestPutPoll() {
	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	for _, poll := range suite.testPolls {
		// Delete this poll from the database.
		err := suite.db.DeletePollByID(ctx, poll.ID)
		suite.NoError(err)

		// Ensure that afterwards we can
		// enter it again into database.
		err = suite.db.PutPoll(ctx, poll)

		// Ensure that afterwards we can fetch poll.
		_, err = suite.db.GetPollByID(ctx, poll.ID)
		suite.NoError(err)
	}
}

func (suite *PollTestSuite) TestPutPollVote() {
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

	for _, poll := range suite.testPolls {
		// Create a new vote to insert for poll.
		vote := &gtsmodel.PollVote{
			ID:        id.NewULID(),
			Choices:   randomChoices(poll),
			PollID:    poll.ID,
			AccountID: id.NewULID(), // random account, doesn't matter
		}

		// Insert this new vote into database.
		err := suite.db.PutPollVote(ctx, vote)
		suite.NoError(err)

		// Fetch latest version of poll from database.
		latest, err := suite.db.GetPollByID(ctx, poll.ID)
		suite.NoError(err)

		// Decr latest version choices by new vote's.
		for _, choice := range vote.Choices {
			latest.Votes[choice]--
		}
		(*latest.Voters)--

		// Old poll and latest model after decr
		// should have equal vote + voter counts.
		suite.Equal(poll.Voters, latest.Voters)
		suite.Equal(poll.Votes, latest.Votes)
	}
}

func (suite *PollTestSuite) TestDeletePoll() {
	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	for _, poll := range suite.testPolls {
		// Delete this poll from the database.
		err := suite.db.DeletePollByID(ctx, poll.ID)
		suite.NoError(err)

		// Ensure that afterwards we cannot fetch poll.
		_, err = suite.db.GetPollByID(ctx, poll.ID)
		suite.ErrorIs(err, db.ErrNoEntries)
	}
}

func (suite *PollTestSuite) TestDeletePollVotesBy() {
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	for _, vote := range suite.testPollVotes {
		// Fetch before version of pollBefore from database.
		pollBefore, err := suite.db.GetPollByID(ctx, vote.PollID)
		suite.NoError(err)

		// Delete this poll vote.
		err = suite.db.DeletePollVoteBy(ctx, vote.PollID, vote.AccountID)
		suite.NoError(err)

		// Fetch after version of poll from database.
		pollAfter, err := suite.db.GetPollByID(ctx, vote.PollID)
		suite.NoError(err)

		// Voters count should be reduced by 1.
		suite.Equal(*pollBefore.Voters-1, *pollAfter.Voters)
	}
}

func (suite *PollTestSuite) TestDeletePollVotesByNoAccount() {
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Try to delete a poll by nonexisting account.
	pollID := suite.testPolls["local_account_1_status_6_poll"].ID
	nonAccountID := "01HF6T545G1G8ZNMY1S3ZXJ608"

	err := suite.db.DeletePollVoteBy(ctx, pollID, nonAccountID)
	suite.NoError(err)
}

func TestPollTestSuite(t *testing.T) {
	suite.Run(t, new(PollTestSuite))
}

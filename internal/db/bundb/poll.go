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

package bundb

import (
	"context"
	"errors"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type pollDB struct {
	db    *DB
	state *state.State
}

func (p *pollDB) GetPollByID(ctx context.Context, id string) (*gtsmodel.Poll, error) {
	return p.getPoll(
		ctx,
		"ID",
		func(poll *gtsmodel.Poll) error {
			return p.db.NewSelect().Model(poll).Where("? = ?", bun.Ident("poll.id"), id).Scan(ctx)
		},
		id,
	)
}

func (p *pollDB) GetPollByStatusID(ctx context.Context, statusID string) (*gtsmodel.Poll, error) {
	return p.getPoll(
		ctx,
		"StatusID",
		func(poll *gtsmodel.Poll) error {
			return p.db.NewSelect().Model(poll).Where("? = ?", bun.Ident("poll.status_id"), statusID).Scan(ctx)
		},
		statusID,
	)
}

func (p *pollDB) getPoll(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Poll) error, keyParts ...any) (*gtsmodel.Poll, error) {
	// Fetch poll from database cache with loader callback
	poll, err := p.state.Caches.GTS.Poll().Load(lookup, func() (*gtsmodel.Poll, error) {
		var poll gtsmodel.Poll

		// Not cached! Perform database query.
		if err := dbQuery(&poll); err != nil {
			return nil, err
		}

		return &poll, nil
	}, keyParts...)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return poll, nil
	}

	// Further populate the poll fields where applicable.
	if err := p.PopulatePoll(ctx, poll); err != nil {
		return nil, err
	}

	return poll, nil
}

func (p *pollDB) PopulatePoll(ctx context.Context, poll *gtsmodel.Poll) error {
	var (
		err  error
		errs gtserror.MultiError
	)

	if poll.Status == nil {
		// Vote account is not set, fetch from database.
		poll.Status, err = p.state.DB.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			poll.StatusID,
		)
		if err != nil {
			errs.Appendf("error populating poll status: %w", err)
		}
	}

	return errs.Combine()
}

func (p *pollDB) PutPoll(ctx context.Context, poll *gtsmodel.Poll) error {
	return p.state.Caches.GTS.Poll().Store(poll, func() error {
		_, err := p.db.NewInsert().Model(poll).Exec(ctx)
		return err
	})
}

func (p *pollDB) DeletePollByID(ctx context.Context, id string) error {
	// Delete poll by ID from database.
	if _, err := p.db.NewDelete().
		Table("polls").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx); err != nil {
		return err
	}

	// Invalidate poll by ID from cache.
	p.state.Caches.GTS.Poll().Invalidate("ID", id)

	return nil
}

func (p *pollDB) GetPollVoteByID(ctx context.Context, id string) (*gtsmodel.PollVote, error) {
	return p.getPollVote(
		ctx,
		"ID",
		func(vote *gtsmodel.PollVote) error {
			return p.db.NewSelect().Model(vote).Where("? = ?", bun.Ident("poll_vote.id"), id).Scan(ctx)
		},
		id,
	)
}

func (p *pollDB) getPollVote(ctx context.Context, lookup string, dbQuery func(*gtsmodel.PollVote) error, keyParts ...any) (*gtsmodel.PollVote, error) {
	// Fetch vote from database cache with loader callback
	vote, err := p.state.Caches.GTS.PollVote().Load(lookup, func() (*gtsmodel.PollVote, error) {
		var vote gtsmodel.PollVote

		// Not cached! Perform database query.
		if err := dbQuery(&vote); err != nil {
			return nil, err
		}

		return &vote, nil
	}, keyParts...)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return vote, nil
	}

	// Further populate the vote fields where applicable.
	if err := p.PopulatePollVote(ctx, vote); err != nil {
		return nil, err
	}

	return vote, nil
}

func (p *pollDB) PopulatePollVote(ctx context.Context, vote *gtsmodel.PollVote) error {
	var (
		err  error
		errs gtserror.MultiError
	)

	if vote.Account == nil {
		// Vote account is not set, fetch from database.
		vote.Account, err = p.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			vote.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating vote account: %w", err)
		}
	}

	if vote.Poll == nil {
		// Vote poll is not set, fetch from database.
		vote.Poll, err = p.GetPollByID(
			gtscontext.SetBarebones(ctx),
			vote.PollID,
		)
		if err != nil {
			errs.Appendf("error populating vote poll: %w", err)
		}
	}

	return errs.Combine()
}

func (p *pollDB) GetPollVotes(ctx context.Context, pollID string) (map[string][]*gtsmodel.PollVote, error) {
	// Get all account IDs of those who voted in poll.
	accountIDs, err := p.getPollVoterIDs(ctx, pollID)
	if err != nil {
		return nil, err
	}

	// Fetch slices of poll votes made by each account ID.
	votes := make(map[string][]*gtsmodel.PollVote, len(accountIDs))
	for _, id := range accountIDs {
		votesBy, err := p.GetPollVotesBy(ctx, pollID, id)
		if err != nil {
			return nil, gtserror.Newf("error getting votes by %s in %s: %w", id, pollID, err)
		}
		votes[id] = votesBy
	}

	return votes, nil
}

func (p *pollDB) GetPollVotesBy(ctx context.Context, pollID string, accountID string) ([]*gtsmodel.PollVote, error) {
	// Get all vote IDs by the given account in given poll.
	voteIDs, err := p.getPollVoteIDs(ctx, pollID, accountID)
	if err != nil {
		return nil, err
	}

	// Fetch poll vote models for all fetched vote IDs.
	votes := make([]*gtsmodel.PollVote, 0, len(voteIDs))
	for _, id := range voteIDs {
		vote, err := p.GetPollVoteByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting vote with id %s: %v", id, err)
			continue
		}
		votes = append(votes, vote)
	}

	return votes, nil
}

func (p *pollDB) PutPollVotes(ctx context.Context, votes ...*gtsmodel.PollVote) error {
	// Insert all votes into DB in transaction.
	err := p.db.RunInTx(ctx, func(tx Tx) error {
		for _, vote := range votes {
			if _, err := p.db.NewInsert().
				Model(vote).
				Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, vote := range votes {
		// Only cache votes on confirmed insert, with no-op store func.
		_ = p.state.Caches.GTS.PollVote().Store(vote, noopStore)
	}

	return nil
}

func (p *pollDB) DeletePollVotes(ctx context.Context, pollID string) error {
	// Get all account IDs of those who voted in poll.
	accountIDs, err := p.getPollVoterIDs(ctx, pollID)
	if err != nil {
		return err
	}

	for _, id := range accountIDs {
		// Delete all votes by this account in each of the polls,
		// this way ensures that all necessary caches are invalidated.
		if err := p.DeletePollVotesBy(ctx, pollID, id); err != nil {
			log.Errorf(ctx, "error deleting votes by %s in %s: %v", id, pollID, err)
		}
	}

	return nil
}

func (p *pollDB) DeletePollVotesBy(ctx context.Context, pollID string, accountID string) error {
	var voteIDs []string

	// Delete all votes in poll by account,
	// returning the IDs of all the votes.
	if err := p.db.NewDelete().
		Table("poll_votes").
		Where("? = ?", bun.Ident("poll_id"), pollID).
		Where("? = ?", bun.Ident("account_id"), accountID).
		Returning("id").
		Scan(ctx, &voteIDs); err != nil {
		return err
	}

	// Invalidate cache of all voters in this poll.
	p.state.Caches.GTS.PollVoterIDs().Invalidate(pollID)

	// Invalidate cache of all votes by this account in this poll.
	p.state.Caches.GTS.PollVoteIDs().Invalidate(pollID + "." + accountID)

	for _, id := range voteIDs {
		// Invalidate all poll votes with ID from cache.
		p.state.Caches.GTS.PollVote().Invalidate(id)
	}

	return nil
}

func (p *pollDB) DeletePollVotesByAccountID(ctx context.Context, accountID string) error {
	var pollIDs []string

	// Select all polls this account
	// has registered a poll vote in.
	if err := p.db.NewSelect().
		Table("poll_votes").
		Column("poll_id").
		Where("? = ?", bun.Ident("account_id"), accountID).
		Scan(ctx, &pollIDs); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	for _, id := range pollIDs {
		// Delete all votes by this account in each of the polls,
		// this way ensures that all necessary caches are invalidated.
		if err := p.DeletePollVotesBy(ctx, id, accountID); err != nil {
			log.Errorf(ctx, "error deleting votes by %s in %s: %v", accountID, id, err)
		}
	}

	return nil
}

func (p *pollDB) getPollVoterIDs(ctx context.Context, pollID string) ([]string, error) {
	return p.state.Caches.GTS.PollVoterIDs().Load(pollID, func() ([]string, error) {
		var accountIDs []string

		// Voter account IDs not in cache, perform DB query!
		q := newSelectPollVoters(p.db, pollID)
		if _, err := q.Exec(ctx, &accountIDs); err != nil {
			return nil, err
		}

		return accountIDs, nil
	})
}

func (p *pollDB) getPollVoteIDs(ctx context.Context, pollID string, accountID string) ([]string, error) {
	return p.state.Caches.GTS.PollVoteIDs().Load(pollID+"."+accountID, func() ([]string, error) {
		var voteIDs []string

		// Poll vote IDs not in cache, perform DB query!
		q := newSelectPollVotes(p.db, pollID, accountID)
		if _, err := q.Exec(ctx, &voteIDs); err != nil {
			return nil, err
		}

		return voteIDs, nil
	})
}

// newSelectFollowers returns a new select query for all id column values in the poll votes table with poll_id = pollID and account_id = accountID.
func newSelectPollVotes(db *DB, pollID string, accountID string) *bun.SelectQuery {
	return db.NewSelect().
		Table("poll_votes").
		Column("id").
		Where("? = ?", bun.Ident("poll_id"), pollID).
		Where("? = ?", bun.Ident("account_id"), accountID).
		OrderExpr("? DESC", bun.Ident("id"))
}

// newSelectFollowers returns a new select query for all account_id column values in the poll votes table with poll_id = pollID.
func newSelectPollVoters(db *DB, pollID string) *bun.SelectQuery {
	return db.NewSelect().
		Table("poll_votes").
		Column("account_id").
		Where("? = ?", bun.Ident("poll_id"), pollID).
		OrderExpr("? DESC", bun.Ident("id"))
}

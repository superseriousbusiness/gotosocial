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
	"slices"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

type pollDB struct {
	db    *bun.DB
	state *state.State
}

func (p *pollDB) GetPollByID(ctx context.Context, id string) (*gtsmodel.Poll, error) {
	return p.getPoll(
		ctx,
		"ID",
		func(poll *gtsmodel.Poll) error {
			return p.db.NewSelect().
				Model(poll).
				Where("? = ?", bun.Ident("poll.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (p *pollDB) getPoll(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Poll) error, keyParts ...any) (*gtsmodel.Poll, error) {
	// Fetch poll from database cache with loader callback
	poll, err := p.state.Caches.DB.Poll.LoadOne(lookup, func() (*gtsmodel.Poll, error) {
		var poll gtsmodel.Poll

		// Not cached! Perform database query.
		if err := dbQuery(&poll); err != nil {
			return nil, err
		}

		// Ensure vote slice
		// is non nil and set.
		poll.CheckVotes()

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

func (p *pollDB) GetOpenPolls(ctx context.Context) ([]*gtsmodel.Poll, error) {
	var pollIDs []string

	// Select all polls with:
	// - UNSET `closed_at`
	// - SET   `expires_at`
	if err := p.db.NewSelect().
		Table("polls").
		Column("polls.id").
		Join("JOIN ? ON ? = ?", bun.Ident("statuses"), bun.Ident("polls.id"), bun.Ident("statuses.poll_id")).
		Where("? = true", bun.Ident("statuses.local")).
		Where("? IS NOT NULL", bun.Ident("polls.expires_at")).
		Where("? IS NULL", bun.Ident("polls.closed_at")).
		Scan(ctx, &pollIDs); err != nil {
		return nil, err
	}

	// Preallocate a slice to contain the poll models.
	polls := make([]*gtsmodel.Poll, 0, len(pollIDs))

	for _, id := range pollIDs {
		// Attempt to fetch poll from DB.
		poll, err := p.GetPollByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting poll %s: %v", id, err)
			continue
		}

		// Append poll to return slice.
		polls = append(polls, poll)
	}

	return polls, nil
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
	// Ensure vote slice
	// is non nil and set.
	poll.CheckVotes()

	return p.state.Caches.DB.Poll.Store(poll, func() error {
		_, err := p.db.NewInsert().Model(poll).Exec(ctx)
		return err
	})
}

func (p *pollDB) UpdatePoll(ctx context.Context, poll *gtsmodel.Poll, cols ...string) error {
	// Ensure vote slice
	// is non nil and set.
	poll.CheckVotes()

	return p.state.Caches.DB.Poll.Store(poll, func() error {
		return p.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			_, err := tx.NewUpdate().
				Model(poll).
				Column(cols...).
				Where("? = ?", bun.Ident("id"), poll.ID).
				Exec(ctx)
			return err
		})
	})
}

func (p *pollDB) DeletePollByID(ctx context.Context, id string) error {
	// Delete poll vote with ID, and its associated votes from the database.
	if err := p.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

		// Delete poll from database.
		if _, err := tx.NewDelete().
			Table("polls").
			Where("? = ?", bun.Ident("id"), id).
			Exec(ctx); err != nil {
			return err
		}

		// Delete the poll votes.
		_, err := tx.NewDelete().
			Table("poll_votes").
			Where("? = ?", bun.Ident("poll_id"), id).
			Exec(ctx)
		return err
	}); err != nil {
		return err
	}

	// Wrap provided ID in a poll
	// model for calling cache hook.
	var deleted gtsmodel.Poll
	deleted.ID = id

	// Invalidate cached poll with ID, manually
	// call invalidate hook in case not cached.
	p.state.Caches.DB.Poll.Invalidate("ID", id)
	p.state.Caches.OnInvalidatePoll(&deleted)

	return nil
}

func (p *pollDB) GetPollVoteByID(ctx context.Context, id string) (*gtsmodel.PollVote, error) {
	return p.getPollVote(
		ctx,
		"ID",
		func(vote *gtsmodel.PollVote) error {
			return p.db.NewSelect().
				Model(vote).
				Where("? = ?", bun.Ident("poll_vote.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (p *pollDB) GetPollVoteBy(ctx context.Context, pollID string, accountID string) (*gtsmodel.PollVote, error) {
	return p.getPollVote(
		ctx,
		"PollID,AccountID",
		func(vote *gtsmodel.PollVote) error {
			return p.db.NewSelect().
				Model(vote).
				Where("? = ?", bun.Ident("poll_vote.account_id"), accountID).
				Where("? = ?", bun.Ident("poll_vote.poll_id"), pollID).
				Scan(ctx)
		},
		pollID,
		accountID,
	)
}

func (p *pollDB) getPollVote(ctx context.Context, lookup string, dbQuery func(*gtsmodel.PollVote) error, keyParts ...any) (*gtsmodel.PollVote, error) {
	// Fetch vote from database cache with loader callback
	vote, err := p.state.Caches.DB.PollVote.LoadOne(lookup, func() (*gtsmodel.PollVote, error) {
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

func (p *pollDB) GetPollVotes(ctx context.Context, pollID string) ([]*gtsmodel.PollVote, error) {

	// Load vote IDs known for given poll ID using loader callback.
	voteIDs, err := p.state.Caches.DB.PollVoteIDs.Load(pollID, func() ([]string, error) {
		var voteIDs []string

		// Vote IDs not in cache, perform DB query!
		q := newSelectPollVotes(p.db, pollID)
		if _, err := q.Exec(ctx, &voteIDs); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return voteIDs, nil
	})
	if err != nil {
		return nil, err
	}

	// Load all votes from IDs via cache loader callbacks.
	votes, err := p.state.Caches.DB.PollVote.LoadIDs("ID",
		voteIDs,
		func(uncached []string) ([]*gtsmodel.PollVote, error) {
			// Preallocate expected length of uncached votes.
			votes := make([]*gtsmodel.PollVote, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) IDs.
			if err := p.db.NewSelect().
				Model(&votes).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return votes, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the poll votes by their
	// IDs to ensure in correct order.
	getID := func(v *gtsmodel.PollVote) string { return v.ID }
	xslices.OrderBy(votes, voteIDs, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return votes, nil
	}

	// Populate all loaded votes, removing those we fail to
	// populate (removes needing so many nil checks everywhere).
	votes = slices.DeleteFunc(votes, func(vote *gtsmodel.PollVote) bool {
		if err := p.PopulatePollVote(ctx, vote); err != nil {
			log.Errorf(ctx, "error populating vote %s: %v", vote.ID, err)
			return true
		}
		return false
	})

	return votes, nil
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

func (p *pollDB) PutPollVote(ctx context.Context, vote *gtsmodel.PollVote) error {
	return p.state.Caches.DB.PollVote.Store(vote, func() error {
		return p.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Try insert vote into database.
			if _, err := tx.NewInsert().
				Model(vote).
				Exec(ctx); err != nil {
				return err
			}

			var poll gtsmodel.Poll

			// Select current poll counts from DB,
			// taking minimal columns needed to
			// increment/decrement votes.
			if err := tx.NewSelect().
				Model(&poll).
				Column("options", "votes", "voters").
				Where("? = ?", bun.Ident("id"), vote.PollID).
				Scan(ctx); err != nil {
				return err
			}

			// Increment vote choices and voters count.
			poll.IncrementVotes(vote.Choices, true)

			// Finally, update the poll entry.
			_, err := tx.NewUpdate().
				Model(&poll).
				Column("votes", "voters").
				Where("? = ?", bun.Ident("id"), vote.PollID).
				Exec(ctx)
			return err
		})
	})
}

func (p *pollDB) UpdatePollVote(ctx context.Context, vote *gtsmodel.PollVote, cols ...string) error {
	return p.state.Caches.DB.PollVote.Store(vote, func() error {
		_, err := p.db.NewUpdate().
			Model(vote).
			Column(cols...).
			Where("? = ?", bun.Ident("id"), vote.ID).
			Exec(ctx)
		return err
	})
}

func (p *pollDB) DeletePollVoteBy(ctx context.Context, pollID string, accountID string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted gtsmodel.PollVote
	deleted.AccountID = accountID
	deleted.PollID = pollID

	// Delete the poll vote with given poll and account IDs, and update vote counts.
	if err := p.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

		// Delete vote in poll by account,
		// returning deleted model info.
		switch _, err := tx.NewDelete().
			Model(&deleted).
			Where("? = ?", bun.Ident("poll_id"), pollID).
			Where("? = ?", bun.Ident("account_id"), accountID).
			Returning("?", bun.Ident("choices")).
			Exec(ctx); {

		case err == nil:
			// no issue
		case errors.Is(err, db.ErrNoEntries):
			return nil
		default:
			return err
		}

		// Update the votes for this deleted poll.
		err := updatePollCounts(ctx, tx, &deleted)
		return err
	}); err != nil {
		return err
	}

	// Invalidate the poll vote cache by given poll + account IDs, also
	// manually call invalidation hook in case not actually stored in cache.
	p.state.Caches.DB.PollVote.Invalidate("PollID,AccountID", pollID, accountID)
	p.state.Caches.OnInvalidatePollVote(&deleted)

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
		if err := p.DeletePollVoteBy(ctx, id, accountID); err != nil {
			log.Errorf(ctx, "error deleting vote by %s in %s: %v", accountID, id, err)
		}
	}

	return nil
}

// updatePollCounts updates the vote counts on a poll for the given deleted PollVote model.
func updatePollCounts(ctx context.Context, tx bun.Tx, deleted *gtsmodel.PollVote) error {

	// Select current poll counts from DB,
	// taking minimal columns needed to
	// increment/decrement votes.
	var poll gtsmodel.Poll
	switch err := tx.NewSelect().
		Model(&poll).
		Column("options", "votes", "voters").
		Where("? = ?", bun.Ident("id"), deleted.PollID).
		Scan(ctx); {

	case err == nil:
		// no issue.

	case errors.Is(err, db.ErrNoEntries):
		// no poll found,
		// return here.
		return nil

	default:
		// irrecoverable.
		return err
	}

	// Decrement vote choices and voters count.
	poll.DecrementVotes(deleted.Choices, true)

	// Finally, update the poll entry.
	if _, err := tx.NewUpdate().
		Model(&poll).
		Column("votes", "voters").
		Where("? = ?", bun.Ident("id"), deleted.PollID).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	return nil
}

// newSelectPollVotes returns a new select query for all rows in the poll_votes table with poll_id = pollID.
func newSelectPollVotes(db *bun.DB, pollID string) *bun.SelectQuery {
	return db.NewSelect().
		TableExpr("?", bun.Ident("poll_votes")).
		ColumnExpr("?", bun.Ident("id")).
		Where("? = ?", bun.Ident("poll_id"), pollID).
		OrderExpr("? DESC", bun.Ident("id"))
}

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

	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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
	err := p.db.RunInTx(ctx, func(tx Tx) error {
		// Delete poll from database.
		if _, err := p.db.NewDelete().
			Table("polls").
			Where("? = ?", bun.Ident("id"), id).
			Exec(ctx); err != nil {
			return err
		}

		// Delete related poll votes from database.
		if _, err := p.db.NewDelete().
			Table("poll_votes").
			Where("? = ?", bun.Ident("poll_id"), id).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	})

	// Invalidate poll vote from cache by ID.
	p.state.Caches.GTS.PollVote().Invalidate("ID", id)

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

func (p *pollDB) GetPollVotes(ctx context.Context, statusID string, accountID string) ([]*gtsmodel.PollVote, error) {
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

func (p *pollDB) PutPollVote(ctx context.Context, vote *gtsmodel.PollVote) error {
	return p.state.Caches.GTS.PollVote().Store(vote, func() error {
		_, err := p.db.NewInsert().Model(vote).Exec(ctx)
		return err
	})
}

func (p *pollDB) DeletePollVoteByID(ctx context.Context, id string) error {
	// Delete vote from database.
	if _, err := p.db.NewDelete().
		Table("poll_votes").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx); err != nil {
		return err
	}

	// Invalidate poll vote from cache by ID.
	p.state.Caches.GTS.PollVote().Invalidate("ID", id)

	return nil
}

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

package db

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type Poll interface {
	// GetPollByID fetches the Poll with given ID from the database.
	GetPollByID(ctx context.Context, id string) (*gtsmodel.Poll, error)

	// GetOpenPolls fetches all local Polls in the database with an unset `closed_at` column.
	GetOpenPolls(ctx context.Context) ([]*gtsmodel.Poll, error)

	// PopulatePoll ensures the given Poll is fully populated with all other related database models.
	PopulatePoll(ctx context.Context, poll *gtsmodel.Poll) error

	// PutPoll puts the given Poll in the database.
	PutPoll(ctx context.Context, poll *gtsmodel.Poll) error

	// UpdatePoll updates the Poll in the database, only on selected columns if provided (else, all).
	UpdatePoll(ctx context.Context, poll *gtsmodel.Poll, cols ...string) error

	// DeletePollByID deletes the Poll with given ID from the
	// database, along with all its associated poll votes.
	DeletePollByID(ctx context.Context, id string) error

	// GetPollVoteByID gets the PollVote with given ID from the database.
	GetPollVoteByID(ctx context.Context, id string) (*gtsmodel.PollVote, error)

	// GetPollVotesBy fetches the PollVote in Poll with ID, by account ID, from the database.
	GetPollVoteBy(ctx context.Context, pollID string, accountID string) (*gtsmodel.PollVote, error)

	// GetPollVotes fetches all PollVotes in Poll with ID, from the database.
	GetPollVotes(ctx context.Context, pollID string) ([]*gtsmodel.PollVote, error)

	// PopulatePollVote ensures the given PollVote is fully populated with all other related database models.
	PopulatePollVote(ctx context.Context, votes *gtsmodel.PollVote) error

	// PutPollVote puts the given PollVote in the database.
	PutPollVote(ctx context.Context, vote *gtsmodel.PollVote) error

	// UpdatePollVote updates the given poll vote in the database, using only given columns (else, all).
	UpdatePollVote(ctx context.Context, vote *gtsmodel.PollVote, cols ...string) error

	// DeletePollVoteBy deletes the PollVote in Poll with ID, by account ID, from the database.
	DeletePollVoteBy(ctx context.Context, pollID string, accountID string) error

	// DeletePollVotesByAccountID deletes all PollVotes in all Polls, by account ID, from the database.
	DeletePollVotesByAccountID(ctx context.Context, accountID string) error
}

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
	// GetPollByID ...
	GetPollByID(ctx context.Context, id string) (*gtsmodel.Poll, error)

	// GetPollByStatusID ...
	GetPollByStatusID(ctx context.Context, statusID string) (*gtsmodel.Poll, error)

	// PutPoll ...
	PutPoll(ctx context.Context, poll *gtsmodel.Poll) error

	// DeletePollByID ...
	DeletePollByID(ctx context.Context, id string) error

	// GetPollVoteByID ...
	GetPollVoteByID(ctx context.Context, id string) (*gtsmodel.PollVote, error)

	// GetPollVotes ...
	GetPollVotes(ctx context.Context, pollID string) (map[string][]*gtsmodel.PollVote, error)

	// GetPollVotesBy ...
	GetPollVotesBy(ctx context.Context, pollID string, accountID string) ([]*gtsmodel.PollVote, error)

	// PutPollVote ...
	PutPollVote(ctx context.Context, vote *gtsmodel.PollVote) error

	// DeletePollVotes ...
	DeletePollVotes(ctx context.Context, pollID string) error

	// DeletePollVotesBy ...
	DeletePollVotesBy(ctx context.Context, pollID string, accountID string) error

	// DeletePollVotesByAccountID ...
	DeletePollVotesByAccountID(ctx context.Context, accountID string) error
}

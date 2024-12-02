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

package gtsmodel

import (
	"time"
)

// Poll represents an attached (to) Status poll, i.e. a questionaire. Can be remote / local.
type Poll struct {
	ID         string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"` // Unique identity string.
	Multiple   *bool     `bun:",nullzero,notnull,default:false"`          // Is this a multiple choice poll? i.e. can you vote on multiple options.
	HideCounts *bool     `bun:",nullzero,notnull,default:false"`          // Hides vote counts until poll ends.
	Options    []string  `bun:",nullzero,notnull"`                        // The available options for this poll.
	Votes      []int     `bun:",nullzero,notnull"`                        // Vote counts per choice.
	Voters     *int      `bun:",nullzero,notnull"`                        // Total no. voters count.
	StatusID   string    `bun:"type:CHAR(26),nullzero,notnull,unique"`    // Status ID of which this Poll is attached to.
	ExpiresAt  time.Time `bun:"type:timestamptz,nullzero,notnull"`        // The expiry date of this Poll.
	ClosedAt   time.Time `bun:"type:timestamptz,nullzero"`                // The closure date of this poll, will be zerotime until set.
	Closing    bool      `bun:"-"`                                        // An ephemeral field only set on Polls in the middle of closing.
	// no creation date, use attached Status.CreatedAt.
}

// PollVote represents a single instance of vote(s) in a Poll by an account.
// If the Poll is single-choice, len(.Choices) = 1, if multiple-choice then
// len(.Choices) >= 1. Can be remote or local.
type PollVote struct {
	ID        string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // Unique identity string.
	Choices   []int     `bun:",nullzero,notnull"`                                           // The Poll's option indices of which these are votes for.
	AccountID string    `bun:"type:CHAR(26),nullzero,notnull,unique:in_poll_by_account"`    // Account ID from which this vote originated.
	PollID    string    `bun:"type:CHAR(26),nullzero,notnull,unique:in_poll_by_account"`    // Poll ID of which this is a vote in.
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // The creation date of this PollVote.
}

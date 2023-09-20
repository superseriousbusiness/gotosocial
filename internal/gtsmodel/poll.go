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

import "time"

// Poll ...
type Poll struct {
	ID         string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // Unique identity string.
	Multiple   *bool     `bun:"nullzero,notnull,default:false"`                              // Is this a multiple choice poll? i.e. can you vote on multiple options.
	HideCounts *bool     `bun:"nullzero,notnull,default:false"`                              // Hides vote counts until poll ends.
	Options    []string  `bun:"nullzero,notnull"`                                            // The available options for this poll.
	StatusID   string    `bun:"type:CHAR(26),nullzero,notnull,unique"`                       // Status ID of which this Poll is attached to.
	Status     *Status   `bun:"-"`                                                           // The related Status for StatusID (not always set).
	CreatedAt  time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // The creation date of this Poll.
	ExpiresAt  time.Time `bun:"type:timestamptz,nullzero,notnull"`                           // The expiry date of this Poll.
}

// Expired returns whether the given poll is expired.
func (p *Poll) Expired() bool {
	return time.Now().After(p.ExpiresAt)
}

// PollVote ...
type PollVote struct {
	ID        string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // Unique identity string.
	Choice    int       `bun:"nullzero,notnull"`                                            // The Poll's option index of which this is a vote for.
	AccountID string    `bun:"type:CHAR(26),nullzero,notnull"`                              // Account ID from which this vote originated.
	Account   *Account  `bun:"-"`                                                           // The related Account for AccountID (not always set).
	PollID    string    `bun:"type:CHAR(26),nullzero,notnull"`                              // Poll ID of which this is a vote in.
	Poll      *Poll     `bun:"-"`                                                           // The related Poll for PollID (not always set).
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // The creation date of this PollVote.
}

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
	"strings"
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
	Status     *Status   `bun:"-"`                                        // The related Status for StatusID (not always set).
	ExpiresAt  time.Time `bun:"type:timestamptz,nullzero"`                // The expiry date of this Poll, will be zerotime until set. (local polls ALWAYS have this set).
	ClosedAt   time.Time `bun:"type:timestamptz,nullzero"`                // The closure date of this poll, anything other than zerotime indicates closed.
	Closing    bool      `bun:"-"`                                        // An ephemeral field only set on Polls in the middle of closing.
	// no creation date, use attached Status.CreatedAt.
}

// GetChoice returns the option index with name.
func (p *Poll) GetChoice(name string) int {
	for i, option := range p.Options {
		if strings.EqualFold(option, name) {
			return i
		}
	}
	return -1
}

// Expired returns whether the Poll is expired (i.e. date is BEFORE now).
func (p *Poll) Expired() bool {
	return !p.ExpiresAt.IsZero() &&
		time.Now().After(p.ExpiresAt)
}

// Closed returns whether the Poll is closed (i.e. date is set and BEFORE now).
func (p *Poll) Closed() bool {
	return !p.ClosedAt.IsZero() &&
		time.Now().After(p.ClosedAt)
}

// IncrementVotes increments Poll vote counts for given choices, and voters if 'isNew' is set.
func (p *Poll) IncrementVotes(choices []int, isNew bool) {
	if len(choices) == 0 {
		return
	}
	p.CheckVotes()
	for _, choice := range choices {
		p.Votes[choice]++
	}
	if isNew {
		(*p.Voters)++
	}
}

// DecrementVotes decrements Poll vote counts for given choices, and voters if 'withVoter' is set.
func (p *Poll) DecrementVotes(choices []int, withVoter bool) {
	if len(choices) == 0 {
		return
	}
	p.CheckVotes()
	for _, choice := range choices {
		if p.Votes[choice] != 0 {
			p.Votes[choice]--
		}
	}
	if (*p.Voters) != 0 &&
		withVoter {
		(*p.Voters)--
	}
}

// ResetVotes resets all stored vote counts.
func (p *Poll) ResetVotes() {
	p.Votes = make([]int, len(p.Options))
	p.Voters = new(int)
}

// CheckVotes ensures that the Poll.Votes slice is not nil,
// else initializing an int slice len+cap equal to Poll.Options.
// Note this should not be needed anywhere other than the
// database and the processor.
func (p *Poll) CheckVotes() {
	if p.Votes == nil {
		p.Votes = make([]int, len(p.Options))
	}
	if p.Voters == nil {
		p.Voters = new(int)
	}
}

// PollVote represents a single instance of vote(s) in a Poll by an account.
// If the Poll is single-choice, len(.Choices) = 1, if multiple-choice then
// len(.Choices) >= 1. Can be remote or local.
type PollVote struct {
	ID        string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // Unique identity string.
	Choices   []int     `bun:",nullzero,notnull"`                                           // The Poll's option indices of which these are votes for.
	AccountID string    `bun:"type:CHAR(26),nullzero,notnull,unique:in_poll_by_account"`    // Account ID from which this vote originated.
	Account   *Account  `bun:"-"`                                                           // The related Account for AccountID (not always set).
	PollID    string    `bun:"type:CHAR(26),nullzero,notnull,unique:in_poll_by_account"`    // Poll ID of which this is a vote in.
	Poll      *Poll     `bun:"-"`                                                           // The related Poll for PollID (not always set).
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // The creation date of this PollVote.
}

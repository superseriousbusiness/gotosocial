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

package model

import "code.superseriousbusiness.org/gotosocial/internal/language"

// Poll represents a poll attached to a status.
//
// swagger:model poll
type Poll struct {
	// The ID of the poll in the database.
	// example: 01FBYKMD1KBMJ0W6JF1YZ3VY5D
	ID string `json:"id"`

	// When the poll ends. (ISO 8601 Datetime).
	ExpiresAt *string `json:"expires_at"`

	// Is the poll currently expired?
	Expired bool `json:"expired"`

	// Does the poll allow multiple-choice answers?
	Multiple bool `json:"multiple"`

	// How many votes have been received.
	VotesCount int `json:"votes_count"`

	// How many unique accounts have voted on a multiple-choice poll.
	VotersCount *int `json:"voters_count"`

	// When called with a user token, has the authorized user voted?
	//
	// Omitted when no user token provided.
	Voted *bool `json:"voted,omitempty"`

	// When called with a user token, which options has the authorized
	// user chosen? Contains an array of index values for options.
	//
	// Omitted when no user token provided.
	OwnVotes *[]int `json:"own_votes,omitempty"`

	// Possible answers for the poll.
	Options []PollOption `json:"options"`

	// Custom emoji to be used for rendering poll options.
	Emojis []Emoji `json:"emojis"`
}

// PollOption represents the current vote counts for different poll options.
//
// swagger:model pollOption
type PollOption struct {
	// The text value of the poll option. String.
	Title string `json:"title"`

	// The number of received votes for this option.
	VotesCount *int `json:"votes_count"`
}

// PollRequest models a request to create a poll.
//
// swagger:ignore
type PollRequest struct {
	// Array of possible answers.
	// If provided, media_ids cannot be used, and poll[expires_in] must be provided.
	// name: poll[options]
	Options []string `form:"poll[options][]" json:"options" xml:"options"`

	// Duration the poll should be open, in seconds.
	// If provided, media_ids cannot be used, and poll[options] must be provided.
	ExpiresIn int `form:"poll[expires_in]" xml:"expires_in"`

	// Duration the poll should be open, in seconds.
	// If provided, media_ids cannot be used, and poll[options] must be provided.
	ExpiresInI interface{} `json:"expires_in"`

	// Allow multiple choices on this poll.
	Multiple bool `form:"poll[multiple]" json:"multiple" xml:"multiple"`

	// Hide vote counts until the poll ends.
	HideTotals bool `form:"poll[hide_totals]" json:"hide_totals" xml:"hide_totals"`
}

// PollVoteRequest models a request to vote in a poll.
//
// swagger:ignore
type PollVoteRequest struct {
	// Choices contains poll vote choice indices.
	Choices []int `form:"choices[]" xml:"choices"`

	// ChoicesI contains poll vote choice
	// indices. Can be strings or integers.
	ChoicesI []interface{} `json:"choices"`
}

// WebPollOption models a template-ready poll option entry.
//
// swagger:ignore
type WebPollOption struct {
	PollOption

	// ID of the parent poll.
	PollID string

	// Emojis contained on parent poll.
	Emojis []Emoji

	// LanguageTag of parent status.
	LanguageTag *language.Language

	// Share of total votes as a percentage.
	VoteShare float32

	// String-formatted version of VoteShare.
	VoteShareStr string
}

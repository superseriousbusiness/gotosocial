/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package model

// Poll represents a poll attached to a status.
//
// swagger:model poll
type Poll struct {
	// The ID of the poll in the database.
	// example: 01FBYKMD1KBMJ0W6JF1YZ3VY5D
	ID string `json:"id"`
	// When the poll ends. (ISO 8601 Datetime), or null if the poll does not end
	ExpiresAt string `json:"expires_at,omitempty"`
	// Is the poll currently expired?
	Expired bool `json:"expired"`
	// Does the poll allow multiple-choice answers?
	Multiple bool `json:"multiple"`
	// How many votes have been received.
	VotesCount int `json:"votes_count"`
	// How many unique accounts have voted on a multiple-choice poll. Null if multiple is false.
	VotersCount int `json:"voters_count,omitempty"`
	// When called with a user token, has the authorized user voted?
	Voted bool `json:"voted,omitempty"`
	// When called with a user token, which options has the authorized user chosen? Contains an array of index values for options.
	OwnVotes []int `json:"own_votes,omitempty"`
	// Possible answers for the poll.
	Options []PollOptions `json:"options"`
	// Custom emoji to be used for rendering poll options.
	Emojis []Emoji `json:"emojis"`
}

// PollOptions represents the current vote counts for different poll options.
//
// swagger:model pollOptions
type PollOptions struct {
	// The text value of the poll option. String.
	Title string `json:"title"`
	// The number of received votes for this option.
	// Number, or null if results are not published yet.
	VotesCount int `json:"votes_count,omitempty"`
}

// PollRequest models a request to create a poll.
//
// swagger:parameters createStatus
type PollRequest struct {
	// Array of possible answers.
	// If provided, media_ids cannot be used, and poll[expires_in] must be provided.
	// name: poll[options]
	Options []string `form:"options" json:"options" xml:"options"`
	// Duration the poll should be open, in seconds.
	// If provided, media_ids cannot be used, and poll[options] must be provided.
	ExpiresIn int `form:"expires_in" json:"expires_in" xml:"expires_in"`
	// Allow multiple choices on this poll.
	Multiple bool `form:"multiple" json:"multiple" xml:"multiple"`
	// Hide vote counts until the poll ends.
	HideTotals bool `form:"hide_totals" json:"hide_totals" xml:"hide_totals"`
}

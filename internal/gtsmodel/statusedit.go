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

// StatusEdit represents a **historical** view of a Status
// after a received edit. The Status itself will always
// contain the latest up-to-date information.
//
// Note that stored status edits may not exactly match that
// of the origin server, they are a best-effort by receiver
// to store version history. There is no AP history endpoint.
type StatusEdit struct {
	ID                     string             `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // ID of this item in the database.
	Content                string             `bun:""`                                                            // Content of status at time of edit; likely html-formatted but not guaranteed.
	ContentWarning         string             `bun:",nullzero"`                                                   // Content warning of status at time of edit.
	Text                   string             `bun:""`                                                            // Original status text, without formatting, at time of edit.
	Language               string             `bun:",nullzero"`                                                   // Status language at time of edit.
	Sensitive              *bool              `bun:",nullzero,notnull,default:false"`                             // Status sensitive flag at time of edit.
	AttachmentIDs          []string           `bun:"attachments,array"`                                           // Database IDs of media attachments associated with status at time of edit.
	AttachmentDescriptions []string           `bun:",array"`                                                      // Previous media descriptions of media attachments associated with status at time of edit.
	Attachments            []*MediaAttachment `bun:"-"`                                                           // Media attachments relating to .AttachmentIDs field (not always populated).
	PollOptions            []string           `bun:",array"`                                                      // Poll options of status at time of edit, only set if status contains a poll.
	PollVotes              []int              `bun:",array"`                                                      // Poll vote count at time of status edit, only set if poll votes were reset.
	StatusID               string             `bun:"type:CHAR(26),nullzero,notnull"`                              // The originating status ID this is a historical edit of.
	CreatedAt              time.Time          `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // The creation time of this version of the status content (according to receiving server).

	// We don't bother having a *gtsmodel.Status model here
	// as the StatusEdit is always just attached to a Status,
	// so it doesn't need a self-reference back to it.
}

// AttachmentsPopulated returns whether media attachments
// are populated according to current AttachmentIDs.
func (e *StatusEdit) AttachmentsPopulated() bool {
	if len(e.AttachmentIDs) != len(e.Attachments) {
		// this is the quickest indicator.
		return false
	}
	for i, id := range e.AttachmentIDs {
		if e.Attachments[i].ID != id {
			return false
		}
	}
	return true
}

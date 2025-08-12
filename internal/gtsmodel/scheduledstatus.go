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

// ScheduledStatus represents a status that is scheduled to be published at given time by a local user.
type ScheduledStatus struct {
	ID                string              `bun:"type:CHAR(26),pk,nullzero,notnull,unique"` // id of this item in the database
	AccountID         string              `bun:"type:CHAR(26),nullzero,notnull"`           // which account scheduled this status
	Account           *Account            `bun:"-"`                                        // Account corresponding to AccountID
	ScheduledAt       time.Time           `bun:"type:timestamptz,nullzero,notnull"`        // time at which the status is scheduled
	Text              string              `bun:""`                                         // Text content of the status
	Poll              ScheduledStatusPoll `bun:",embed:poll_,notnull,nullzero"`            //
	MediaIDs          []string            `bun:"attachments,array"`                        // Database IDs of any media attachments associated with this status
	MediaAttachments  []*MediaAttachment  `bun:"-"`                                        // Attachments corresponding to media IDs
	Sensitive         *bool               `bun:",nullzero,notnull,default:false"`          // mark the status as sensitive?
	SpoilerText       string              `bun:""`                                         // Original text of the content warning without formatting
	Visibility        Visibility          `bun:",nullzero,notnull"`                        // visibility entry for this status
	InReplyToID       string              `bun:"type:CHAR(26),nullzero"`                   // id of the status this status replies to
	Language          string              `bun:",nullzero"`                                // what language is this status written in?
	ApplicationID     string              `bun:"type:CHAR(26),nullzero"`                   // Which application was used to create this status?
	Application       *Application        `bun:"-"`                                        //
	LocalOnly         *bool               `bun:",nullzero,notnull,default:false"`          // Whether the status is not federated
	ContentType       string              `bun:",nullzero"`                                // Content type used to process the original text of the status
	InteractionPolicy *InteractionPolicy  `bun:""`                                         // InteractionPolicy for this status. If null then the default InteractionPolicy should be assumed for this status's Visibility. Always null for boost wrappers.
	Idempotency       string              `bun:",nullzero"`                                // Currently unused
}

type ScheduledStatusPoll struct {
	Options    []string `bun:",nullzero,array"`                 // The available options for this poll.
	ExpiresIn  int      `bun:",nullzero"`                       // Duration the poll should be open, in seconds
	Multiple   *bool    `bun:",nullzero,notnull,default:false"` // Is this a multiple choice poll? i.e. can you vote on multiple options.
	HideTotals *bool    `bun:",nullzero,notnull,default:false"` // Hides vote counts until poll ends.
}

// AttachmentsPopulated returns whether media attachments
// are populated according to current AttachmentIDs.
func (s *ScheduledStatus) AttachmentsPopulated() bool {
	if len(s.MediaIDs) != len(s.MediaAttachments) {
		// this is the quickest indicator.
		return false
	}
	for i, id := range s.MediaIDs {
		if s.MediaAttachments[i].ID != id {
			return false
		}
	}
	return true
}

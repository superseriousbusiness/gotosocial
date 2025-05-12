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

	"code.superseriousbusiness.org/gotosocial/internal/id"
)

type Status struct {
	ID                       string             `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`
	CreatedAt                time.Time          `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
	UpdatedAt                time.Time          `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
	FetchedAt                time.Time          `bun:"type:timestamptz,nullzero"`
	PinnedAt                 time.Time          `bun:"type:timestamptz,nullzero"`
	URI                      string             `bun:",unique,nullzero,notnull"`
	URL                      string             `bun:",nullzero"`
	Content                  string             `bun:""`
	AttachmentIDs            []string           `bun:"attachments,array"`
	TagIDs                   []string           `bun:"tags,array"`
	MentionIDs               []string           `bun:"mentions,array"`
	EmojiIDs                 []string           `bun:"emojis,array"`
	Local                    *bool              `bun:",nullzero,notnull,default:false"`
	AccountID                string             `bun:"type:CHAR(26),nullzero,notnull"`
	AccountURI               string             `bun:",nullzero,notnull"`
	InReplyToID              string             `bun:"type:CHAR(26),nullzero"`
	InReplyToURI             string             `bun:",nullzero"`
	InReplyToAccountID       string             `bun:"type:CHAR(26),nullzero"`
	BoostOfID                string             `bun:"type:CHAR(26),nullzero"`
	BoostOfAccountID         string             `bun:"type:CHAR(26),nullzero"`
	ThreadID                 string             `bun:"type:CHAR(26),nullzero"`
	PollID                   string             `bun:"type:CHAR(26),nullzero"`
	ContentWarning           string             `bun:",nullzero"`
	Visibility               string             `bun:",nullzero,notnull"`
	Sensitive                *bool              `bun:",nullzero,notnull,default:false"`
	Language                 string             `bun:",nullzero"`
	CreatedWithApplicationID string             `bun:"type:CHAR(26),nullzero"`
	ActivityStreamsType      string             `bun:",nullzero,notnull"`
	Text                     string             `bun:""`
	Federated                *bool              `bun:",notnull"`
	InteractionPolicy        *InteractionPolicy `bun:""`
	PendingApproval          *bool              `bun:",nullzero,notnull,default:false"`
	ApprovedByURI            string             `bun:",nullzero"`
}

func StatusToInteractionReq(status *Status) *InteractionRequest {
	reqID := id.NewULIDFromTime(status.CreatedAt)

	var (
		targetID        string
		targetAccountID string
		interactionType InteractionType
	)

	if status.InReplyToID != "" {
		// It's a reply.
		targetID = status.InReplyToID
		targetAccountID = status.InReplyToAccountID
		interactionType = InteractionReply
	} else {
		// It's a boost.
		targetID = status.BoostOfID
		targetAccountID = status.BoostOfAccountID
		interactionType = InteractionAnnounce
	}

	return &InteractionRequest{
		ID:                   reqID,
		CreatedAt:            status.CreatedAt,
		StatusID:             targetID,
		TargetAccountID:      targetAccountID,
		InteractingAccountID: status.AccountID,
		InteractionURI:       status.URI,
		InteractionType:      interactionType,
	}
}

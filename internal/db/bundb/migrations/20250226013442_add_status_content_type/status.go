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

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Status represents a user-created 'post' or 'status' in the database, either remote or local
type Status struct {
	ID                       string                      `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt                time.Time                   `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	EditedAt                 time.Time                   `bun:"type:timestamptz,nullzero"`                                   // when this status was last edited (if set)
	FetchedAt                time.Time                   `bun:"type:timestamptz,nullzero"`                                   // when was item (remote) last fetched.
	PinnedAt                 time.Time                   `bun:"type:timestamptz,nullzero"`                                   // Status was pinned by owning account at this time.
	URI                      string                      `bun:",unique,nullzero,notnull"`                                    // activitypub URI of this status
	URL                      string                      `bun:",nullzero"`                                                   // web url for viewing this status
	Content                  string                      `bun:""`                                                            // content of this status; likely html-formatted but not guaranteed
	AttachmentIDs            []string                    `bun:"attachments,array"`                                           // Database IDs of any media attachments associated with this status
	Attachments              []*gtsmodel.MediaAttachment `bun:"attached_media,rel:has-many"`                                 // Attachments corresponding to attachmentIDs
	TagIDs                   []string                    `bun:"tags,array"`                                                  // Database IDs of any tags used in this status
	Tags                     []*gtsmodel.Tag             `bun:"attached_tags,m2m:status_to_tags"`                            // Tags corresponding to tagIDs. https://bun.uptrace.dev/guide/relations.html#many-to-many-relation
	MentionIDs               []string                    `bun:"mentions,array"`                                              // Database IDs of any mentions in this status
	Mentions                 []*gtsmodel.Mention         `bun:"attached_mentions,rel:has-many"`                              // Mentions corresponding to mentionIDs
	EmojiIDs                 []string                    `bun:"emojis,array"`                                                // Database IDs of any emojis used in this status
	Emojis                   []*gtsmodel.Emoji           `bun:"attached_emojis,m2m:status_to_emojis"`                        // Emojis corresponding to emojiIDs. https://bun.uptrace.dev/guide/relations.html#many-to-many-relation
	Local                    *bool                       `bun:",nullzero,notnull,default:false"`                             // is this status from a local account?
	AccountID                string                      `bun:"type:CHAR(26),nullzero,notnull"`                              // which account posted this status?
	Account                  *gtsmodel.Account           `bun:"rel:belongs-to"`                                              // account corresponding to accountID
	AccountURI               string                      `bun:",nullzero,notnull"`                                           // activitypub uri of the owner of this status
	InReplyToID              string                      `bun:"type:CHAR(26),nullzero"`                                      // id of the status this status replies to
	InReplyToURI             string                      `bun:",nullzero"`                                                   // activitypub uri of the status this status is a reply to
	InReplyToAccountID       string                      `bun:"type:CHAR(26),nullzero"`                                      // id of the account that this status replies to
	InReplyTo                *Status                     `bun:"-"`                                                           // status corresponding to inReplyToID
	InReplyToAccount         *gtsmodel.Account           `bun:"rel:belongs-to"`                                              // account corresponding to inReplyToAccountID
	BoostOfID                string                      `bun:"type:CHAR(26),nullzero"`                                      // id of the status this status is a boost of
	BoostOfURI               string                      `bun:"-"`                                                           // URI of the status this status is a boost of; field not inserted in the db, just for dereferencing purposes.
	BoostOfAccountID         string                      `bun:"type:CHAR(26),nullzero"`                                      // id of the account that owns the boosted status
	BoostOf                  *Status                     `bun:"-"`                                                           // status that corresponds to boostOfID
	BoostOfAccount           *gtsmodel.Account           `bun:"rel:belongs-to"`                                              // account that corresponds to boostOfAccountID
	ThreadID                 string                      `bun:"type:CHAR(26),nullzero"`                                      // id of the thread to which this status belongs; only set for remote statuses if a local account is involved at some point in the thread, otherwise null
	EditIDs                  []string                    `bun:"edits,array"`                                                 //
	Edits                    []*gtsmodel.StatusEdit      `bun:"-"`                                                           //
	PollID                   string                      `bun:"type:CHAR(26),nullzero"`                                      //
	Poll                     *gtsmodel.Poll              `bun:"-"`                                                           //
	ContentWarning           string                      `bun:",nullzero"`                                                   // cw string for this status
	Visibility               gtsmodel.Visibility         `bun:",nullzero,notnull"`                                           // visibility entry for this status
	Sensitive                *bool                       `bun:",nullzero,notnull,default:false"`                             // mark the status as sensitive?
	Language                 string                      `bun:",nullzero"`                                                   // what language is this status written in?
	CreatedWithApplicationID string                      `bun:"type:CHAR(26),nullzero"`                                      // Which application was used to create this status?
	CreatedWithApplication   *gtsmodel.Application       `bun:"rel:belongs-to"`                                              // application corresponding to createdWithApplicationID
	ActivityStreamsType      string                      `bun:",nullzero,notnull"`                                           // What is the activitystreams type of this status? See: https://www.w3.org/TR/activitystreams-vocabulary/#object-types. Will probably almost always be Note but who knows!.
	Text                     string                      `bun:""`                                                            // Original text of the status without formatting
	ContentType              StatusContentType           `bun:",nullzero"`                                                   // Content type used to process the original text of the status
	Federated                *bool                       `bun:",notnull"`                                                    // This status will be federated beyond the local timeline(s)
	InteractionPolicy        *gtsmodel.InteractionPolicy `bun:""`                                                            // InteractionPolicy for this status. If null then the default InteractionPolicy should be assumed for this status's Visibility. Always null for boost wrappers.
	PendingApproval          *bool                       `bun:",nullzero,notnull,default:false"`                             // If true then status is a reply or boost wrapper that must be Approved by the reply-ee or boost-ee before being fully distributed.
	PreApproved              bool                        `bun:"-"`                                                           // If true, then status is a reply to or boost wrapper of a status on our instance, has permission to do the interaction, and an Accept should be sent out for it immediately. Field not stored in the DB.
	ApprovedByURI            string                      `bun:",nullzero"`                                                   // URI of an Accept Activity that approves the Announce or Create Activity that this status was/will be attached to.
}

type enumType int16

// StatusContentType is the content type with which a status's text is
// parsed. Can be either plain or markdown. Empty will default to plain.
type StatusContentType enumType

const (
	StatusContentTypePlain    StatusContentType = 1
	StatusContentTypeMarkdown StatusContentType = 2
	StatusContentTypeDefault                    = StatusContentTypePlain
)

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

	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// Status represents a user-created 'post' or 'status' in the database, either remote or local
type Status struct {
	ID                       string             `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                              // id of this item in the database
	CreatedAt                time.Time          `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                       // when was item created
	UpdatedAt                time.Time          `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                       // when was item last updated
	PinnedAt                 time.Time          `validate:"-" bun:"type:timestamptz,nullzero"`                                                         // Status was pinned by owning account at this time.
	URI                      string             `validate:"required,url" bun:",unique,nullzero,notnull"`                                               // activitypub URI of this status
	URL                      string             `validate:"url" bun:",nullzero"`                                                                       // web url for viewing this status
	Content                  string             `validate:"-" bun:""`                                                                                  // content of this status; likely html-formatted but not guaranteed
	AttachmentIDs            []string           `validate:"dive,ulid" bun:"attachments,array"`                                                         // Database IDs of any media attachments associated with this status
	Attachments              []*MediaAttachment `validate:"-" bun:"attached_media,rel:has-many"`                                                       // Attachments corresponding to attachmentIDs
	TagIDs                   []string           `validate:"dive,ulid" bun:"tags,array"`                                                                // Database IDs of any tags used in this status
	Tags                     []*Tag             `validate:"-" bun:"attached_tags,m2m:status_to_tags"`                                                  // Tags corresponding to tagIDs. https://bun.uptrace.dev/guide/relations.html#many-to-many-relation
	MentionIDs               []string           `validate:"dive,ulid" bun:"mentions,array"`                                                            // Database IDs of any mentions in this status
	Mentions                 []*Mention         `validate:"-" bun:"attached_mentions,rel:has-many"`                                                    // Mentions corresponding to mentionIDs
	EmojiIDs                 []string           `validate:"dive,ulid" bun:"emojis,array"`                                                              // Database IDs of any emojis used in this status
	Emojis                   []*Emoji           `validate:"-" bun:"attached_emojis,m2m:status_to_emojis"`                                              // Emojis corresponding to emojiIDs. https://bun.uptrace.dev/guide/relations.html#many-to-many-relation
	Local                    *bool              `validate:"-" bun:",nullzero,notnull,default:false"`                                                   // is this status from a local account?
	AccountID                string             `validate:"required,ulid" bun:"type:CHAR(26),nullzero,notnull"`                                        // which account posted this status?
	Account                  *Account           `validate:"-" bun:"rel:belongs-to"`                                                                    // account corresponding to accountID
	AccountURI               string             `validate:"required,url" bun:",nullzero,notnull"`                                                      // activitypub uri of the owner of this status
	InReplyToID              string             `validate:"required_with=InReplyToURI InReplyToAccountID,omitempty,ulid" bun:"type:CHAR(26),nullzero"` // id of the status this status replies to
	InReplyToURI             string             `validate:"required_with=InReplyToID InReplyToAccountID,omitempty,url" bun:",nullzero"`                // activitypub uri of the status this status is a reply to
	InReplyToAccountID       string             `validate:"required_with=InReplyToID InReplyToURI,omitempty,ulid" bun:"type:CHAR(26),nullzero"`        // id of the account that this status replies to
	InReplyTo                *Status            `validate:"-" bun:"-"`                                                                                 // status corresponding to inReplyToID
	InReplyToAccount         *Account           `validate:"-" bun:"rel:belongs-to"`                                                                    // account corresponding to inReplyToAccountID
	BoostOfID                string             `validate:"required_with=BoostOfAccountID,omitempty,ulid" bun:"type:CHAR(26),nullzero"`                // id of the status this status is a boost of
	BoostOfAccountID         string             `validate:"required_with=BoostOfID,omitempty,ulid" bun:"type:CHAR(26),nullzero"`                       // id of the account that owns the boosted status
	BoostOf                  *Status            `validate:"-" bun:"-"`                                                                                 // status that corresponds to boostOfID
	BoostOfAccount           *Account           `validate:"-" bun:"rel:belongs-to"`                                                                    // account that corresponds to boostOfAccountID
	ContentWarning           string             `validate:"-" bun:",nullzero"`                                                                         // cw string for this status
	Visibility               Visibility         `validate:"oneof=public unlocked followers_only mutuals_only direct" bun:",nullzero,notnull"`          // visibility entry for this status
	Sensitive                *bool              `validate:"-" bun:",nullzero,notnull,default:false"`                                                   // mark the status as sensitive?
	Language                 string             `validate:"-" bun:",nullzero"`                                                                         // what language is this status written in?
	CreatedWithApplicationID string             `validate:"required_if=Local true,omitempty,ulid" bun:"type:CHAR(26),nullzero"`                        // Which application was used to create this status?
	CreatedWithApplication   *Application       `validate:"-" bun:"rel:belongs-to"`                                                                    // application corresponding to createdWithApplicationID
	ActivityStreamsType      string             `validate:"required" bun:",nullzero,notnull"`                                                          // What is the activitystreams type of this status? See: https://www.w3.org/TR/activitystreams-vocabulary/#object-types. Will probably almost always be Note but who knows!.
	Text                     string             `validate:"-" bun:""`                                                                                  // Original text of the status without formatting
	Federated                *bool              `validate:"-" bun:",notnull"`                                                                          // This status will be federated beyond the local timeline(s)
	Boostable                *bool              `validate:"-" bun:",notnull"`                                                                          // This status can be boosted/reblogged
	Replyable                *bool              `validate:"-" bun:",notnull"`                                                                          // This status can be replied to
	Likeable                 *bool              `validate:"-" bun:",notnull"`                                                                          // This status can be liked/faved
}

// GetID implements timeline.Timelineable{}.
func (s *Status) GetID() string {
	return s.ID
}

// GetAccountID implements timeline.Timelineable{}.
func (s *Status) GetAccountID() string {
	return s.AccountID
}

// GetBoostID implements timeline.Timelineable{}.
func (s *Status) GetBoostOfID() string {
	return s.BoostOfID
}

// GetBoostOfAccountID implements timeline.Timelineable{}.
func (s *Status) GetBoostOfAccountID() string {
	return s.BoostOfAccountID
}

// AttachmentsPopulated returns whether media attachments are populated according to current AttachmentIDs.
func (s *Status) AttachmentsPopulated() bool {
	if len(s.AttachmentIDs) != len(s.Attachments) {
		// this is the quickest indicator.
		return false
	}

	// Attachments must be in same order.
	for i, id := range s.AttachmentIDs {
		if s.Attachments[i] == nil {
			log.Warnf(nil, "nil attachment in slice for status %s", s.URI)
			continue
		}
		if s.Attachments[i].ID != id {
			return false
		}
	}

	return true
}

// TagsPopulated returns whether tags are populated according to current TagIDs.
func (s *Status) TagsPopulated() bool {
	if len(s.TagIDs) != len(s.Tags) {
		// this is the quickest indicator.
		return false
	}

	// Tags must be in same order.
	for i, id := range s.TagIDs {
		if s.Tags[i] == nil {
			log.Warnf(nil, "nil tag in slice for status %s", s.URI)
			continue
		}
		if s.Tags[i].ID != id {
			return false
		}
	}

	return true
}

// MentionsPopulated returns whether mentions are populated according to current MentionIDs.
func (s *Status) MentionsPopulated() bool {
	if len(s.MentionIDs) != len(s.Mentions) {
		// this is the quickest indicator.
		return false
	}

	// Mentions must be in same order.
	for i, id := range s.MentionIDs {
		if s.Mentions[i] == nil {
			log.Warnf(nil, "nil mention in slice for status %s", s.URI)
			continue
		}
		if s.Mentions[i].ID != id {
			return false
		}
	}

	return true
}

// EmojisPopulated returns whether emojis are populated according to current EmojiIDs.
func (s *Status) EmojisPopulated() bool {
	if len(s.EmojiIDs) != len(s.Emojis) {
		// this is the quickest indicator.
		return false
	}

	// Emojis must be in same order.
	for i, id := range s.EmojiIDs {
		if s.Emojis[i] == nil {
			log.Warnf(nil, "nil emoji in slice for status %s", s.URI)
			continue
		}
		if s.Emojis[i].ID != id {
			return false
		}
	}

	return true
}

// MentionsAccount returns whether status mentions the given account ID.
func (s *Status) MentionsAccount(id string) bool {
	for _, mention := range s.Mentions {
		if mention.TargetAccountID == id {
			return true
		}
	}
	return false
}

// StatusToTag is an intermediate struct to facilitate the many2many relationship between a status and one or more tags.
type StatusToTag struct {
	StatusID string  `validate:"ulid,required" bun:"type:CHAR(26),unique:statustag,nullzero,notnull"`
	Status   *Status `validate:"-" bun:"rel:belongs-to"`
	TagID    string  `validate:"ulid,required" bun:"type:CHAR(26),unique:statustag,nullzero,notnull"`
	Tag      *Tag    `validate:"-" bun:"rel:belongs-to"`
}

// StatusToEmoji is an intermediate struct to facilitate the many2many relationship between a status and one or more emojis.
type StatusToEmoji struct {
	StatusID string  `validate:"ulid,required" bun:"type:CHAR(26),unique:statusemoji,nullzero,notnull"`
	Status   *Status `validate:"-" bun:"rel:belongs-to"`
	EmojiID  string  `validate:"ulid,required" bun:"type:CHAR(26),unique:statusemoji,nullzero,notnull"`
	Emoji    *Emoji  `validate:"-" bun:"rel:belongs-to"`
}

// Visibility represents the visibility granularity of a status.
type Visibility string

const (
	// VisibilityPublic means this status will be visible to everyone on all timelines.
	VisibilityPublic Visibility = "public"
	// VisibilityUnlocked means this status will be visible to everyone, but will only show on home timeline to followers, and in lists.
	VisibilityUnlocked Visibility = "unlocked"
	// VisibilityFollowersOnly means this status is viewable to followers only.
	VisibilityFollowersOnly Visibility = "followers_only"
	// VisibilityMutualsOnly means this status is visible to mutual followers only.
	VisibilityMutualsOnly Visibility = "mutuals_only"
	// VisibilityDirect means this status is visible only to mentioned recipients.
	VisibilityDirect Visibility = "direct"
	// VisibilityDefault is used when no other setting can be found.
	VisibilityDefault Visibility = VisibilityUnlocked
)

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

package gtsmodel

import (
	"time"
)

// Status represents a user-created 'post' or 'status' in the database, either remote or local
type Status struct {
	ID                       string             `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                              // id of this item in the database
	CreatedAt                time.Time          `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                       // when was item created
	UpdatedAt                time.Time          `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                       // when was item last updated
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
	Pinned                   *bool              `validate:"-" bun:",nullzero,notnull,default:false"`                                                   // Has this status been pinned by its owner?
	Federated                *bool              `validate:"-" bun:",notnull"`                                                                          // This status will be federated beyond the local timeline(s)
	Boostable                *bool              `validate:"-" bun:",notnull"`                                                                          // This status can be boosted/reblogged
	Replyable                *bool              `validate:"-" bun:",notnull"`                                                                          // This status can be replied to
	Likeable                 *bool              `validate:"-" bun:",notnull"`                                                                          // This status can be liked/faved
}

/*
	The below functions are added onto the gtsmodel status so that it satisfies
	the Timelineable interface in internal/timeline.
*/

func (s *Status) GetID() string {
	return s.ID
}

func (s *Status) GetAccountID() string {
	return s.AccountID
}

func (s *Status) GetBoostOfID() string {
	return s.BoostOfID
}

func (s *Status) GetBoostOfAccountID() string {
	return s.BoostOfAccountID
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

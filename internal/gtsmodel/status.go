/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	// id of the status in the database
	ID string `bun:"type:CHAR(26),pk,notnull"`
	// uri at which this status is reachable
	URI string `bun:",unique,nullzero"`
	// web url for viewing this status
	URL string `bun:",unique,nullzero"`
	// the html-formatted content of this status
	Content string `bun:",nullzero"`
	// Database IDs of any media attachments associated with this status
	AttachmentIDs []string           `bun:"attachments,array"`
	Attachments   []*MediaAttachment `bun:"attached_media,rel:has-many"`
	// Database IDs of any tags used in this status
	TagIDs []string `bun:"tags,array"`
	Tags   []*Tag   `bun:"attached_tags,m2m:status_to_tags"` // https://bun.uptrace.dev/guide/relations.html#many-to-many-relation
	// Database IDs of any mentions in this status
	MentionIDs []string   `bun:"mentions,array"`
	Mentions   []*Mention `bun:"attached_mentions,rel:has-many"`
	// Database IDs of any emojis used in this status
	EmojiIDs []string `bun:"emojis,array"`
	Emojis   []*Emoji `bun:"attached_emojis,m2m:status_to_emojis"` // https://bun.uptrace.dev/guide/relations.html#many-to-many-relation
	// when was this status created?
	CreatedAt time.Time `bun:",notnull,nullzero,default:current_timestamp"`
	// when was this status updated?
	UpdatedAt time.Time `bun:",notnull,nullzero,default:current_timestamp"`
	// is this status from a local account?
	Local bool
	// which account posted this status?
	AccountID string   `bun:"type:CHAR(26),notnull"`
	Account   *Account `bun:"rel:belongs-to"`
	// AP uri of the owner of this status
	AccountURI string `bun:",nullzero"`
	// id of the status this status is a reply to
	InReplyToID string  `bun:"type:CHAR(26),nullzero"`
	InReplyTo   *Status `bun:"-"`
	// AP uri of the status this status is a reply to
	InReplyToURI string `bun:",nullzero"`
	// id of the account that this status replies to
	InReplyToAccountID string   `bun:"type:CHAR(26),nullzero"`
	InReplyToAccount   *Account `bun:"rel:belongs-to"`
	// id of the status this status is a boost of
	BoostOfID string  `bun:"type:CHAR(26),nullzero"`
	BoostOf   *Status `bun:"-"`
	// id of the account that owns the boosted status
	BoostOfAccountID string   `bun:"type:CHAR(26),nullzero"`
	BoostOfAccount   *Account `bun:"rel:belongs-to"`
	// cw string for this status
	ContentWarning string `bun:",nullzero"`
	// visibility entry for this status
	Visibility Visibility `bun:",notnull"`
	// mark the status as sensitive?
	Sensitive bool
	// what language is this status written in?
	Language string `bun:",nullzero"`
	// Which application was used to create this status?
	CreatedWithApplicationID string       `bun:"type:CHAR(26),nullzero"`
	CreatedWithApplication   *Application `bun:"rel:belongs-to"`
	// advanced visibility for this status
	VisibilityAdvanced *VisibilityAdvanced
	// What is the activitystreams type of this status? See: https://www.w3.org/TR/activitystreams-vocabulary/#object-types
	// Will probably almost always be Note but who knows!.
	ActivityStreamsType string `bun:",nullzero"`
	// Original text of the status without formatting
	Text string `bun:",nullzero"`
	// Has this status been pinned by its owner?
	Pinned bool
}

// StatusToTag is an intermediate struct to facilitate the many2many relationship between a status and one or more tags.
type StatusToTag struct {
	StatusID string  `bun:"type:CHAR(26),unique:statustag,nullzero"`
	Status   *Status `bun:"rel:belongs-to"`
	TagID    string  `bun:"type:CHAR(26),unique:statustag,nullzero"`
	Tag      *Tag    `bun:"rel:belongs-to"`
}

// StatusToEmoji is an intermediate struct to facilitate the many2many relationship between a status and one or more emojis.
type StatusToEmoji struct {
	StatusID string  `bun:"type:CHAR(26),unique:statusemoji,nullzero"`
	Status   *Status `bun:"rel:belongs-to"`
	EmojiID  string  `bun:"type:CHAR(26),unique:statusemoji,nullzero"`
	Emoji    *Emoji  `bun:"rel:belongs-to"`
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

// VisibilityAdvanced models flags for fine-tuning visibility and interactivity of a status.
//
// All flags default to true.
//
// If PUBLIC is selected, flags will all be overwritten to TRUE regardless of what is selected.
//
// If UNLOCKED is selected, any flags can be turned on or off in any combination.
//
// If FOLLOWERS-ONLY or MUTUALS-ONLY are selected, boostable will always be FALSE. Other flags can be turned on or off as desired.
//
// If DIRECT is selected, boostable will be FALSE, and all other flags will be TRUE.
type VisibilityAdvanced struct {
	// This status will be federated beyond the local timeline(s)
	Federated bool `bun:"default:true"`
	// This status can be boosted/reblogged
	Boostable bool `bun:"default:true"`
	// This status can be replied to
	Replyable bool `bun:"default:true"`
	// This status can be liked/faved
	Likeable bool `bun:"default:true"`
}

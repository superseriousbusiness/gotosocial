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
	ID string `pg:"type:CHAR(26),pk,notnull"`
	// uri at which this status is reachable
	URI string `pg:",unique"`
	// web url for viewing this status
	URL string `pg:",unique"`
	// the html-formatted content of this status
	Content string
	// Database IDs of any media attachments associated with this status
	AttachmentIDs []string           `pg:"attachments,array"`
	Attachments   []*MediaAttachment `pg:"attached_media,rel:has-many"`
	// Database IDs of any tags used in this status
	TagIDs []string `pg:"tags,array"`
	Tags   []*Tag   `pg:"attached_tags,many2many:status_to_tags"` // https://pg.uptrace.dev/orm/many-to-many-relation/
	// Database IDs of any mentions in this status
	MentionIDs []string   `pg:"mentions,array"`
	Mentions   []*Mention `pg:"attached_mentions,rel:has-many"`
	// Database IDs of any emojis used in this status
	EmojiIDs []string `pg:"emojis,array"`
	Emojis   []*Emoji `pg:"attached_emojis,many2many:status_to_emojis"` // https://pg.uptrace.dev/orm/many-to-many-relation/
	// when was this status created?
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// when was this status updated?
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// is this status from a local account?
	Local bool
	// which account posted this status?
	AccountID string   `pg:"type:CHAR(26),notnull"`
	Account   *Account `pg:"rel:has-one"`
	// AP uri of the owner of this status
	AccountURI string
	// id of the status this status is a reply to
	InReplyToID string  `pg:"type:CHAR(26)"`
	InReplyTo   *Status `pg:"rel:has-one"`
	// AP uri of the status this status is a reply to
	InReplyToURI string
	// id of the account that this status replies to
	InReplyToAccountID string   `pg:"type:CHAR(26)"`
	InReplyToAccount   *Account `pg:"rel:has-one"`
	// id of the status this status is a boost of
	BoostOfID string  `pg:"type:CHAR(26)"`
	BoostOf   *Status `pg:"rel:has-one"`
	// id of the account that owns the boosted status
	BoostOfAccountID string   `pg:"type:CHAR(26)"`
	BoostOfAccount   *Account `pg:"rel:has-one"`
	// cw string for this status
	ContentWarning string
	// visibility entry for this status
	Visibility Visibility `pg:",notnull"`
	// mark the status as sensitive?
	Sensitive bool
	// what language is this status written in?
	Language string
	// Which application was used to create this status?
	CreatedWithApplicationID string       `pg:"type:CHAR(26)"`
	CreatedWithApplication   *Application `pg:"rel:has-one"`
	// advanced visibility for this status
	VisibilityAdvanced *VisibilityAdvanced
	// What is the activitystreams type of this status? See: https://www.w3.org/TR/activitystreams-vocabulary/#object-types
	// Will probably almost always be Note but who knows!.
	ActivityStreamsType string
	// Original text of the status without formatting
	Text string
	// Has this status been pinned by its owner?
	Pinned bool
}

// StatusToTag is an intermediate struct to facilitate the many2many relationship between a status and one or more tags.
type StatusToTag struct {
	StatusID string `pg:"unique:statustag"`
	TagID    string `pg:"unique:statustag"`
}

// StatusToEmoji is an intermediate struct to facilitate the many2many relationship between a status and one or more emojis.
type StatusToEmoji struct {
	StatusID string `pg:"unique:statusemoji"`
	EmojiID  string `pg:"unique:statusemoji"`
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
	Federated bool `pg:"default:true"`
	// This status can be boosted/reblogged
	Boostable bool `pg:"default:true"`
	// This status can be replied to
	Replyable bool `pg:"default:true"`
	// This status can be liked/faved
	Likeable bool `pg:"default:true"`
}

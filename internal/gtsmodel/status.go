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

import "time"

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
	Attachments []string `pg:",array"`
	// Database IDs of any tags used in this status
	Tags []string `pg:",array"`
	// Database IDs of any mentions in this status
	Mentions []string `pg:",array"`
	// Database IDs of any emojis used in this status
	Emojis []string `pg:",array"`
	// when was this status created?
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// when was this status updated?
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// is this status from a local account?
	Local bool
	// which account posted this status?
	AccountID string `pg:"type:CHAR(26),notnull"`
	// AP uri of the owner of this status
	AccountURI string
	// id of the status this status is a reply to
	InReplyToID string `pg:"type:CHAR(26)"`
	// AP uri of the status this status is a reply to
	InReplyToURI string
	// id of the account that this status replies to
	InReplyToAccountID string `pg:"type:CHAR(26)"`
	// id of the status this status is a boost of
	BoostOfID string `pg:"type:CHAR(26)"`
	// id of the account that owns the boosted status
	BoostOfAccountID string `pg:"type:CHAR(26)"`
	// cw string for this status
	ContentWarning string
	// visibility entry for this status
	Visibility Visibility `pg:",notnull"`
	// mark the status as sensitive?
	Sensitive bool
	// what language is this status written in?
	Language string
	// Which application was used to create this status?
	CreatedWithApplicationID string `pg:"type:CHAR(26)"`
	// advanced visibility for this status
	VisibilityAdvanced *VisibilityAdvanced
	// What is the activitystreams type of this status? See: https://www.w3.org/TR/activitystreams-vocabulary/#object-types
	// Will probably almost always be Note but who knows!.
	ActivityStreamsType string
	// Original text of the status without formatting
	Text string
	// Has this status been pinned by its owner?
	Pinned bool

	/*
		INTERNAL MODEL NON-DATABASE FIELDS

		These are for convenience while passing the status around internally,
		but these fields should *never* be put in the db.
	*/

	// Account that created this status
	GTSAuthorAccount *Account `pg:"-"`
	// Mentions created in this status
	GTSMentions []*Mention `pg:"-"`
	// Hashtags used in this status
	GTSTags []*Tag `pg:"-"`
	// Emojis used in this status
	GTSEmojis []*Emoji `pg:"-"`
	// MediaAttachments used in this status
	GTSMediaAttachments []*MediaAttachment `pg:"-"`
	// Status being replied to
	GTSReplyToStatus *Status `pg:"-"`
	// Account being replied to
	GTSReplyToAccount *Account `pg:"-"`
	// Status being boosted
	GTSBoostedStatus *Status `pg:"-"`
	// Account of the boosted status
	GTSBoostedAccount *Account `pg:"-"`
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

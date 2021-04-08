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
	ID string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull"`
	// uri at which this status is reachable
	URI string `pg:",unique"`
	// web url for viewing this status
	URL string `pg:",unique"`
	// the html-formatted content of this status
	Content string
	// when was this status created?
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// when was this status updated?
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// is this status from a local account?
	Local bool
	// which account posted this status?
	AccountID string
	// id of the status this status is a reply to
	InReplyToID string
	// id of the account that this status replies to
	InReplyToAccountID string
	// id of the status this status is a boost of
	BoostOfID string
	// cw string for this status
	ContentWarning string
	// visibility entry for this status
	Visibility Visibility
	// mark the status as sensitive?
	Sensitive bool
	// what language is this status written in?
	Language string
	// advanced visibility for this status
	VisibilityAdvanced *VisibilityAdvanced
	// What is the activitystreams type of this status? See: https://www.w3.org/TR/activitystreams-vocabulary/#object-types
	// Will probably almost always be Note but who knows!.
	ActivityStreamsType ActivityStreamsObject

	/*
		NON-DATABASE FIELDS

		These are for convenience while passing the status around internally,
		but these fields should never be put in the db.
	*/

	// Mentions created in this status
	Mentions []*Mention `pg:"-"`
	// Hashtags used in this status
	Tags []*Tag `pg:"-"`
	// Emojis used in this status
	Emojis []*Emoji `pg:"-"`
	// Attachments used in this status
	Attachments []*MediaAttachment `pg:"-"`
	// Status being replied to
	ReplyToStatus *Status `pg:"-"`
	// Account being replied to
	ReplyToAccount *Account `pg:"-"`
}

// Visibility represents the visibility granularity of a status.
type Visibility string

const (
	// This status will be visible to everyone on all timelines.
	VisibilityPublic Visibility = "public"
	// This status will be visible to everyone, but will only show on home timeline to followers, and in lists.
	VisibilityUnlocked Visibility = "unlocked"
	// This status is viewable to followers only.
	VisibilityFollowersOnly Visibility = "followers_only"
	// This status is visible to mutual followers only.
	VisibilityMutualsOnly Visibility = "mutuals_only"
	// This status is visible only to mentioned recipients
	VisibilityDirect Visibility = "direct"
)

type VisibilityAdvanced struct {
	/*
		ADVANCED SETTINGS -- These should all default to TRUE.

		If PUBLIC is selected, they will all be overwritten to TRUE regardless of what is selected.
		If UNLOCKED is selected, any of them can be turned on or off in any combination.
		If FOLLOWERS-ONLY or MUTUALS-ONLY are selected, boostable will always be FALSE. The others can be turned on or off as desired.
		If DIRECT is selected, boostable will be FALSE, and all other flags will be TRUE.
	*/
	// This status will be federated beyond the local timeline(s)
	Federated bool `pg:"default:true"`
	// This status can be boosted/reblogged
	Boostable bool `pg:"default:true"`
	// This status can be replied to
	Replyable bool `pg:"default:true"`
	// This status can be liked/faved
	Likeable bool `pg:"default:true"`
}

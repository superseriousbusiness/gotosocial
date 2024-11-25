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

// Status represents a user-created 'post' or 'status' in the database, either remote or local
type Status struct {
	ID                       string     `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt                time.Time  `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt                time.Time  `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	FetchedAt                time.Time  `bun:"type:timestamptz,nullzero"`                                   // when was item (remote) last fetched.
	PinnedAt                 time.Time  `bun:"type:timestamptz,nullzero"`                                   // Status was pinned by owning account at this time.
	URI                      string     `bun:",unique,nullzero,notnull"`                                    // activitypub URI of this status
	URL                      string     `bun:",nullzero"`                                                   // web url for viewing this status
	Content                  string     `bun:""`                                                            // content of this status; likely html-formatted but not guaranteed
	AttachmentIDs            []string   `bun:"attachments,array"`                                           // Database IDs of any media attachments associated with this status
	TagIDs                   []string   `bun:"tags,array"`                                                  // Database IDs of any tags used in this status
	MentionIDs               []string   `bun:"mentions,array"`                                              // Database IDs of any mentions in this status
	EmojiIDs                 []string   `bun:"emojis,array"`                                                // Database IDs of any emojis used in this status
	Local                    *bool      `bun:",nullzero,notnull,default:false"`                             // is this status from a local account?
	AccountID                string     `bun:"type:CHAR(26),nullzero,notnull"`                              // which account posted this status?
	AccountURI               string     `bun:",nullzero,notnull"`                                           // activitypub uri of the owner of this status
	InReplyToID              string     `bun:"type:CHAR(26),nullzero"`                                      // id of the status this status replies to
	InReplyToURI             string     `bun:",nullzero"`                                                   // activitypub uri of the status this status is a reply to
	InReplyToAccountID       string     `bun:"type:CHAR(26),nullzero"`                                      // id of the account that this status replies to
	BoostOfID                string     `bun:"type:CHAR(26),nullzero"`                                      // id of the status this status is a boost of
	BoostOfAccountID         string     `bun:"type:CHAR(26),nullzero"`                                      // id of the account that owns the boosted status
	ThreadID                 string     `bun:"type:CHAR(26),nullzero"`                                      // id of the thread to which this status belongs; only set for remote statuses if a local account is involved at some point in the thread, otherwise null
	PollID                   string     `bun:"type:CHAR(26),nullzero"`                                      //
	ContentWarning           string     `bun:",nullzero"`                                                   // cw string for this status
	Visibility               Visibility `bun:",nullzero,notnull"`                                           // visibility entry for this status
	Sensitive                *bool      `bun:",nullzero,notnull,default:false"`                             // mark the status as sensitive?
	Language                 string     `bun:",nullzero"`                                                   // what language is this status written in?
	CreatedWithApplicationID string     `bun:"type:CHAR(26),nullzero"`                                      // Which application was used to create this status?
	ActivityStreamsType      string     `bun:",nullzero,notnull"`                                           // What is the activitystreams type of this status? See: https://www.w3.org/TR/activitystreams-vocabulary/#object-types. Will probably almost always be Note but who knows!.
	Text                     string     `bun:""`                                                            // Original text of the status without formatting
	Federated                *bool      `bun:",notnull"`                                                    // This status will be federated beyond the local timeline(s)
	Boostable                *bool      `bun:",notnull"`                                                    // This status can be boosted/reblogged
	Replyable                *bool      `bun:",notnull"`                                                    // This status can be replied to
	Likeable                 *bool      `bun:",notnull"`                                                    // This status can be liked/faved
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

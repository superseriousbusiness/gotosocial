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

// SinBinStatus represents a status that's been rejected and/or reported + quarantined.
//
// Automatically rejected statuses are not put in the sin bin, only statuses that were
// stored on the instance and which someone (local or remote) has subsequently rejected.
type SinBinStatus struct {
	ID                  string     `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // ID of this item in the database.
	CreatedAt           time.Time  `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // Creation time of this item.
	UpdatedAt           time.Time  `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // Last-updated time of this item.
	URI                 string     `bun:",unique,nullzero,notnull"`                                    // ActivityPub URI/ID of this status.
	URL                 string     `bun:",nullzero"`                                                   // Web url for viewing this status.
	Domain              string     `bun:",nullzero"`                                                   // Domain of the status, will be null if this is a local status, otherwise something like `example.org`.
	AccountURI          string     `bun:",nullzero,notnull"`                                           // ActivityPub uri of the author of this status.
	InReplyToURI        string     `bun:",nullzero"`                                                   // ActivityPub uri of the status this status is a reply to.
	Content             string     `bun:",nullzero"`                                                   // Content of this status.
	AttachmentLinks     []string   `bun:",nullzero,array"`                                             // Links to attachments of this status.
	MentionTargetURIs   []string   `bun:",nullzero,array"`                                             // URIs of mentioned accounts.
	EmojiLinks          []string   `bun:",nullzero,array"`                                             // Links to any emoji images used in this status.
	PollOptions         []string   `bun:",nullzero,array"`                                             // String values of any poll options used in this status.
	ContentWarning      string     `bun:",nullzero"`                                                   // CW / subject string for this status.
	Visibility          Visibility `bun:",nullzero,notnull"`                                           // Visibility level of this status.
	Sensitive           *bool      `bun:",nullzero,notnull,default:false"`                             // Mark the status as sensitive.
	Language            string     `bun:",nullzero"`                                                   // Language code for this status.
	ActivityStreamsType string     `bun:",nullzero,notnull"`                                           // ActivityStreams type of this status.
}

// Visibility represents the visibility granularity of a status.
type Visibility string

const (
	// VisibilityNone means nobody can see this.
	// It's only used for web status visibility.
	VisibilityNone Visibility = "none"
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

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

type AccountSettings struct {
	ID                string     `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt         time.Time  `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created.
	UpdatedAt         time.Time  `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item was last updated.
	Reason            string     `bun:",nullzero"`                                                   // What reason was given for signing up when this account was created?
	Privacy           Visibility `bun:",nullzero"`                                                   // Default post privacy for this account
	Sensitive         *bool      `bun:",nullzero,notnull,default:false"`                             // Set posts from this account to sensitive by default?
	Language          string     `bun:",nullzero,notnull,default:'en'"`                              // What language does this account post in?
	StatusContentType string     `bun:",nullzero"`                                                   // What is the default format for statuses posted by this account (only for local accounts).
	CustomCSS         string     `bun:",nullzero"`                                                   // Custom CSS that should be displayed for this Account's profile and statuses.
	EnableRSS         *bool      `bun:",nullzero,notnull,default:false"`                             // enable RSS feed subscription for this account's public posts at [URL]/feed
	HideCollections   *bool      `bun:",nullzero,notnull,default:false"`                             // Hide this account's followers/following collections.
}

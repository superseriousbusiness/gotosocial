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
	"regexp"
	"time"
)

// Filter stores a filter created by a local account.
type Filter struct {
	ID                   string           `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt            time.Time        `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt            time.Time        `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	ExpiresAt            time.Time        `bun:"type:timestamptz,nullzero"`                                   // Time filter should expire. If null, should not expire.
	AccountID            string           `bun:"type:CHAR(26),notnull,nullzero"`                              // ID of the local account that created the filter.
	Title                string           `bun:",nullzero,notnull,unique"`                                    // The name of the filter.
	Action               string           `bun:",nullzero,notnull"`                                           // The action to take.
	Keywords             []*FilterKeyword `bun:"-"`                                                           // Keywords for this filter.
	Statuses             []*FilterStatus  `bun:"-"`                                                           // Statuses for this filter.
	ContextHome          *bool            `bun:",nullzero,notnull,default:false"`                             // Apply filter to home timeline and lists.
	ContextNotifications *bool            `bun:",nullzero,notnull,default:false"`                             // Apply filter to notifications.
	ContextPublic        *bool            `bun:",nullzero,notnull,default:false"`                             // Apply filter to home timeline and lists.
	ContextThread        *bool            `bun:",nullzero,notnull,default:false"`                             // Apply filter when viewing a status's associated thread.
	ContextAccount       *bool            `bun:",nullzero,notnull,default:false"`                             // Apply filter when viewing an account profile.
}

// FilterKeyword stores a single keyword to filter statuses against.
type FilterKeyword struct {
	ID        string         `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                                     // id of this item in the database
	CreatedAt time.Time      `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                  // when was item created
	UpdatedAt time.Time      `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                  // when was item last updated
	AccountID string         `bun:"type:CHAR(26),notnull,nullzero"`                                               // ID of the local account that created the filter keyword.
	FilterID  string         `bun:"type:CHAR(26),notnull,nullzero,unique:filter_keywords_filter_id_keyword_uniq"` // ID of the filter that this keyword belongs to.
	Filter    *Filter        `bun:"-"`                                                                            // Filter corresponding to FilterID
	Keyword   string         `bun:",nullzero,notnull,unique:filter_keywords_filter_id_keyword_uniq"`              // The keyword or phrase to filter against.
	WholeWord *bool          `bun:",nullzero,notnull,default:false"`                                              // Should the filter consider word boundaries?
	Regexp    *regexp.Regexp `bun:"-"`                                                                            // pre-prepared regular expression
}

// FilterStatus stores a single status to filter.
type FilterStatus struct {
	ID        string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                                       // id of this item in the database
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                    // when was item created
	UpdatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                    // when was item last updated
	AccountID string    `bun:"type:CHAR(26),notnull,nullzero"`                                                 // ID of the local account that created the filter keyword.
	FilterID  string    `bun:"type:CHAR(26),notnull,nullzero,unique:filter_statuses_filter_id_status_id_uniq"` // ID of the filter that this keyword belongs to.
	Filter    *Filter   `bun:"-"`                                                                              // Filter corresponding to FilterID
	StatusID  string    `bun:"type:CHAR(26),notnull,nullzero,unique:filter_statuses_filter_id_status_id_uniq"` // ID of the status to filter.
}

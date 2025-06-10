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

type DomainPermissionSubscription struct {
	// ID of this item in the database.
	ID string `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`

	// Priority of this subscription compared
	// to others of the same permission type.
	// 0-255 (higher = higher priority).
	Priority uint8 `bun:""`

	// Moderator-set title for this list.
	Title string `bun:",nullzero,unique"`

	// Permission type of the subscription.
	PermissionType DomainPermissionType `bun:",nullzero,notnull"`

	// Create domain permission entries
	// resulting from this subscription as drafts.
	AsDraft *bool `bun:",nullzero,notnull,default:true"`

	// Adopt orphaned domain permissions
	// present in this subscription's entries.
	AdoptOrphans *bool `bun:",nullzero,notnull,default:false"`

	// Account ID of the creator of this subscription.
	CreatedByAccountID string `bun:"type:CHAR(26),nullzero,notnull"`

	// Account corresponding to createdByAccountID.
	CreatedByAccount *Account `bun:"-"`

	// URI of the domain permission list.
	URI string `bun:",nullzero,notnull,unique"`

	// Content type to expect from the URI.
	ContentType DomainPermSubContentType `bun:",nullzero,notnull"`

	// Username to send when doing
	// a GET of URI using basic auth.
	FetchUsername string `bun:",nullzero"`

	// Password to send when doing
	// a GET of URI using basic auth.
	FetchPassword string `bun:",nullzero"`

	// Time when fetch of URI was last attempted.
	FetchedAt time.Time `bun:"type:timestamptz,nullzero"`

	// Time when the domain permission list
	// was last *successfuly* fetched, to be
	// transmitted as If-Modified-Since header.
	SuccessfullyFetchedAt time.Time `bun:"type:timestamptz,nullzero"`

	// "Last-Modified" time received from the
	// server (if any) on last successful fetch.
	// Used for HTTP request caching.
	LastModified time.Time `bun:"type:timestamptz,nullzero"`

	// "ETag" header last received from the
	// server (if any) on last successful fetch.
	// Used for HTTP request caching.
	ETag string `bun:"etag,nullzero"`

	// If latest fetch attempt errored,
	// this field stores the error message.
	// Cleared on latest successful fetch.
	Error string `bun:",nullzero"`

	// If true, then when a list is processed, if the
	// list does *not* contain entries that it *did*
	// contain previously, ie., retracted entries,
	// then domain permissions corresponding to those
	// entries will be removed.
	//
	// If false, they will just be orphaned instead.
	RemoveRetracted *bool `bun:",nullzero,notnull,default:true"`
}

type DomainPermSubContentType enumType

const (
	DomainPermSubContentTypeUnknown DomainPermSubContentType = 0 // ???
	DomainPermSubContentTypeCSV     DomainPermSubContentType = 1 // text/csv
	DomainPermSubContentTypeJSON    DomainPermSubContentType = 2 // application/json
	DomainPermSubContentTypePlain   DomainPermSubContentType = 3 // text/plain
)

func (p DomainPermSubContentType) String() string {
	switch p {
	case DomainPermSubContentTypeCSV:
		return "text/csv"
	case DomainPermSubContentTypeJSON:
		return "application/json"
	case DomainPermSubContentTypePlain:
		return "text/plain"
	default:
		panic("unknown content type")
	}
}

func NewDomainPermSubContentType(in string) DomainPermSubContentType {
	switch in {
	case "text/csv":
		return DomainPermSubContentTypeCSV
	case "application/json":
		return DomainPermSubContentTypeJSON
	case "text/plain":
		return DomainPermSubContentTypePlain
	default:
		return DomainPermSubContentTypeUnknown
	}
}

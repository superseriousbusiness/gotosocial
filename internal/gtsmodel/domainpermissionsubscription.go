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
	ID                    string                   `bun:"type:CHAR(26),pk,nullzero,notnull,unique"` // ID of this item in the database.
	Priority              uint8                    `bun:""`                                         // Priority of this subscription compared to others of the same permission type. 0-255 (higher = higher priority).
	Title                 string                   `bun:",nullzero,unique"`                         // Moderator-set title for this list.
	PermissionType        DomainPermissionType     `bun:",nullzero,notnull"`                        // Permission type of the subscription.
	AsDraft               *bool                    `bun:",nullzero,notnull,default:true"`           // Create domain permission entries resulting from this subscription as drafts.
	AdoptOrphans          *bool                    `bun:",nullzero,notnull,default:false"`          // Adopt orphaned domain permissions present in this subscription's entries.
	CreatedByAccountID    string                   `bun:"type:CHAR(26),nullzero,notnull"`           // Account ID of the creator of this subscription.
	CreatedByAccount      *Account                 `bun:"-"`                                        // Account corresponding to createdByAccountID.
	URI                   string                   `bun:",nullzero,notnull,unique"`                 // URI of the domain permission list.
	ContentType           DomainPermSubContentType `bun:",nullzero,notnull"`                        // Content type to expect from the URI.
	FetchUsername         string                   `bun:",nullzero"`                                // Username to send when doing a GET of URI using basic auth.
	FetchPassword         string                   `bun:",nullzero"`                                // Password to send when doing a GET of URI using basic auth.
	FetchedAt             time.Time                `bun:"type:timestamptz,nullzero"`                // Time when fetch of URI was last attempted.
	SuccessfullyFetchedAt time.Time                `bun:"type:timestamptz,nullzero"`                // Time when the domain permission list was last *successfuly* fetched, to be transmitted as If-Modified-Since header.
	LastModified          time.Time                `bun:"type:timestamptz,nullzero"`                // "Last-Modified" time received from the server (if any) on last successful fetch. Used for HTTP request caching.
	ETag                  string                   `bun:"etag,nullzero"`                            // "ETag" header last received from the server (if any) on last successful fetch. Used for HTTP request caching.
	Error                 string                   `bun:",nullzero"`                                // If latest fetch attempt errored, this field stores the error message. Cleared on latest successful fetch.
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

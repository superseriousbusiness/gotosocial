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
	ID                    string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`
	Priority              uint8     `bun:""`
	Title                 string    `bun:",nullzero,unique"`
	PermissionType        uint8     `bun:",nullzero,notnull"`
	AsDraft               *bool     `bun:",nullzero,notnull,default:true"`
	AdoptOrphans          *bool     `bun:",nullzero,notnull,default:false"`
	CreatedByAccountID    string    `bun:"type:CHAR(26),nullzero,notnull"`
	URI                   string    `bun:",nullzero,notnull,unique"`
	ContentType           int16     `bun:",nullzero,notnull"`
	FetchUsername         string    `bun:",nullzero"`
	FetchPassword         string    `bun:",nullzero"`
	FetchedAt             time.Time `bun:"type:timestamptz,nullzero"`
	SuccessfullyFetchedAt time.Time `bun:"type:timestamptz,nullzero"`
	LastModified          time.Time `bun:"type:timestamptz,nullzero"`
	ETag                  string    `bun:"etag,nullzero"`
	Error                 string    `bun:",nullzero"`

	// This is the field added by this migration.
	RemoveRetracted *bool `bun:",nullzero,notnull,default:true"`
}

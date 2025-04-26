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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// DomainPermissionExclude represents one domain that should be excluded
// when domain permission (excludes) are created from subscriptions.
type DomainPermissionExclude struct {
	ID                 string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // ID of this item in the database.
	CreatedAt          time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // Time when this item was created.
	UpdatedAt          time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // Time when this item was last updated.
	Domain             string    `bun:",nullzero,notnull,unique"`                                    // Domain to exclude. Eg. 'whatever.com'.
	CreatedByAccountID string    `bun:"type:CHAR(26),nullzero,notnull"`                              // Account ID of the creator of this exclude.
	CreatedByAccount   *Account  `bun:"-"`                                                           // Account corresponding to createdByAccountID.
	PrivateComment     string    `bun:",nullzero"`                                                   // Private comment on this exclude, viewable to admins.
}

func (d *DomainPermissionExclude) GetID() string {
	return d.ID
}

func (d *DomainPermissionExclude) GetCreatedAt() time.Time {
	return d.CreatedAt
}

func (d *DomainPermissionExclude) GetUpdatedAt() time.Time {
	return d.UpdatedAt
}

func (d *DomainPermissionExclude) SetUpdatedAt(i time.Time) {
	d.UpdatedAt = i
}

func (d *DomainPermissionExclude) GetDomain() string {
	return d.Domain
}

func (d *DomainPermissionExclude) GetCreatedByAccountID() string {
	return d.CreatedByAccountID
}

func (d *DomainPermissionExclude) SetCreatedByAccountID(i string) {
	d.CreatedByAccountID = i
}

func (d *DomainPermissionExclude) GetCreatedByAccount() *Account {
	return d.CreatedByAccount
}

func (d *DomainPermissionExclude) SetCreatedByAccount(i *Account) {
	d.CreatedByAccount = i
}

func (d *DomainPermissionExclude) GetPrivateComment() string {
	return d.PrivateComment
}

func (d *DomainPermissionExclude) SetPrivateComment(i string) {
	d.PrivateComment = i
}

/*
	Stubbed functions for interface purposes.
*/

func (d *DomainPermissionExclude) GetPublicComment() string      { return "" }
func (d *DomainPermissionExclude) SetPublicComment(_ string)     {}
func (d *DomainPermissionExclude) GetObfuscate() *bool           { return util.Ptr(false) }
func (d *DomainPermissionExclude) SetObfuscate(_ *bool)          {}
func (d *DomainPermissionExclude) GetSubscriptionID() string     { return "" }
func (d *DomainPermissionExclude) SetSubscriptionID(_ string)    {}
func (d *DomainPermissionExclude) GetType() DomainPermissionType { return DomainPermissionUnknown }
func (d *DomainPermissionExclude) IsOrphan() bool                { return true }

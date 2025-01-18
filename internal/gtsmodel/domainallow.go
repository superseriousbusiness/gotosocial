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

// DomainAllow represents a federation allow towards a particular domain.
type DomainAllow struct {
	ID                 string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt          time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt          time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	Domain             string    `bun:",nullzero,notnull"`                                           // domain to allow. Eg. 'whatever.com'
	CreatedByAccountID string    `bun:"type:CHAR(26),nullzero,notnull"`                              // Account ID of the creator of this allow
	CreatedByAccount   *Account  `bun:"rel:belongs-to"`                                              // Account corresponding to createdByAccountID
	PrivateComment     string    `bun:""`                                                            // Private comment on this allow, viewable to admins
	PublicComment      string    `bun:""`                                                            // Public comment on this allow, viewable (optionally) by everyone
	Obfuscate          *bool     `bun:",nullzero,notnull,default:false"`                             // whether the domain name should appear obfuscated when displaying it publicly
	SubscriptionID     string    `bun:"type:CHAR(26),nullzero"`                                      // if this allow was created through a subscription, what's the subscription ID?
}

func (d *DomainAllow) GetID() string {
	return d.ID
}

func (d *DomainAllow) GetCreatedAt() time.Time {
	return d.CreatedAt
}

func (d *DomainAllow) GetUpdatedAt() time.Time {
	return d.UpdatedAt
}

func (d *DomainAllow) SetUpdatedAt(i time.Time) {
	d.UpdatedAt = i
}

func (d *DomainAllow) GetDomain() string {
	return d.Domain
}

func (d *DomainAllow) GetCreatedByAccountID() string {
	return d.CreatedByAccountID
}

func (d *DomainAllow) SetCreatedByAccountID(i string) {
	d.CreatedByAccountID = i
}

func (d *DomainAllow) GetCreatedByAccount() *Account {
	return d.CreatedByAccount
}

func (d *DomainAllow) SetCreatedByAccount(i *Account) {
	d.CreatedByAccount = i
}

func (d *DomainAllow) GetPrivateComment() string {
	return d.PrivateComment
}

func (d *DomainAllow) SetPrivateComment(i string) {
	d.PrivateComment = i
}

func (d *DomainAllow) GetPublicComment() string {
	return d.PublicComment
}

func (d *DomainAllow) SetPublicComment(i string) {
	d.PublicComment = i
}

func (d *DomainAllow) GetObfuscate() *bool {
	return d.Obfuscate
}

func (d *DomainAllow) SetObfuscate(i *bool) {
	d.Obfuscate = i
}

func (d *DomainAllow) GetSubscriptionID() string {
	return d.SubscriptionID
}

func (d *DomainAllow) SetSubscriptionID(i string) {
	d.SubscriptionID = i
}

func (d *DomainAllow) GetType() DomainPermissionType {
	return DomainPermissionAllow
}

func (d *DomainAllow) IsOrphan() bool {
	return d.SubscriptionID == ""
}

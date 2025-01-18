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

type DomainPermissionDraft struct {
	ID                 string               `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                                                      // ID of this item in the database.
	CreatedAt          time.Time            `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                                   // Time when this item was created.
	UpdatedAt          time.Time            `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                                   // Time when this item was last updated.
	PermissionType     DomainPermissionType `bun:",notnull,unique:domain_permission_drafts_permission_type_domain_subscription_id_uniq"`          // Permission type of the draft.
	Domain             string               `bun:",nullzero,notnull,unique:domain_permission_drafts_permission_type_domain_subscription_id_uniq"` // Domain to block or allow. Eg. 'whatever.com'.
	CreatedByAccountID string               `bun:"type:CHAR(26),nullzero,notnull"`                                                                // Account ID of the creator of this subscription.
	CreatedByAccount   *Account             `bun:"-"`                                                                                             // Account corresponding to createdByAccountID.
	PrivateComment     string               `bun:",nullzero"`                                                                                     // Private comment on this perm, viewable to admins.
	PublicComment      string               `bun:",nullzero"`                                                                                     // Public comment on this perm, viewable (optionally) by everyone.
	Obfuscate          *bool                `bun:",nullzero,notnull,default:false"`                                                               // Obfuscate domain name when displaying it publicly.
	SubscriptionID     string               `bun:"type:CHAR(26),unique:domain_permission_drafts_permission_type_domain_subscription_id_uniq"`     // ID of the subscription that created this draft, if any.
}

func (d *DomainPermissionDraft) GetID() string {
	return d.ID
}

func (d *DomainPermissionDraft) GetCreatedAt() time.Time {
	return d.CreatedAt
}

func (d *DomainPermissionDraft) GetUpdatedAt() time.Time {
	return d.UpdatedAt
}

func (d *DomainPermissionDraft) SetUpdatedAt(i time.Time) {
	d.UpdatedAt = i
}

func (d *DomainPermissionDraft) GetDomain() string {
	return d.Domain
}

func (d *DomainPermissionDraft) GetCreatedByAccountID() string {
	return d.CreatedByAccountID
}

func (d *DomainPermissionDraft) SetCreatedByAccountID(i string) {
	d.CreatedByAccountID = i
}

func (d *DomainPermissionDraft) GetCreatedByAccount() *Account {
	return d.CreatedByAccount
}

func (d *DomainPermissionDraft) SetCreatedByAccount(i *Account) {
	d.CreatedByAccount = i
}

func (d *DomainPermissionDraft) GetPrivateComment() string {
	return d.PrivateComment
}

func (d *DomainPermissionDraft) SetPrivateComment(i string) {
	d.PrivateComment = i
}

func (d *DomainPermissionDraft) GetPublicComment() string {
	return d.PublicComment
}

func (d *DomainPermissionDraft) SetPublicComment(i string) {
	d.PublicComment = i
}

func (d *DomainPermissionDraft) GetObfuscate() *bool {
	return d.Obfuscate
}

func (d *DomainPermissionDraft) SetObfuscate(i *bool) {
	d.Obfuscate = i
}

func (d *DomainPermissionDraft) GetSubscriptionID() string {
	return d.SubscriptionID
}

func (d *DomainPermissionDraft) SetSubscriptionID(i string) {
	d.SubscriptionID = i
}

func (d *DomainPermissionDraft) GetType() DomainPermissionType {
	return d.PermissionType
}

func (d *DomainPermissionDraft) IsOrphan() bool {
	return d.SubscriptionID == ""
}

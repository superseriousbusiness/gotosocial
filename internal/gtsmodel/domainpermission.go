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

// DomainPermission models a domain permission
// entry -- block / allow / draft / exclude.
type DomainPermission interface {
	GetID() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetUpdatedAt(i time.Time)
	GetDomain() string
	GetCreatedByAccountID() string
	SetCreatedByAccountID(i string)
	GetCreatedByAccount() *Account
	SetCreatedByAccount(i *Account)
	GetPrivateComment() string
	SetPrivateComment(i string)
	GetPublicComment() string
	SetPublicComment(i string)
	GetObfuscate() *bool
	SetObfuscate(i *bool)
	GetSubscriptionID() string
	SetSubscriptionID(i string)
	GetType() DomainPermissionType
}

// Domain permission type.
type DomainPermissionType uint8

const (
	DomainPermissionUnknown DomainPermissionType = iota
	DomainPermissionBlock                        // Explicitly block a domain.
	DomainPermissionAllow                        // Explicitly allow a domain.
)

func (p DomainPermissionType) String() string {
	switch p {
	case DomainPermissionBlock:
		return "block"
	case DomainPermissionAllow:
		return "allow"
	default:
		return "unknown"
	}
}

func NewDomainPermissionType(in string) DomainPermissionType {
	switch in {
	case "block":
		return DomainPermissionBlock
	case "allow":
		return DomainPermissionAllow
	default:
		return DomainPermissionUnknown
	}
}

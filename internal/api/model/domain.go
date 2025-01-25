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

package model

import "mime/multipart"

// Domain represents a remote domain
//
// swagger:model domain
type Domain struct {
	// The hostname of the domain.
	// example: example.org
	Domain string `form:"domain" json:"domain" validate:"required"`
	// Time at which this domain was suspended. Key will not be present on open domains.
	// example: 2021-07-30T09:20:25+00:00
	SuspendedAt string `json:"suspended_at,omitempty"`
	// Time at which this domain was silenced. Key will not be present on open domains.
	// example: 2021-07-30T09:20:25+00:00
	SilencedAt string `json:"silenced_at,omitempty"`
	// If the domain is blocked, what's the publicly-stated reason for the block.
	// example: they smell
	PublicComment string `form:"public_comment" json:"public_comment,omitempty"`
}

// DomainPermission represents a permission applied to one domain (explicit block/allow).
//
// swagger:model domainPermission
type DomainPermission struct {
	Domain
	// The ID of the domain permission entry.
	// example: 01FBW21XJA09XYX51KV5JVBW0F
	// readonly: true
	ID string `json:"id,omitempty"`
	// Obfuscate the domain name when serving this domain permission entry publicly.
	// example: false
	Obfuscate bool `json:"obfuscate,omitempty"`
	// Private comment for this permission entry, visible to this instance's admins only.
	// example: they are poopoo
	PrivateComment string `json:"private_comment,omitempty"`
	// If applicable, the ID of the subscription that caused this domain permission entry to be created.
	// example: 01FBW25TF5J67JW3HFHZCSD23K
	SubscriptionID string `json:"subscription_id,omitempty"`
	// ID of the account that created this domain permission entry.
	// example: 01FBW2758ZB6PBR200YPDDJK4C
	CreatedBy string `json:"created_by,omitempty"`
	// Time at which the permission entry was created (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	CreatedAt string `json:"created_at,omitempty"`
	// Permission type of this entry (block, allow).
	// Only set for domain permission drafts.
	PermissionType string `json:"permission_type,omitempty"`
}

// DomainPermissionRequest is the form submitted as a POST to create a new domain permission entry (allow/block).
//
// swagger:ignore
type DomainPermissionRequest struct {
	// A list of domains for which this permission request should apply.
	// Only used if import=true is specified.
	Domains *multipart.FileHeader `form:"domains" json:"domains"`
	// A single domain for which this permission request should apply.
	// Only used if import=true is NOT specified or if import=false.
	// example: example.org
	Domain string `form:"domain" json:"domain"`
	// Obfuscate the domain name when displaying this permission entry publicly.
	// Ie., instead of 'example.org' show something like 'e**mpl*.or*'.
	// example: false
	Obfuscate bool `form:"obfuscate" json:"obfuscate"`
	// Private comment for other admins on why this permission entry was created.
	// example: don't like 'em!!!!
	PrivateComment string `form:"private_comment" json:"private_comment"`
	// Public comment on why this permission entry was created.
	// Will be visible to requesters at /api/v1/instance/peers if this endpoint is exposed.
	// example: foss dorks 😫
	PublicComment string `form:"public_comment" json:"public_comment"`
	// Permission type to create (only applies to domain permission drafts, not explicit blocks and allows).
	PermissionType string `form:"permission_type" json:"permission_type"`
}

// DomainKeysExpireRequest is the form submitted as a POST to /api/v1/admin/domain_keys_expire to expire a domain's public keys.
//
// swagger:parameters domainKeysExpire
type DomainKeysExpireRequest struct {
	// hostname/domain to expire keys for.
	Domain string `form:"domain" json:"domain"`
}

// DomainPermissionSubscription represents an auto-refreshing subscription to a list of domain permissions (allows, blocks).
//
// swagger:model domainPermissionSubscription
type DomainPermissionSubscription struct {
	// The ID of the domain permission subscription.
	// example: 01FBW21XJA09XYX51KV5JVBW0F
	// readonly: true
	ID string `json:"id"`
	// Priority of this subscription compared to others of the same permission type. 0-255 (higher = higher priority).
	// example: 100
	Priority uint8 `json:"priority"`
	// Title of this subscription, as set by admin who created or updated it.
	// example: really cool list of neato pals
	Title string `json:"title"`
	// The type of domain permission subscription (allow, block).
	// example: block
	PermissionType string `json:"permission_type"`
	// If true, domain permissions arising from this subscription will be created as drafts that must be approved by a moderator to take effect. If false, domain permissions from this subscription will come into force immediately.
	// example: true
	AsDraft bool `json:"as_draft"`
	// If true, this domain permission subscription will "adopt" domain permissions which already exist on the instance, and which meet the following conditions: 1) they have no subscription ID (ie., they're "orphaned") and 2) they are present in the subscribed list. Such orphaned domain permissions will be given this subscription's subscription ID value.
	// example: false
	AdoptOrphans bool `json:"adopt_orphans"`
	// Time at which the subscription was created (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	CreatedAt string `json:"created_at"`
	// ID of the account that created this subscription.
	// example: 01FBW21XJA09XYX51KV5JVBW0F
	// readonly: true
	CreatedBy string `json:"created_by"`
	// URI to call in order to fetch the permissions list.
	// example: https://www.example.org/blocklists/list1.csv
	URI string `json:"uri"`
	// MIME content type to use when parsing the permissions list.
	// example: text/csv
	ContentType string `json:"content_type"`
	// (Optional) username to set for basic auth when doing a fetch of URI.
	// example: admin123
	FetchUsername string `json:"fetch_username,omitempty"`
	// (Optional) password to set for basic auth when doing a fetch of URI.
	// example: admin123
	FetchPassword string `json:"fetch_password,omitempty"`
	// Time of the most recent fetch attempt (successful or otherwise) (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	// readonly: true
	FetchedAt string `json:"fetched_at,omitempty"`
	// Time of the most recent successful fetch (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	// readonly: true
	SuccessfullyFetchedAt string `json:"successfully_fetched_at,omitempty"`
	// If most recent fetch attempt failed, this field will contain an error message related to the fetch attempt.
	// example: Oopsie doopsie, we made a fucky wucky.
	// readonly: true
	Error string `json:"error,omitempty"`
	// Count of domain permission entries discovered at URI on last (successful) fetch.
	// example: 53
	// readonly: true
	Count uint64 `json:"count"`
}

// DomainPermissionSubscriptionRequest represents a request to create or update a domain permission subscription..
//
// swagger:ignore
type DomainPermissionSubscriptionRequest struct {
	// Priority of this subscription compared to others of the same permission type. 0-255 (higher = higher priority).
	// example: 100
	Priority *int `form:"priority" json:"priority"`
	// Title of this subscription, as set by admin who created or updated it.
	// example: really cool list of neato pals
	Title *string `form:"title" json:"title"`
	// The type of domain permission subscription (allow, block).
	// example: block
	PermissionType *string `form:"permission_type" json:"permission_type"`
	// URI to call in order to fetch the permissions list.
	// example: https://www.example.org/blocklists/list1.csv
	URI *string `form:"uri" json:"uri"`
	// MIME content type to use when parsing the permissions list.
	// example: text/csv
	ContentType *string `form:"content_type" json:"content_type"`
	// If true, domain permissions arising from this subscription will be
	// created as drafts that must be approved by a moderator to take effect.
	// If false, domain permissions from this subscription will come into force immediately.
	// example: true
	AsDraft *bool `form:"as_draft" json:"as_draft"`
	//	If true, this domain permission subscription will "adopt" domain permissions
	//	which already exist on the instance, and which meet the following conditions:
	//	1) they have no subscription ID (ie., they're "orphaned") and 2) they are present
	//	in the subscribed list. Such orphaned domain permissions will be given this
	//	subscription's subscription ID value and be managed by this subscription.
	AdoptOrphans *bool `form:"adopt_orphans" json:"adopt_orphans"`
	// (Optional) username to set for basic auth when doing a fetch of URI.
	// example: admin123
	FetchUsername *string `form:"fetch_username" json:"fetch_username"`
	// (Optional) password to set for basic auth when doing a fetch of URI.
	// example: admin123
	FetchPassword *string `form:"fetch_password" json:"fetch_password"`
}

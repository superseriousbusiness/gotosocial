/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

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

// DomainBlock represents a block on one domain
//
// swagger:model domainBlock
type DomainBlock struct {
	Domain
	// The ID of the domain block.
	// example: 01FBW21XJA09XYX51KV5JVBW0F
	// readonly: true
	ID string `json:"id,omitempty"`
	// Obfuscate the domain name when serving this domain block publicly.
	// A useful anti-harassment tool.
	// example: false
	Obfuscate bool `json:"obfuscate,omitempty"`
	// Private comment for this block, visible to our instance admins only.
	// example: they are poopoo
	PrivateComment string `json:"private_comment,omitempty"`
	// The ID of the subscription that created/caused this domain block.
	// example: 01FBW25TF5J67JW3HFHZCSD23K
	SubscriptionID string `json:"subscription_id,omitempty"`
	// ID of the account that created this domain block.
	// example: 01FBW2758ZB6PBR200YPDDJK4C
	CreatedBy string `json:"created_by,omitempty"`
	// Time at which this block was created (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	CreatedAt string `json:"created_at,omitempty"`
}

// DomainBlockCreateRequest is the form submitted as a POST to /api/v1/admin/domain_blocks to create a new block.
//
// swagger:model domainBlockCreateRequest
type DomainBlockCreateRequest struct {
	// A list of domains to block. Only used if import=true is specified.
	Domains *multipart.FileHeader `form:"domains" json:"domains" xml:"domains"`
	// hostname/domain to block
	Domain string `form:"domain" json:"domain" xml:"domain"`
	// whether the domain should be obfuscated when being displayed publicly
	Obfuscate bool `form:"obfuscate" json:"obfuscate" xml:"obfuscate"`
	// private comment for other admins on why the domain was blocked
	PrivateComment string `form:"private_comment" json:"private_comment" xml:"private_comment"`
	// public comment on the reason for the domain block
	PublicComment string `form:"public_comment" json:"public_comment" xml:"public_comment"`
}

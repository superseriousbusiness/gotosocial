/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

// DomainBlock represents a block on one domain
type DomainBlock struct {
	ID             string `json:"id,omitempty"`
	Domain         string `json:"domain"`
	Obfuscate      bool   `json:"obfuscate,omitempty"`
	PrivateComment string `json:"private_comment,omitempty"`
	PublicComment  string `json:"public_comment,omitempty"`
	SubscriptionID string `json:"subscription_id,omitempty"`
	CreatedBy      string `json:"created_by,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
}

// DomainBlockCreateRequest is the form submitted as a POST to /api/v1/admin/domain_blocks to create a new block.
type DomainBlockCreateRequest struct {
	// hostname/domain to block
	Domain string `form:"domain" json:"domain" xml:"domain" validation:"required"`
	// whether the domain should be obfuscated when being displayed publicly
	Obfuscate bool `form:"obfuscate" json:"obfuscate" xml:"obfuscate"`
	// private comment for other admins on why the domain was blocked
	PrivateComment string `form:"private_comment" json:"private_comment" xml:"private_comment"`
	// public comment on the reason for the domain block
	PublicComment string `form:"public_comment" json:"public_comment" xml:"public_comment"`
}

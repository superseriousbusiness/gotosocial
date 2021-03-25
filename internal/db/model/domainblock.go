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

import "time"

// DomainBlock represents a federation block against a particular domain, of varying severity.
type DomainBlock struct {
	// ID of this block in the database
	ID string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull,unique"`
	// Domain to block. If ANY PART of the candidate domain contains this string, it will be blocked.
	// For example: 'example.org' also blocks 'gts.example.org'. '.com' blocks *any* '.com' domains.
	// TODO: implement wildcards here
	Domain string `pg:",notnull"`
	// When was this block created
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// When was this block updated
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// Account ID of the creator of this block
	CreatedByAccountID string `pg:",notnull"`
	// TODO: define this
	Severity           int
	// Reject media from this domain?
	RejectMedia        bool
	// Reject reports from this domain?
	RejectReports      bool
	// Private comment on this block, viewable to admins
	PrivateComment     string
	// Public comment on this block, viewable (optionally) by everyone
	PublicComment      string
}

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

// HeaderFilter represents a regex value filter applied to one particular HTTP header (allow / block).
//
// swagger:model headerFilter
type HeaderFilter struct {
	// The ID of the header filter.
	// example: 01FBW21XJA09XYX51KV5JVBW0F
	// readonly: true
	ID string `json:"id"`

	// The HTTP header to match against.
	// example: User-Agent
	Header string `json:"header"`

	// The header value matching regular expression.
	// example: .*Firefox.*
	Regex string `json:"regex"`

	// The ID of the admin account that created this header filter.
	// example: 01FBW2758ZB6PBR200YPDDJK4C
	// readonly: true
	CreatedBy string `json:"created_by"`

	// Time at which the header filter was created (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	// readonly: true
	CreatedAt string `json:"created_at"`
}

// HeaderFilterRequest is the form submitted as a POST to create a new header filter entry (allow / block).
//
// swagger:parameters headerFilterAllowCreate headerFilterBlockCreate
type HeaderFilterRequest struct {
	// The HTTP header to match against (e.g. User-Agent).
	// required: true
	// in: formData
	Header string `form:"header" json:"header" xml:"header"`

	// The header value matching regular expression.
	// required: true
	// in: formData
	Regex string `form:"regex" json:"regex" xml:"regex"`
}

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

type HeaderFilter struct {
	ID string `json:"id"`

	Header string `json:"header"`

	Regex string `json:"regex"`

	CreatedBy string `json:"created_by,omitempty"`

	CreatedAt string `json:"created_at,omitempty"`
}

type HeaderFilterRequest struct {
	Header string `form:"header" json:"header" xml:"header"`

	Regex string `form:"regex" json:"regex" xml:"regex"`
}

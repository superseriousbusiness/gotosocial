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

// Tag represents a hashtag used within the content of a status.
//
// swagger:model tag
type Tag struct {
	// The value of the hashtag after the # sign.
	// example: helloworld
	Name string `json:"name"`
	// Web link to the hashtag.
	// example: https://example.org/tags/helloworld
	URL string `json:"url"`
	// History of this hashtag's usage.
	// Currently just a stub, if provided will always be an empty array.
	// example: []
	History *[]any `json:"history,omitempty"`
	// Following is true if the user is following this tag, false if they're not,
	// and not present if there is no currently authenticated user.
	Following *bool `json:"following,omitempty"`
}

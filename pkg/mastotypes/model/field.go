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

package mastotypes

// Field represents a profile field as a name-value pair with optional verification. See https://docs.joinmastodon.org/entities/field/
type Field struct {
	// REQUIRED

	// The key of a given field's key-value pair.
	Name string `json:"name"`
	// The value associated with the name key.
	Value string `json:"value"`

	// OPTIONAL

	// Timestamp of when the server verified a URL value for a rel="me” link. String (ISO 8601 Datetime) if value is a verified URL
	VerifiedAt string `json:"verified_at,omitempty"`
}
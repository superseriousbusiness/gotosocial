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

package trans

// TypeKey should be set on a TransEntry to indicate the type of entry it is.
const TypeKey = "type"

// Type describes the type of a trans entry, and how it should be read/serialized.
type Type string

// Type of the trans entry. Describes how it should be read from file.
const (
	TransAccount          Type = "account"
	TransBlock            Type = "block"
	TransDomainBlock      Type = "domainBlock"
	TransEmailDomainBlock Type = "emailDomainBlock"
	TransFollow           Type = "follow"
	TransFollowRequest    Type = "followRequest"
	TransInstance         Type = "instance"
	TransUser             Type = "user"
)

// Entry is used for deserializing trans entries into a rough interface so that
// the TypeKey can be fetched, before continuing with full parsing.
type Entry map[string]interface{}

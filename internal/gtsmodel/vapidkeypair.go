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

// VAPIDKeyPair represents the instance's VAPID keys (stored as Base64 strings).
// This table should only ever have one entry, with a known ID of 0.
//
// See: https://datatracker.ietf.org/doc/html/rfc8292
type VAPIDKeyPair struct {
	ID      int    `bun:",pk,notnull"`
	Public  string `bun:",notnull,nullzero"`
	Private string `bun:",notnull,nullzero"`
}

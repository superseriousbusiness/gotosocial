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

// Thread represents one thread of statuses.
// TODO: add more fields here if necessary.
type Thread struct {
	ID        string   `bun:"type:CHAR(26),pk,nullzero,notnull,unique"` // id of this item in the database
	StatusIDs []string `bun:"-"`                                        // ids of statuses belonging to this thread (order not guaranteed)
}

// ThreadToStatus is an intermediate struct to facilitate the
// many2many relationship between a thread and one or more statuses.
type ThreadToStatus struct {
	ThreadID string `bun:"type:CHAR(26),unique:statusthread,nullzero,notnull"`
	StatusID string `bun:"type:CHAR(26),unique:statusthread,nullzero,notnull"`
}

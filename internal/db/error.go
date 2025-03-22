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

package db

import (
	"database/sql"
	"errors"
)

var (
	// ErrNoEntries is a direct ptr to sql.ErrNoRows since that is returned regardless
	// of DB dialect. It is returned when no rows (entries) can be found for a query.
	ErrNoEntries = sql.ErrNoRows

	// ErrAlreadyExists is returned when a conflict was encountered in the db when doing an insert.
	ErrAlreadyExists = errors.New("already exists")

	// ErrMultipleEntries is returned when multiple entries
	// are found in the db when only one entry is sought.
	ErrMultipleEntries = errors.New("multiple entries")
)

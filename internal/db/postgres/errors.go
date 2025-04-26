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

package postgres

import (
	"fmt"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"github.com/jackc/pgx/v5/pgconn"
)

// processPostgresError processes an error, replacing any
// postgres specific errors with our own error type
func processPostgresError(err error) error {
	// Attempt to cast as postgres
	pgErr, ok := err.(*pgconn.PgError)
	if !ok {
		return err
	}

	// Handle supplied error code:
	// (https://www.postgresql.org/docs/10/errcodes-appendix.html)
	switch pgErr.Code { //nolint
	case "23505" /* unique_violation */ :
		return db.ErrAlreadyExists
	}

	// Wrap the returned error with the code and
	// extended code for easier debugging later.
	return fmt.Errorf("%w (code=%s)", err, pgErr.Code)
}

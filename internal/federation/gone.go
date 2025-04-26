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

package federation

import (
	"context"
	"net/url"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
)

// CheckGone checks if a tombstone exists in the database for AP Actor or Object with the given uri.
func (f *Federator) CheckGone(ctx context.Context, uri *url.URL) (bool, error) {
	return f.db.TombstoneExistsWithURI(ctx, uri.String())
}

// HandleGone puts a tombstone in the database, which marks an AP Actor or Object with the given uri as gone.
func (f *Federator) HandleGone(ctx context.Context, uri *url.URL) error {
	tombstone := &gtsmodel.Tombstone{
		ID:     id.NewULID(),
		Domain: uri.Host,
		URI:    uri.String(),
	}
	return f.db.PutTombstone(ctx, tombstone)
}

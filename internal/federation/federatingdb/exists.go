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

package federatingdb

import (
	"context"
	"net/url"
)

// Exists is an implementation of pub.Database{}.Exists(), optimized specifically for
// the only usecase in which go-fed/activity/pub actually calls it. Do not use otherwise!
func (f *federatingDB) Exists(ctx context.Context, id *url.URL) (exists bool, err error) {
	return f.activityIDs.Has(id.String()), nil
}

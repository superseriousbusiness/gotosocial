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
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type SinBinStatus interface {
	// GetSinBinStatusByID fetches the sin bin status from the database with matching id column.
	GetSinBinStatusByID(ctx context.Context, id string) (*gtsmodel.SinBinStatus, error)

	// GetSinBinStatusByURI fetches the sin bin status from the database with matching uri column.
	GetSinBinStatusByURI(ctx context.Context, uri string) (*gtsmodel.SinBinStatus, error)

	// PutSinBinStatus stores one sin bin status in the database.
	PutSinBinStatus(ctx context.Context, sbStatus *gtsmodel.SinBinStatus) error

	// UpdateSinBinStatus updates one sin bin status in the database.
	UpdateSinBinStatus(ctx context.Context, sbStatus *gtsmodel.SinBinStatus, columns ...string) error

	// DeleteSinBinStatusByID deletes one sin bin status from the database.
	DeleteSinBinStatusByID(ctx context.Context, id string) error
}

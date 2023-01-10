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

package db

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Report handles getting/creation/deletion/updating of user reports/flags.
type Report interface {
	// GetReportByID gets one report by its db id
	GetReportByID(ctx context.Context, id string) (*gtsmodel.Report, Error)
	// PutReport puts the given report in the database.
	PutReport(ctx context.Context, report *gtsmodel.Report) Error
	// UpdateReport updates one report by its db id.
	// The given columns will be updated; if no columns are
	// provided, then all columns will be updated.
	// updated_at will also be updated, no need to pass this
	// as a specific column.
	UpdateReport(ctx context.Context, report *gtsmodel.Report, columns ...string) (*gtsmodel.Report, Error)
	// DeleteReportByID deletes report with the given id.
	DeleteReportByID(ctx context.Context, id string) Error
}

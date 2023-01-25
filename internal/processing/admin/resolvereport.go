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

package admin

import (
	"context"
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) ReportResolve(ctx context.Context, account *gtsmodel.Account, id string, actionTakenComment *string) (*apimodel.AdminReport, gtserror.WithCode) {
	report, err := p.db.GetReportByID(ctx, id)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	columns := []string{
		"action_taken_at",
		"action_taken_by_account_id",
	}

	report.ActionTakenAt = time.Now()
	report.ActionTakenByAccountID = account.ID

	if actionTakenComment != nil {
		report.ActionTaken = *actionTakenComment
		columns = append(columns, "action_taken")
	}

	updatedReport, err := p.db.UpdateReport(ctx, report, columns...)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	apimodelReport, err := p.tc.ReportToAdminAPIReport(ctx, updatedReport, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apimodelReport, nil
}

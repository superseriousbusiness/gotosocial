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

package report

import (
	"context"
	"errors"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (p *processor) Create(ctx context.Context, account *gtsmodel.Account, form *apimodel.ReportCreateRequest) (*apimodel.Report, gtserror.WithCode) {
	if account.ID == form.AccountID {
		err := errors.New("cannot report your own account")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// validate + fetch target account
	targetAccount, err := p.db.GetAccountByID(ctx, form.AccountID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			err = fmt.Errorf("account with ID %s does not exist", form.AccountID)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		err = fmt.Errorf("db error fetching report target account: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// fetch statuses by IDs given in the report form (noop if no statuses given)
	statuses, err := p.db.GetStatuses(ctx, form.StatusIDs)
	if err != nil {
		err = fmt.Errorf("db error fetching report target statuses: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	for _, s := range statuses {
		if s.AccountID != form.AccountID {
			err = fmt.Errorf("status with ID %s does not belong to account %s", s.ID, form.AccountID)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
	}

	reportID := id.NewULID()
	report := &gtsmodel.Report{
		ID:              reportID,
		URI:             uris.GenerateURIForReport(reportID),
		AccountID:       account.ID,
		Account:         account,
		TargetAccountID: form.AccountID,
		TargetAccount:   targetAccount,
		Comment:         form.Comment,
		StatusIDs:       form.StatusIDs,
		Statuses:        statuses,
		Forwarded:       &form.Forward,
	}

	if err := p.db.PutReport(ctx, report); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	p.clientWorker.Queue(messages.FromClientAPI{
		APObjectType:   ap.ObjectProfile,
		APActivityType: ap.ActivityFlag,
		GTSModel:       report,
		OriginAccount:  account,
		TargetAccount:  targetAccount,
	})

	apiReport, err := p.tc.ReportToAPIReport(ctx, report)
	if err != nil {
		err = fmt.Errorf("error converting report to frontend representation: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiReport, nil
}

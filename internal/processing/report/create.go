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

package report

import (
	"context"
	"errors"
	"fmt"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
)

// Create creates one user report / flag, using the provided form parameters.
func (p *Processor) Create(ctx context.Context, account *gtsmodel.Account, form *apimodel.ReportCreateRequest) (*apimodel.Report, gtserror.WithCode) {
	if account.ID == form.AccountID {
		err := errors.New("cannot report your own account")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// validate + fetch target account
	targetAccount, err := p.state.DB.GetAccountByID(ctx, form.AccountID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			err = fmt.Errorf("account with ID %s does not exist", form.AccountID)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		err = fmt.Errorf("db error fetching report target account: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// fetch statuses by IDs given in the report form (noop if no statuses given)
	statuses, err := p.state.DB.GetStatusesByIDs(ctx, form.StatusIDs)
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

	// fetch rules by IDs given in the report form (noop if no rules given)
	rules, err := p.state.DB.GetRulesByIDs(ctx, form.RuleIDs)
	if err != nil {
		err = fmt.Errorf("db error fetching report target rules: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
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
		RuleIDs:         form.RuleIDs,
		Rules:           rules,
		Forwarded:       &form.Forward,
	}

	if err := p.state.DB.PutReport(ctx, report); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActorPerson,
		APActivityType: ap.ActivityFlag,
		GTSModel:       report,
		Origin:         account,
		Target:         targetAccount,
	})

	apiReport, err := p.converter.ReportToAPIReport(ctx, report)
	if err != nil {
		err = fmt.Errorf("error converting report to frontend representation: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiReport, nil
}

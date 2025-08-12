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

package status

import (
	"context"
	"errors"
	"time"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// ScheduledStatusesGetPage returns a page of scheduled statuses authored
// by the requester.
func (p *Processor) ScheduledStatusesGetPage(
	ctx context.Context,
	requester *gtsmodel.Account,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	scheduledStatuses, err := p.state.DB.GetScheduledStatusesForAcct(
		ctx,
		requester.ID,
		page,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting scheduled statuses: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(scheduledStatuses)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	var (
		// Get the lowest and highest
		// ID values, used for paging.
		lo = scheduledStatuses[count-1].ID
		hi = scheduledStatuses[0].ID

		// Best-guess items length.
		items = make([]interface{}, 0, count)
	)

	for _, scheduledStatus := range scheduledStatuses {
		apiScheduledStatus, err := p.converter.ScheduledStatusToAPIScheduledStatus(
			ctx, scheduledStatus,
		)
		if err != nil {
			log.Errorf(ctx, "error converting scheduled status to api scheduled status: %v", err)
			continue
		}

		// Append scheduledStatus to return items.
		items = append(items, apiScheduledStatus)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/scheduled_statuses",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}

// ScheduledStatusesGetOne returns one scheduled
// status with the given ID.
func (p *Processor) ScheduledStatusesGetOne(
	ctx context.Context,
	requester *gtsmodel.Account,
	id string,
) (*apimodel.ScheduledStatus, gtserror.WithCode) {
	scheduledStatus, err := p.state.DB.GetScheduledStatusByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting scheduled status: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if scheduledStatus == nil {
		err := gtserror.New("scheduled status not found")
		return nil, gtserror.NewErrorNotFound(err)
	}

	if scheduledStatus.AccountID != requester.ID {
		err := gtserror.Newf(
			"scheduled status %s is not authored by account %s",
			scheduledStatus.ID, requester.ID,
		)
		return nil, gtserror.NewErrorNotFound(err)
	}

	apiScheduledStatus, err := p.converter.ScheduledStatusToAPIScheduledStatus(
		ctx, scheduledStatus,
	)
	if err != nil {
		err := gtserror.Newf("error converting scheduled status to api scheduled status: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiScheduledStatus, nil
}

func (p *Processor) ScheduledStatusesScheduleAll(ctx context.Context) error {
	// Fetch all pending statuses from the database (barebones models are enough).
	statuses, err := p.state.DB.GetAllScheduledStatuses(gtscontext.SetBarebones(ctx))
	if err != nil {
		return gtserror.Newf("error getting scheduled statuses from db: %w", err)
	}

	var errs gtserror.MultiError

	for _, status := range statuses {
		// Schedule publication of each of the statuses and catch any errors.
		if err := p.ScheduledStatusesSchedulePublication(ctx, status.ID); err != nil {
			errs.Append(err)
		}
	}

	return errs.Combine()
}

func (p *Processor) ScheduledStatusesSchedulePublication(ctx context.Context, statusID string) gtserror.WithCode {
	status, err := p.state.DB.GetScheduledStatusByID(ctx, statusID)

	if err != nil {
		return gtserror.NewErrorNotFound(gtserror.Newf("failed to get scheduled status %s", statusID))
	}

	// Add the given status to the scheduler.
	ok := p.state.Workers.Scheduler.AddOnce(
		status.ID,
		status.ScheduledAt,
		p.onPublish(status.ID),
	)

	if !ok {
		// Failed to add the status to the scheduler, either it was
		// starting / stopping or there already exists a task for status.
		return gtserror.NewErrorInternalError(gtserror.Newf("failed adding status %s to scheduler", status.ID))
	}

	atStr := status.ScheduledAt.Local().Format("Jan _2 2006 15:04:05")
	log.Infof(ctx, "scheduled status publication for %s at '%s'", status.ID, atStr)
	return nil
}

// onPublish returns a callback function to be used by the scheduler on the scheduled date.
func (p *Processor) onPublish(statusID string) func(context.Context, time.Time) {
	return func(ctx context.Context, now time.Time) {
		// Get the latest version of status from database.
		status, err := p.state.DB.GetScheduledStatusByID(ctx, statusID)
		if err != nil {
			log.Errorf(ctx, "error getting status %s from db: %v", statusID, err)
			return
		}

		request := &apimodel.StatusCreateRequest{
			Status:      status.Text,
			MediaIDs:    status.MediaIDs,
			Poll:        nil,
			InReplyToID: status.InReplyToID,
			Sensitive:   *status.Sensitive,
			SpoilerText: status.SpoilerText,
			Visibility:  typeutils.VisToAPIVis(status.Visibility),
			Language:    status.Language,
		}

		if status.Poll.Options != nil && len(status.Poll.Options) > 1 {
			request.Poll = &apimodel.PollRequest{
				Options:    status.Poll.Options,
				ExpiresIn:  status.Poll.ExpiresIn,
				Multiple:   *status.Poll.Multiple,
				HideTotals: *status.Poll.HideTotals,
			}
		}

		_, errWithCode := p.Create(ctx, status.Account, status.Application, request, &statusID)

		if errWithCode != nil {
			log.Errorf(ctx, "could not publish scheduled status: %v", errWithCode.Unwrap())
			return
		}

		err = p.state.DB.DeleteScheduledStatusByID(ctx, statusID)

		if err != nil {
			log.Error(ctx, err)
		}
	}
}

// Update scheduled status schedule data
func (p *Processor) ScheduledStatusesUpdate(
	ctx context.Context,
	requester *gtsmodel.Account,
	id string,
	scheduledAt *time.Time,
) (*apimodel.ScheduledStatus, gtserror.WithCode) {
	scheduledStatus, err := p.state.DB.GetScheduledStatusByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting scheduled status: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if scheduledStatus == nil {
		err := gtserror.New("scheduled status not found")
		return nil, gtserror.NewErrorNotFound(err)
	}

	if scheduledStatus.AccountID != requester.ID {
		err := gtserror.Newf(
			"scheduled status %s is not authored by account %s",
			scheduledStatus.ID, requester.ID,
		)
		return nil, gtserror.NewErrorNotFound(err)
	}

	if errWithCode := p.validateScheduledStatusLimits(ctx, requester.ID, scheduledAt, &scheduledStatus.ScheduledAt); errWithCode != nil {
		return nil, errWithCode
	}

	scheduledStatus.ScheduledAt = *scheduledAt
	err = p.state.DB.UpdateScheduledStatusScheduledDate(ctx, scheduledStatus, scheduledAt)

	if err != nil {
		err := gtserror.Newf("db error getting scheduled status: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	ok := p.state.Workers.Scheduler.Cancel(id)

	if !ok {
		err := gtserror.Newf("failed to cancel scheduled status")
		return nil, gtserror.NewErrorInternalError(err)
	}

	err = p.ScheduledStatusesSchedulePublication(ctx, id)

	if err != nil {
		err := gtserror.Newf("error scheduling status: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiScheduledStatus, err := p.converter.ScheduledStatusToAPIScheduledStatus(
		ctx, scheduledStatus,
	)
	if err != nil {
		err := gtserror.Newf("error converting scheduled status to api req: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiScheduledStatus, nil
}

// Cancel a scheduled status
func (p *Processor) ScheduledStatusesDelete(ctx context.Context, requester *gtsmodel.Account, id string) gtserror.WithCode {
	scheduledStatus, err := p.state.DB.GetScheduledStatusByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting scheduled status: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	if scheduledStatus == nil {
		err := gtserror.New("scheduled status not found")
		return gtserror.NewErrorNotFound(err)
	}

	if scheduledStatus.AccountID != requester.ID {
		err := gtserror.Newf(
			"scheduled status %s is not authored by account %s",
			scheduledStatus.ID, requester.ID,
		)
		return gtserror.NewErrorNotFound(err)
	}

	ok := p.state.Workers.Scheduler.Cancel(id)

	if !ok {
		err := gtserror.Newf("failed to cancel scheduled status")
		return gtserror.NewErrorInternalError(err)
	}

	err = p.state.DB.DeleteScheduledStatusByID(ctx, id)

	if err != nil {
		err := gtserror.Newf("db error deleting scheduled status: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}

func (p *Processor) validateScheduledStatusLimits(ctx context.Context, acctID string, scheduledAt *time.Time, prevScheduledAt *time.Time) gtserror.WithCode {
	// Skip check when the scheduled status already exists and the day stays the same
	if prevScheduledAt != nil {
		y1, m1, d1 := scheduledAt.Date()
		y2, m2, d2 := prevScheduledAt.Date()

		if y1 == y2 && m1 == m2 && d1 == d2 {
			return nil
		}
	}

	scheduledDaily, err := p.state.DB.GetScheduledStatusesCountForAcct(ctx, acctID, scheduledAt)

	if err != nil {
		err := gtserror.Newf("error getting scheduled statuses count for day: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	if max := config.GetScheduledStatusesMaxDaily(); scheduledDaily >= max {
		err := gtserror.Newf("scheduled statuses count for day is at the limit (%d)", max)
		return gtserror.NewErrorUnprocessableEntity(err)
	}

	// Skip total check when editing an existing scheduled status
	if prevScheduledAt != nil {
		return nil
	}

	scheduledTotal, err := p.state.DB.GetScheduledStatusesCountForAcct(ctx, acctID, nil)

	if err != nil {
		err := gtserror.Newf("error getting total scheduled statuses count: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	if max := config.GetScheduledStatusesMaxTotal(); scheduledTotal >= max {
		err := gtserror.Newf("total scheduled statuses count is at the limit (%d)", max)
		return gtserror.NewErrorUnprocessableEntity(err)
	}

	return nil
}

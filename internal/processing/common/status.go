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

package common

import (
	"context"
	"errors"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// GetOwnStatus fetches the given status with ID,
// and ensures that it belongs to given requester.
func (p *Processor) GetOwnStatus(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetID string,
) (
	*gtsmodel.Status,
	gtserror.WithCode,
) {
	target, err := p.state.DB.GetStatusByID(ctx, targetID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error getting from db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	switch {
	case target == nil:
		const text = "target status not found"
		return nil, gtserror.NewErrorNotFound(
			errors.New(text),
			text,
		)

	case target.AccountID != requester.ID:
		return nil, gtserror.NewErrorNotFound(
			errors.New("status does not belong to requester"),
			"target status not found",
		)
	}

	return target, nil
}

// GetTargetStatusBy fetches the target status with db load
// function, given the authorized (or, nil) requester's
// account. This returns an approprate gtserror.WithCode
// accounting for not found and visibility to requester.
//
// window can be used to force refresh of the target if it's
// deemed to be stale. Falls back to default window if nil.
func (p *Processor) GetTargetStatusBy(
	ctx context.Context,
	requester *gtsmodel.Account,
	getTargetFromDB func() (*gtsmodel.Status, error),
	window *dereferencing.FreshnessWindow,
) (
	status *gtsmodel.Status,
	visible bool,
	errWithCode gtserror.WithCode,
) {
	// Fetch the target status from db.
	target, err := getTargetFromDB()
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error getting from db: %w", err)
		return nil, false, gtserror.NewErrorInternalError(err)
	}

	if target == nil {
		// DB loader could not find status in database.
		const text = "target status not found"
		return nil, false, gtserror.NewErrorNotFound(
			errors.New(text),
			text,
		)
	}

	// Check whether target status is visible to requesting account.
	visible, err = p.visFilter.StatusVisible(ctx, requester, target)
	if err != nil {
		err := gtserror.Newf("error checking visibility: %w", err)
		return nil, false, gtserror.NewErrorInternalError(err)
	}

	if requester != nil && visible {
		// Only refresh status if visible to requester,
		// and there is *authorized* requester to prevent
		// a possible DOS vector for unauthorized clients.
		latest, _, err := p.federator.RefreshStatus(ctx,
			requester.Username,
			target,
			nil,
			window,
		)
		if err != nil {
			log.Errorf(ctx, "error refreshing target %s: %v", target.URI, err)
			return target, visible, nil
		}

		// Set latest.
		target = latest
	}

	return target, visible, nil
}

// GetVisibleTargetStatus calls GetTargetStatusBy(),
// but converts a non-visible result to not-found error.
//
// window can be used to force refresh of the target if it's
// deemed to be stale. Falls back to default window if nil.
func (p *Processor) GetVisibleTargetStatusBy(
	ctx context.Context,
	requester *gtsmodel.Account,
	getTargetFromDB func() (*gtsmodel.Status, error),
	window *dereferencing.FreshnessWindow,
) (
	status *gtsmodel.Status,
	errWithCode gtserror.WithCode,
) {
	// Fetch the target status by ID from the database.
	target, visible, errWithCode := p.GetTargetStatusBy(ctx,
		requester,
		getTargetFromDB,
		window,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if !visible {
		// Target should not be seen by requester.
		const text = "target status not found"
		return nil, gtserror.NewErrorNotFound(
			errors.New(text),
			text,
		)
	}

	return target, nil
}

// GetVisibleTargetStatus calls GetVisibleTargetStatusBy(),
// passing in a database function that fetches by status ID.
//
// window can be used to force refresh of the target if it's
// deemed to be stale. Falls back to default window if nil.
func (p *Processor) GetVisibleTargetStatus(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetID string,
	window *dereferencing.FreshnessWindow,
) (
	status *gtsmodel.Status,
	errWithCode gtserror.WithCode,
) {
	return p.GetVisibleTargetStatusBy(ctx, requester, func() (*gtsmodel.Status, error) {
		return p.state.DB.GetStatusByID(ctx, targetID)
	}, window)
}

// UnwrapIfBoost "unwraps" the given status if
// it's a boost wrapper, by returning the boosted
// status it targets (pending visibility checks).
//
// Just returns the input status if it's not a boost.
func (p *Processor) UnwrapIfBoost(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) (*gtsmodel.Status, gtserror.WithCode) {
	if status.BoostOfID == "" {
		return status, nil
	}
	return p.GetVisibleTargetStatus(ctx,
		requester,
		status.BoostOfID,
		nil,
	)
}

// GetAPIStatus fetches the appropriate API status model for target.
func (p *Processor) GetAPIStatus(
	ctx context.Context,
	requester *gtsmodel.Account,
	target *gtsmodel.Status,
) (
	apiStatus *apimodel.Status,
	errWithCode gtserror.WithCode,
) {
	apiStatus, err := p.converter.StatusToAPIStatus(ctx,
		target,
		requester,
		statusfilter.FilterContextNone,
		nil,
		nil,
	)
	if err != nil {
		err := gtserror.Newf("error converting: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	return apiStatus, nil
}

// GetVisibleAPIStatuses converts a slice of statuses to API
// model statuses, filtering according to visibility to requester
// along with given filter context, filters and user mutes.
//
// Please note that all errors will be logged at ERROR level,
// but will not be returned. Callers are likely to run into
// show-stopping errors in the lead-up to this function.
func (p *Processor) GetVisibleAPIStatuses(
	ctx context.Context,
	requester *gtsmodel.Account,
	statuses []*gtsmodel.Status,
	filterContext statusfilter.FilterContext,
	filters []*gtsmodel.Filter,
	userMutes []*gtsmodel.UserMute,
) []apimodel.Status {

	// Start new log entry with
	// the calling function name
	// as a field in each entry.
	l := log.WithContext(ctx).
		WithField("caller", log.Caller(3))

	// Compile mutes to useable user mutes for type converter.
	compUserMutes := usermute.NewCompiledUserMuteList(userMutes)

	// Iterate filtered statuses for conversion to API model.
	apiStatuses := make([]apimodel.Status, 0, len(statuses))
	for _, status := range statuses {

		// Check whether status is visible to requester.
		visible, err := p.visFilter.StatusVisible(ctx,
			requester,
			status,
		)
		if err != nil {
			l.Errorf("error checking visibility: %v", err)
			continue
		}

		if !visible {
			continue
		}

		// Convert to API status, taking mute / filter into account.
		apiStatus, err := p.converter.StatusToAPIStatus(ctx,
			status,
			requester,
			filterContext,
			filters,
			compUserMutes,
		)
		if err != nil && !errors.Is(err, statusfilter.ErrHideStatus) {
			l.Errorf("error converting: %v", err)
			continue
		}

		if apiStatus == nil {
			// Status was
			// filtered out.
			continue
		}

		// Append converted status to return slice.
		apiStatuses = append(apiStatuses, *apiStatus)
	}

	return apiStatuses
}

// InvalidateTimelinedStatus is a shortcut function for invalidating the cached
// representation one status in the home timeline and all list timelines of the
// given accountID. It should only be called in cases where a status update
// does *not* need to be passed into the processor via the worker queue, since
// such invalidation will, in that case, be handled by the processor instead.
func (p *Processor) InvalidateTimelinedStatus(ctx context.Context, accountID string, statusID string) error {
	// Get lists first + bail if this fails.
	lists, err := p.state.DB.GetListsByAccountID(ctx, accountID)
	if err != nil {
		return gtserror.Newf("db error getting lists for account %s: %w", accountID, err)
	}

	// Start new log entry with
	// the above calling func's name.
	l := log.
		WithContext(ctx).
		WithField("caller", log.Caller(3)).
		WithField("accountID", accountID).
		WithField("statusID", statusID)

	// Unprepare item from home + list timelines, just log
	// if something goes wrong since this is not a showstopper.

	if err := p.state.Timelines.Home.UnprepareItem(ctx, accountID, statusID); err != nil {
		l.Errorf("error unpreparing item from home timeline: %v", err)
	}

	for _, list := range lists {
		if err := p.state.Timelines.List.UnprepareItem(ctx, list.ID, statusID); err != nil {
			l.Errorf("error unpreparing item from list timeline %s: %v", list.ID, err)
		}
	}

	return nil
}

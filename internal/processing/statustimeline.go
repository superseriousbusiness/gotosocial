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

package processing

import (
	"context"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

const boostReinsertionDepth = 50

// StatusGrabFunction returns a function that satisfies the GrabFunction interface in internal/timeline.
func StatusGrabFunction(database db.DB) timeline.GrabFunction {
	return func(ctx context.Context, timelineAccountID string, maxID string, sinceID string, minID string, limit int) ([]timeline.Timelineable, bool, error) {
		statuses, err := database.GetHomeTimeline(ctx, timelineAccountID, maxID, sinceID, minID, limit, false)
		if err != nil {
			if err == db.ErrNoEntries {
				return nil, true, nil // we just don't have enough statuses left in the db so return stop = true
			}
			return nil, false, fmt.Errorf("statusGrabFunction: error getting statuses from db: %s", err)
		}

		items := []timeline.Timelineable{}
		for _, s := range statuses {
			items = append(items, s)
		}

		return items, false, nil
	}
}

// StatusFilterFunction returns a function that satisfies the FilterFunction interface in internal/timeline.
func StatusFilterFunction(database db.DB, filter visibility.Filter) timeline.FilterFunction {
	return func(ctx context.Context, timelineAccountID string, item timeline.Timelineable) (shouldIndex bool, err error) {
		status, ok := item.(*gtsmodel.Status)
		if !ok {
			return false, errors.New("statusFilterFunction: could not convert item to *gtsmodel.Status")
		}

		requestingAccount, err := database.GetAccountByID(ctx, timelineAccountID)
		if err != nil {
			return false, fmt.Errorf("statusFilterFunction: error getting account with id %s", timelineAccountID)
		}

		timelineable, err := filter.StatusHometimelineable(ctx, status, requestingAccount)
		if err != nil {
			log.Warnf("error checking hometimelineability of status %s for account %s: %s", status.ID, timelineAccountID, err)
		}

		return timelineable, nil // we don't return the error here because we want to just skip this item if something goes wrong
	}
}

// StatusPrepareFunction returns a function that satisfies the PrepareFunction interface in internal/timeline.
func StatusPrepareFunction(database db.DB, tc typeutils.TypeConverter) timeline.PrepareFunction {
	return func(ctx context.Context, timelineAccountID string, itemID string) (timeline.Preparable, error) {
		status, err := database.GetStatusByID(ctx, itemID)
		if err != nil {
			return nil, fmt.Errorf("statusPrepareFunction: error getting status with id %s", itemID)
		}

		requestingAccount, err := database.GetAccountByID(ctx, timelineAccountID)
		if err != nil {
			return nil, fmt.Errorf("statusPrepareFunction: error getting account with id %s", timelineAccountID)
		}

		return tc.StatusToAPIStatus(ctx, status, requestingAccount)
	}
}

// StatusSkipInsertFunction returns a function that satisifes the SkipInsertFunction interface in internal/timeline.
func StatusSkipInsertFunction() timeline.SkipInsertFunction {
	return func(
		ctx context.Context,
		newItemID string,
		newItemAccountID string,
		newItemBoostOfID string,
		newItemBoostOfAccountID string,
		nextItemID string,
		nextItemAccountID string,
		nextItemBoostOfID string,
		nextItemBoostOfAccountID string,
		depth int,
	) (bool, error) {
		// make sure we don't insert a duplicate
		if newItemID == nextItemID {
			return true, nil
		}

		// check if it's a boost
		if newItemBoostOfID != "" {
			// skip if we've recently put another boost of this status in the timeline
			if newItemBoostOfID == nextItemBoostOfID {
				if depth < boostReinsertionDepth {
					return true, nil
				}
			}

			// skip if we've recently put the original status in the timeline
			if newItemBoostOfID == nextItemID {
				if depth < boostReinsertionDepth {
					return true, nil
				}
			}
		}

		// insert the item
		return false, nil
	}
}

func (p *processor) HomeTimelineGet(ctx context.Context, authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, local bool) (*apimodel.PageableResponse, gtserror.WithCode) {
	preparedItems, err := p.statusTimelines.GetTimeline(ctx, authed.Account.ID, maxID, sinceID, minID, limit, local)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(preparedItems)

	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	items := []interface{}{}
	nextMaxIDValue := ""
	prevMinIDValue := ""
	for i, item := range preparedItems {
		if i == count-1 {
			nextMaxIDValue = item.GetID()
		}

		if i == 0 {
			prevMinIDValue = item.GetID()
		}
		items = append(items, item)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "api/v1/timelines/home",
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}

func (p *processor) PublicTimelineGet(ctx context.Context, authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, local bool) (*apimodel.PageableResponse, gtserror.WithCode) {
	statuses, err := p.db.GetPublicTimeline(ctx, maxID, sinceID, minID, limit, local)
	if err != nil {
		if err == db.ErrNoEntries {
			// there are just no entries left
			return util.EmptyPageableResponse(), nil
		}
		// there's an actual error
		return nil, gtserror.NewErrorInternalError(err)
	}

	filtered, err := p.filterPublicStatuses(ctx, authed, statuses)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(filtered)

	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	items := []interface{}{}
	nextMaxIDValue := ""
	prevMinIDValue := ""
	for i, item := range filtered {
		if i == count-1 {
			nextMaxIDValue = item.GetID()
		}

		if i == 0 {
			prevMinIDValue = item.GetID()
		}
		items = append(items, item)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "api/v1/timelines/public",
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}

func (p *processor) FavedTimelineGet(ctx context.Context, authed *oauth.Auth, maxID string, minID string, limit int) (*apimodel.PageableResponse, gtserror.WithCode) {
	statuses, nextMaxID, prevMinID, err := p.db.GetFavedTimeline(ctx, authed.Account.ID, maxID, minID, limit)
	if err != nil {
		if err == db.ErrNoEntries {
			// there are just no entries left
			return util.EmptyPageableResponse(), nil
		}
		// there's an actual error
		return nil, gtserror.NewErrorInternalError(err)
	}

	filtered, err := p.filterFavedStatuses(ctx, authed, statuses)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if len(filtered) == 0 {
		return util.EmptyPageableResponse(), nil
	}

	items := []interface{}{}
	for _, item := range filtered {
		items = append(items, item)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "api/v1/favourites",
		NextMaxIDValue: nextMaxID,
		PrevMinIDValue: prevMinID,
		Limit:          limit,
	})
}

func (p *processor) filterPublicStatuses(ctx context.Context, authed *oauth.Auth, statuses []*gtsmodel.Status) ([]*apimodel.Status, error) {
	apiStatuses := []*apimodel.Status{}
	for _, s := range statuses {
		targetAccount := &gtsmodel.Account{}
		if err := p.db.GetByID(ctx, s.AccountID, targetAccount); err != nil {
			if err == db.ErrNoEntries {
				log.Debugf("filterPublicStatuses: skipping status %s because account %s can't be found in the db", s.ID, s.AccountID)
				continue
			}
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("filterPublicStatuses: error getting status author: %s", err))
		}

		timelineable, err := p.filter.StatusPublictimelineable(ctx, s, authed.Account)
		if err != nil {
			log.Debugf("filterPublicStatuses: skipping status %s because of an error checking status visibility: %s", s.ID, err)
			continue
		}
		if !timelineable {
			continue
		}

		apiStatus, err := p.tc.StatusToAPIStatus(ctx, s, authed.Account)
		if err != nil {
			log.Debugf("filterPublicStatuses: skipping status %s because it couldn't be converted to its api representation: %s", s.ID, err)
			continue
		}

		apiStatuses = append(apiStatuses, apiStatus)
	}

	return apiStatuses, nil
}

func (p *processor) filterFavedStatuses(ctx context.Context, authed *oauth.Auth, statuses []*gtsmodel.Status) ([]*apimodel.Status, error) {
	apiStatuses := []*apimodel.Status{}
	for _, s := range statuses {
		targetAccount := &gtsmodel.Account{}
		if err := p.db.GetByID(ctx, s.AccountID, targetAccount); err != nil {
			if err == db.ErrNoEntries {
				log.Debugf("filterFavedStatuses: skipping status %s because account %s can't be found in the db", s.ID, s.AccountID)
				continue
			}
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("filterPublicStatuses: error getting status author: %s", err))
		}

		timelineable, err := p.filter.StatusVisible(ctx, s, authed.Account)
		if err != nil {
			log.Debugf("filterFavedStatuses: skipping status %s because of an error checking status visibility: %s", s.ID, err)
			continue
		}
		if !timelineable {
			continue
		}

		apiStatus, err := p.tc.StatusToAPIStatus(ctx, s, authed.Account)
		if err != nil {
			log.Debugf("filterFavedStatuses: skipping status %s because it couldn't be converted to its api representation: %s", s.ID, err)
			continue
		}

		apiStatuses = append(apiStatuses, apiStatus)
	}

	return apiStatuses, nil
}

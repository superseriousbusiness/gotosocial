/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"fmt"
	"github.com/sirupsen/logrus"
	"net/url"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) packageStatusResponse(statuses []*apimodel.Status, path string, nextMaxID string, prevMinID string, limit int) (*apimodel.StatusTimelineResponse, gtserror.WithCode) {
	resp := &apimodel.StatusTimelineResponse{
		Statuses: []*apimodel.Status{},
	}
	resp.Statuses = statuses

	// prepare the next and previous links
	if len(statuses) != 0 {
		nextLink := &url.URL{
			Scheme:   p.config.Protocol,
			Host:     p.config.Host,
			Path:     path,
			RawQuery: fmt.Sprintf("limit=%d&max_id=%s", limit, nextMaxID),
		}
		next := fmt.Sprintf("<%s>; rel=\"next\"", nextLink.String())

		prevLink := &url.URL{
			Scheme:   p.config.Protocol,
			Host:     p.config.Host,
			Path:     path,
			RawQuery: fmt.Sprintf("limit=%d&min_id=%s", limit, prevMinID),
		}
		prev := fmt.Sprintf("<%s>; rel=\"prev\"", prevLink.String())
		resp.LinkHeader = fmt.Sprintf("%s, %s", next, prev)
	}

	return resp, nil
}

func (p *processor) HomeTimelineGet(ctx context.Context, authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, local bool) (*apimodel.StatusTimelineResponse, gtserror.WithCode) {
	statuses, err := p.timelineManager.HomeTimeline(ctx, authed.Account.ID, maxID, sinceID, minID, limit, local)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if len(statuses) == 0 {
		return &apimodel.StatusTimelineResponse{
			Statuses: []*apimodel.Status{},
		}, nil
	}

	return p.packageStatusResponse(statuses, "api/v1/timelines/home", statuses[len(statuses)-1].ID, statuses[0].ID, limit)
}

func (p *processor) PublicTimelineGet(ctx context.Context, authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, local bool) (*apimodel.StatusTimelineResponse, gtserror.WithCode) {
	statuses, err := p.db.GetPublicTimeline(ctx, authed.Account.ID, maxID, sinceID, minID, limit, local)
	if err != nil {
		if err == db.ErrNoEntries {
			// there are just no entries left
			return &apimodel.StatusTimelineResponse{
				Statuses: []*apimodel.Status{},
			}, nil
		}
		// there's an actual error
		return nil, gtserror.NewErrorInternalError(err)
	}

	s, err := p.filterPublicStatuses(ctx, authed, statuses)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if len(s) == 0 {
		return &apimodel.StatusTimelineResponse{
			Statuses: []*apimodel.Status{},
		}, nil
	}

	return p.packageStatusResponse(s, "api/v1/timelines/public", s[len(s)-1].ID, s[0].ID, limit)
}

func (p *processor) FavedTimelineGet(ctx context.Context, authed *oauth.Auth, maxID string, minID string, limit int) (*apimodel.StatusTimelineResponse, gtserror.WithCode) {
	statuses, nextMaxID, prevMinID, err := p.db.GetFavedTimeline(ctx, authed.Account.ID, maxID, minID, limit)
	if err != nil {
		if err == db.ErrNoEntries {
			// there are just no entries left
			return &apimodel.StatusTimelineResponse{
				Statuses: []*apimodel.Status{},
			}, nil
		}
		// there's an actual error
		return nil, gtserror.NewErrorInternalError(err)
	}

	s, err := p.filterFavedStatuses(ctx, authed, statuses)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if len(s) == 0 {
		return &apimodel.StatusTimelineResponse{
			Statuses: []*apimodel.Status{},
		}, nil
	}

	return p.packageStatusResponse(s, "api/v1/favourites", nextMaxID, prevMinID, limit)
}

func (p *processor) filterPublicStatuses(ctx context.Context, authed *oauth.Auth, statuses []*gtsmodel.Status) ([]*apimodel.Status, error) {
	l := logrus.WithField("func", "filterPublicStatuses")

	apiStatuses := []*apimodel.Status{}
	for _, s := range statuses {
		targetAccount := &gtsmodel.Account{}
		if err := p.db.GetByID(ctx, s.AccountID, targetAccount); err != nil {
			if err == db.ErrNoEntries {
				l.Debugf("filterPublicStatuses: skipping status %s because account %s can't be found in the db", s.ID, s.AccountID)
				continue
			}
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("filterPublicStatuses: error getting status author: %s", err))
		}

		timelineable, err := p.filter.StatusPublictimelineable(ctx, s, authed.Account)
		if err != nil {
			l.Debugf("filterPublicStatuses: skipping status %s because of an error checking status visibility: %s", s.ID, err)
			continue
		}
		if !timelineable {
			continue
		}

		apiStatus, err := p.tc.StatusToAPIStatus(ctx, s, authed.Account)
		if err != nil {
			l.Debugf("filterPublicStatuses: skipping status %s because it couldn't be converted to its api representation: %s", s.ID, err)
			continue
		}

		apiStatuses = append(apiStatuses, apiStatus)
	}

	return apiStatuses, nil
}

func (p *processor) filterFavedStatuses(ctx context.Context, authed *oauth.Auth, statuses []*gtsmodel.Status) ([]*apimodel.Status, error) {
	l := logrus.WithField("func", "filterFavedStatuses")

	apiStatuses := []*apimodel.Status{}
	for _, s := range statuses {
		targetAccount := &gtsmodel.Account{}
		if err := p.db.GetByID(ctx, s.AccountID, targetAccount); err != nil {
			if err == db.ErrNoEntries {
				l.Debugf("filterFavedStatuses: skipping status %s because account %s can't be found in the db", s.ID, s.AccountID)
				continue
			}
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("filterPublicStatuses: error getting status author: %s", err))
		}

		timelineable, err := p.filter.StatusVisible(ctx, s, authed.Account)
		if err != nil {
			l.Debugf("filterFavedStatuses: skipping status %s because of an error checking status visibility: %s", s.ID, err)
			continue
		}
		if !timelineable {
			continue
		}

		apiStatus, err := p.tc.StatusToAPIStatus(ctx, s, authed.Account)
		if err != nil {
			l.Debugf("filterFavedStatuses: skipping status %s because it couldn't be converted to its api representation: %s", s.ID, err)
			continue
		}

		apiStatuses = append(apiStatuses, apiStatus)
	}

	return apiStatuses, nil
}

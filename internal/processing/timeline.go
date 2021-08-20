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
	"fmt"
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

func (p *processor) HomeTimelineGet(authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, local bool) (*apimodel.StatusTimelineResponse, gtserror.WithCode) {
	statuses, err := p.timelineManager.HomeTimeline(authed.Account.ID, maxID, sinceID, minID, limit, local)
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

func (p *processor) PublicTimelineGet(authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, local bool) (*apimodel.StatusTimelineResponse, gtserror.WithCode) {
	statuses, err := p.db.GetPublicTimeline(authed.Account.ID, maxID, sinceID, minID, limit, local)
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

	s, err := p.filterPublicStatuses(authed, statuses)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.packageStatusResponse(s, "api/v1/timelines/public", s[len(s)-1].ID, s[0].ID, limit)
}

func (p *processor) FavedTimelineGet(authed *oauth.Auth, maxID string, minID string, limit int) (*apimodel.StatusTimelineResponse, gtserror.WithCode) {
	statuses, nextMaxID, prevMinID, err := p.db.GetFavedTimeline(authed.Account.ID, maxID, minID, limit)
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

	s, err := p.filterFavedStatuses(authed, statuses)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.packageStatusResponse(s, "api/v1/favourites", nextMaxID, prevMinID, limit)
}

func (p *processor) filterPublicStatuses(authed *oauth.Auth, statuses []*gtsmodel.Status) ([]*apimodel.Status, error) {
	l := p.log.WithField("func", "filterPublicStatuses")

	apiStatuses := []*apimodel.Status{}
	for _, s := range statuses {
		targetAccount := &gtsmodel.Account{}
		if err := p.db.GetByID(s.AccountID, targetAccount); err != nil {
			if err == db.ErrNoEntries {
				l.Debugf("filterPublicStatuses: skipping status %s because account %s can't be found in the db", s.ID, s.AccountID)
				continue
			}
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("filterPublicStatuses: error getting status author: %s", err))
		}

		timelineable, err := p.filter.StatusPublictimelineable(s, authed.Account)
		if err != nil {
			l.Debugf("filterPublicStatuses: skipping status %s because of an error checking status visibility: %s", s.ID, err)
			continue
		}
		if !timelineable {
			continue
		}

		apiStatus, err := p.tc.StatusToMasto(s, authed.Account)
		if err != nil {
			l.Debugf("filterPublicStatuses: skipping status %s because it couldn't be converted to its mastodon representation: %s", s.ID, err)
			continue
		}

		apiStatuses = append(apiStatuses, apiStatus)
	}

	return apiStatuses, nil
}

func (p *processor) filterFavedStatuses(authed *oauth.Auth, statuses []*gtsmodel.Status) ([]*apimodel.Status, error) {
	l := p.log.WithField("func", "filterFavedStatuses")

	apiStatuses := []*apimodel.Status{}
	for _, s := range statuses {
		targetAccount := &gtsmodel.Account{}
		if err := p.db.GetByID(s.AccountID, targetAccount); err != nil {
			if err == db.ErrNoEntries {
				l.Debugf("filterFavedStatuses: skipping status %s because account %s can't be found in the db", s.ID, s.AccountID)
				continue
			}
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("filterPublicStatuses: error getting status author: %s", err))
		}

		timelineable, err := p.filter.StatusVisible(s, authed.Account)
		if err != nil {
			l.Debugf("filterFavedStatuses: skipping status %s because of an error checking status visibility: %s", s.ID, err)
			continue
		}
		if !timelineable {
			continue
		}

		apiStatus, err := p.tc.StatusToMasto(s, authed.Account)
		if err != nil {
			l.Debugf("filterFavedStatuses: skipping status %s because it couldn't be converted to its mastodon representation: %s", s.ID, err)
			continue
		}

		apiStatuses = append(apiStatuses, apiStatus)
	}

	return apiStatuses, nil
}

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
	"sync"

	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) HomeTimelineGet(authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, local bool) (*apimodel.StatusTimelineResponse, gtserror.WithCode) {
	resp := &apimodel.StatusTimelineResponse{
		Statuses: []*apimodel.Status{},
	}

	apiStatuses, err := p.timelineManager.HomeTimeline(authed.Account.ID, maxID, sinceID, minID, limit, local)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	resp.Statuses = apiStatuses

	// prepare the next and previous links
	if len(apiStatuses) != 0 {
		nextLink := &url.URL{
			Scheme:   p.config.Protocol,
			Host:     p.config.Host,
			Path:     "/api/v1/timelines/home",
			RawPath:  url.PathEscape("api/v1/timelines/home"),
			RawQuery: fmt.Sprintf("limit=%d&max_id=%s", limit, apiStatuses[len(apiStatuses)-1].ID),
		}
		next := fmt.Sprintf("<%s>; rel=\"next\"", nextLink.String())

		prevLink := &url.URL{
			Scheme:   p.config.Protocol,
			Host:     p.config.Host,
			Path:     "/api/v1/timelines/home",
			RawQuery: fmt.Sprintf("limit=%d&min_id=%s", limit, apiStatuses[0].ID),
		}
		prev := fmt.Sprintf("<%s>; rel=\"prev\"", prevLink.String())
		resp.LinkHeader = fmt.Sprintf("%s, %s", next, prev)
	}

	return resp, nil
}

func (p *processor) PublicTimelineGet(authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, local bool) ([]*apimodel.Status, gtserror.WithCode) {
	statuses, err := p.db.GetPublicTimelineForAccount(authed.Account.ID, maxID, sinceID, minID, limit, local)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	s, err := p.filterStatuses(authed, statuses)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return s, nil
}

func (p *processor) filterStatuses(authed *oauth.Auth, statuses []*gtsmodel.Status) ([]*apimodel.Status, error) {
	l := p.log.WithField("func", "filterStatuses")

	apiStatuses := []*apimodel.Status{}
	for _, s := range statuses {
		targetAccount := &gtsmodel.Account{}
		if err := p.db.GetByID(s.AccountID, targetAccount); err != nil {
			if _, ok := err.(db.ErrNoEntries); ok {
				l.Debugf("skipping status %s because account %s can't be found in the db", s.ID, s.AccountID)
				continue
			}
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("HomeTimelineGet: error getting status author: %s", err))
		}

		timelineable, err := p.filter.StatusHometimelineable(s, authed.Account)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("HomeTimelineGet: error checking status visibility: %s", err))
		}
		if !timelineable {
			continue
		}

		apiStatus, err := p.tc.StatusToMasto(s, authed.Account)
		if err != nil {
			l.Debugf("skipping status %s because it couldn't be converted to its mastodon representation: %s", s.ID, err)
			continue
		}

		apiStatuses = append(apiStatuses, apiStatus)
	}

	return apiStatuses, nil
}

func (p *processor) initTimelines() error {
	// get all local accounts (ie., domain = nil) that aren't suspended (suspended_at = nil)
	localAccounts := []*gtsmodel.Account{}
	where := []db.Where{
		{
			Key: "domain", Value: nil,
		},
		{
			Key: "suspended_at", Value: nil,
		},
	}
	if err := p.db.GetWhere(where, &localAccounts); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return nil
		}
		return fmt.Errorf("initTimelines: db error initializing timelines: %s", err)
	}

	// we want to wait until all timelines are populated so created a waitgroup here
	wg := &sync.WaitGroup{}
	wg.Add(len(localAccounts))

	for _, localAccount := range localAccounts {
		// to save time we can populate the timelines asynchronously
		// this will go heavy on the database, but since we're not actually serving yet it doesn't really matter
		go p.initTimelineFor(localAccount, wg)
	}

	// wait for all timelines to be populated before we exit
	wg.Wait()
	return nil
}

func (p *processor) initTimelineFor(account *gtsmodel.Account, wg *sync.WaitGroup) {
	defer wg.Done()

	l := p.log.WithFields(logrus.Fields{
		"func":      "initTimelineFor",
		"accountID": account.ID,
	})

	desiredIndexLength := p.timelineManager.GetDesiredIndexLength()

	statuses, err := p.db.GetStatusesWhereFollowing(account.ID, "", "", "", desiredIndexLength, false)
	if err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			l.Error(fmt.Errorf("initTimelineFor: error getting statuses: %s", err))
		}
		return
	}
	p.indexAndIngest(statuses, account, desiredIndexLength)

	lengthNow := p.timelineManager.GetIndexedLength(account.ID)
	if lengthNow < desiredIndexLength {
		// try and get more posts from the last ID onwards
		rearmostStatusID, err := p.timelineManager.GetOldestIndexedID(account.ID)
		if err != nil {
			l.Error(fmt.Errorf("initTimelineFor: error getting id of rearmost status: %s", err))
			return
		}

		if rearmostStatusID != "" {
			moreStatuses, err := p.db.GetStatusesWhereFollowing(account.ID, rearmostStatusID, "", "", desiredIndexLength/2, false)
			if err != nil {
				l.Error(fmt.Errorf("initTimelineFor: error getting more statuses: %s", err))
				return
			}
			p.indexAndIngest(moreStatuses, account, desiredIndexLength)
		}
	}

	l.Debugf("prepared timeline of length %d for account %s", lengthNow, account.ID)
}

func (p *processor) indexAndIngest(statuses []*gtsmodel.Status, timelineAccount *gtsmodel.Account, desiredIndexLength int) {
	l := p.log.WithFields(logrus.Fields{
		"func":      "indexAndIngest",
		"accountID": timelineAccount.ID,
	})

	for _, s := range statuses {
		timelineable, err := p.filter.StatusHometimelineable(s, timelineAccount)
		if err != nil {
			l.Error(fmt.Errorf("initTimelineFor: error checking home timelineability of status %s: %s", s.ID, err))
			continue
		}
		if timelineable {
			if err := p.timelineManager.Ingest(s, timelineAccount.ID); err != nil {
				l.Error(fmt.Errorf("initTimelineFor: error ingesting status %s: %s", s.ID, err))
				continue
			}

			// check if we have enough posts now and return if we do
			if p.timelineManager.GetIndexedLength(timelineAccount.ID) >= desiredIndexLength {
				return
			}
		}
	}
}

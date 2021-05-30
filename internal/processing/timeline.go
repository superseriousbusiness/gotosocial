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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) HomeTimelineGet(authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, local bool) ([]apimodel.Status, ErrorWithCode) {
	l := p.log.WithField("func", "HomeTimelineGet")

	statuses, err := p.db.GetHomeTimelineForAccount(authed.Account.ID, maxID, sinceID, minID, limit, local)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	apiStatuses := []apimodel.Status{}
	for _, s := range statuses {
		targetAccount := &gtsmodel.Account{}
		if err := p.db.GetByID(s.AccountID, targetAccount); err != nil {
			if _, ok := err.(db.ErrNoEntries); ok {
				l.Debugf("skipping status %s because account %s can't be found in the db", s.ID, s.AccountID)
				continue
			}
			return nil, NewErrorInternalError(fmt.Errorf("HomeTimelineGet: error getting status author: %s", err))
		}

		relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(s)
		if err != nil {
			l.Debugf("skipping status %s because we couldn't pull relevant accounts from the db", s.ID)
			continue
		}

		visible, err := p.db.StatusVisible(s, targetAccount, authed.Account, relevantAccounts)
		if err != nil {
			return nil, NewErrorInternalError(fmt.Errorf("HomeTimelineGet: error checking status visibility: %s", err))
		}
		if !visible {
			continue
		}

		var boostedStatus *gtsmodel.Status
		if s.BoostOfID != "" {
			bs := &gtsmodel.Status{}
			if err := p.db.GetByID(s.BoostOfID, bs); err != nil {
				if _, ok := err.(db.ErrNoEntries); ok {
					l.Debugf("skipping status %s because status %s can't be found in the db", s.ID, s.BoostOfID)
					continue
				}
				return nil, NewErrorInternalError(fmt.Errorf("HomeTimelineGet: error getting boosted status: %s", err))
			}
			boostedRelevantAccounts, err := p.db.PullRelevantAccountsFromStatus(bs)
			if err != nil {
				l.Debugf("skipping status %s because we couldn't pull relevant accounts from the db", s.ID)
				continue
			}

			boostedVisible, err := p.db.StatusVisible(bs, relevantAccounts.BoostedAccount, authed.Account, boostedRelevantAccounts)
			if err != nil {
				return nil, NewErrorInternalError(fmt.Errorf("HomeTimelineGet: error checking boosted status visibility: %s", err))
			}

			if boostedVisible {
				boostedStatus = bs
			}
		}

		apiStatus, err := p.tc.StatusToMasto(s, targetAccount, authed.Account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, boostedStatus)
		if err != nil {
			l.Debugf("skipping status %s because it couldn't be converted to its mastodon representation: %s", s.ID, err)
			continue
		}

		apiStatuses = append(apiStatuses, *apiStatus)
	}

	return apiStatuses, nil
}

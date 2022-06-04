/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package account

import (
	"context"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) StatusesGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string, limit int, excludeReplies bool, excludeReblogs bool, maxID string, minID string, pinnedOnly bool, mediaOnly bool, publicOnly bool) (*apimodel.TimelineResponse, gtserror.WithCode) {
	if requestingAccount != nil {
		if blocked, err := p.db.IsBlocked(ctx, requestingAccount.ID, targetAccountID, true); err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		} else if blocked {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("block exists between accounts"))
		}
	}

	statuses, err := p.db.GetAccountStatuses(ctx, targetAccountID, limit, excludeReplies, excludeReblogs, maxID, minID, pinnedOnly, mediaOnly, publicOnly)
	if err != nil {
		if err == db.ErrNoEntries {
			return util.EmptyTimelineResponse(), nil
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	var filtered []*gtsmodel.Status
	for _, s := range statuses {
		visible, err := p.filter.StatusVisible(ctx, s, requestingAccount)
		if err == nil && visible {
			filtered = append(filtered, s)
		}
	}

	if len(filtered) == 0 {
		return util.EmptyTimelineResponse(), nil
	}

	timelineables := []timeline.Timelineable{}
	for _, i := range filtered {
		apiStatus, err := p.tc.StatusToAPIStatus(ctx, i, requestingAccount)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status to api: %s", err))
		}

		timelineables = append(timelineables, apiStatus)
	}

	return util.PackageTimelineableResponse(util.TimelineableResponseParams{
		Items:          timelineables,
		Path:           fmt.Sprintf("/api/v1/accounts/%s/statuses", targetAccountID),
		NextMaxIDValue: timelineables[len(timelineables)-1].GetID(),
		PrevMinIDValue: timelineables[0].GetID(),
		Limit:          limit,
		ExtraQueryParams: []string{
			fmt.Sprintf("exclude_replies=%t", excludeReplies),
			fmt.Sprintf("exclude_reblogs=%t", excludeReblogs),
			fmt.Sprintf("pinned_only=%t", pinnedOnly),
			fmt.Sprintf("only_media=%t", mediaOnly),
			fmt.Sprintf("only_public=%t", publicOnly),
		},
	})
}

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

package timeline

import (
	"context"
	"errors"
	"fmt"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	statusfilter "code.superseriousbusiness.org/gotosocial/internal/filter/status"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// FavedTimelineGet ...
func (p *Processor) FavedTimelineGet(ctx context.Context, authed *apiutil.Auth, maxID string, minID string, limit int) (*apimodel.PageableResponse, gtserror.WithCode) {
	statuses, nextMaxID, prevMinID, err := p.state.DB.GetFavedTimeline(ctx, authed.Account.ID, maxID, minID, limit)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = fmt.Errorf("FavedTimelineGet: db error getting statuses: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(statuses)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	items := make([]interface{}, 0, count)
	for _, s := range statuses {
		visible, err := p.visFilter.StatusVisible(ctx, authed.Account, s)
		if err != nil {
			log.Errorf(ctx, "error checking status visibility: %v", err)
			continue
		}

		if !visible {
			continue
		}

		apiStatus, err := p.converter.StatusToAPIStatus(ctx, s, authed.Account, statusfilter.FilterContextNone, nil, nil)
		if err != nil {
			log.Errorf(ctx, "error convering to api status: %v", err)
			continue
		}

		items = append(items, apiStatus)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "/api/v1/favourites",
		NextMaxIDValue: nextMaxID,
		PrevMinIDValue: prevMinID,
		Limit:          limit,
	})
}

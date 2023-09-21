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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *Processor) PublicTimelineGet(ctx context.Context, authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, local bool) (*apimodel.PageableResponse, gtserror.WithCode) {
	statuses, err := p.state.DB.GetPublicTimeline(ctx, maxID, sinceID, minID, limit, local)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = fmt.Errorf("PublicTimelineGet: db error getting statuses: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(statuses)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	var (
		items = make([]interface{}, 0, count)

		// Set next + prev values before filtering and API
		// converting, so caller can still page properly.
		nextMaxIDValue = statuses[count-1].ID
		prevMinIDValue = statuses[0].ID
	)

	for _, s := range statuses {
		timelineable, err := p.filter.StatusPublicTimelineable(ctx, authed.Account, s)
		if err != nil {
			log.Errorf(ctx, "error checking status visibility: %v", err)
			continue
		}

		if !timelineable {
			continue
		}

		apiStatus, err := p.converter.StatusToAPIStatus(ctx, s, authed.Account)
		if err != nil {
			log.Errorf(ctx, "error convert to api status: %v", err)
			continue
		}

		items = append(items, apiStatus)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "/api/v1/timelines/public",
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}

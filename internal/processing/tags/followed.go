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

package tags

import (
	"context"
	"errors"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Followed gets the user's list of followed tags.
func (p *Processor) Followed(
	ctx context.Context,
	accountID string,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	tags, err := p.state.DB.GetFollowedTags(ctx,
		accountID,
		page,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(
			gtserror.Newf("DB error getting followed tags for account %s: %w", accountID, err),
		)
	}

	count := len(tags)
	if len(tags) == 0 {
		return util.EmptyPageableResponse(), nil
	}

	lo := tags[count-1].ID
	hi := tags[0].ID

	items := make([]interface{}, 0, count)
	following := util.Ptr(true)
	for _, tag := range tags {
		apiTag, err := p.converter.TagToAPITag(ctx, tag, true, following)
		if err != nil {
			log.Errorf(ctx, "error converting tag %s to API representation: %v", tag.ID, err)
			continue
		}
		items = append(items, apiTag)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/followed_tags",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}

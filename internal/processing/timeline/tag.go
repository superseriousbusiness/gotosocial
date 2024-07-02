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
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// TagTimelineGet gets a pageable timeline for the given
// tagName and given paging parameters. It will ensure
// that each status in the timeline is actually visible
// to requestingAcct before returning it.
func (p *Processor) TagTimelineGet(
	ctx context.Context,
	requestingAcct *gtsmodel.Account,
	tagName string,
	maxID string,
	sinceID string,
	minID string,
	limit int,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	tag, errWithCode := p.getTag(ctx, tagName)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if tag == nil || !*tag.Useable || !*tag.Listable {
		// Obey mastodon API by returning 404 for this.
		err := fmt.Errorf("tag was not found, or not useable/listable on this instance")
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	statuses, err := p.state.DB.GetTagTimeline(ctx, tag.ID, maxID, sinceID, minID, limit)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting statuses: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.packageTagResponse(
		ctx,
		requestingAcct,
		statuses,
		limit,
		// Use API URL for tag.
		"/api/v1/timelines/tag/"+tagName,
	)
}

func (p *Processor) getTag(ctx context.Context, tagName string) (*gtsmodel.Tag, gtserror.WithCode) {
	// Normalize + validate tag name.
	tagNameNormal, ok := text.NormalizeHashtag(tagName)
	if !ok {
		err := gtserror.Newf("string '%s' could not be normalized to a valid hashtag", tagName)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Ensure we have tag with this name in the db.
	tag, err := p.state.DB.GetTagByName(ctx, tagNameNormal)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Real db error.
		err = gtserror.Newf("db error getting tag by name: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return tag, nil
}

func (p *Processor) packageTagResponse(
	ctx context.Context,
	requestingAcct *gtsmodel.Account,
	statuses []*gtsmodel.Status,
	limit int,
	requestPath string,
) (*apimodel.PageableResponse, gtserror.WithCode) {
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

	filters, err := p.state.DB.GetFiltersForAccountID(ctx, requestingAcct.ID)
	if err != nil {
		err = gtserror.Newf("couldn't retrieve filters for account %s: %w", requestingAcct.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	mutes, err := p.state.DB.GetAccountMutes(gtscontext.SetBarebones(ctx), requestingAcct.ID, nil)
	if err != nil {
		err = gtserror.Newf("couldn't retrieve mutes for account %s: %w", requestingAcct.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	compiledMutes := usermute.NewCompiledUserMuteList(mutes)

	for _, s := range statuses {
		timelineable, err := p.visFilter.StatusTagTimelineable(ctx, requestingAcct, s)
		if err != nil {
			log.Errorf(ctx, "error checking status visibility: %v", err)
			continue
		}

		if !timelineable {
			continue
		}

		apiStatus, err := p.converter.StatusToAPIStatus(ctx, s, requestingAcct, statusfilter.FilterContextPublic, filters, compiledMutes)
		if errors.Is(err, statusfilter.ErrHideStatus) {
			continue
		}
		if err != nil {
			log.Errorf(ctx, "error converting to api status: %v", err)
			continue
		}

		items = append(items, apiStatus)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           requestPath,
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}

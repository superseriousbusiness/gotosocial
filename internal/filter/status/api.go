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

package status

import (
	"time"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// NOTE: the below functions have all been copied
// from typeutils to prevent an import cycle. when
// we move the filtering logic out of the converter
// then we can safely remove these and call necessary
// function without any worry of import cycles.

func toAPIFilterV2(filter *gtsmodel.Filter) apimodel.FilterV2 {
	apiFilterKeywords := make([]apimodel.FilterKeyword, len(filter.Keywords))
	if len(apiFilterKeywords) != len(filter.Keywords) {
		// bound check eliminiation compiler-hint
		panic(gtserror.New("BCE"))
	}
	for i, filterKeyword := range filter.Keywords {
		apiFilterKeywords[i] = apimodel.FilterKeyword{
			ID:        filterKeyword.ID,
			Keyword:   filterKeyword.Keyword,
			WholeWord: util.PtrOrValue(filterKeyword.WholeWord, false),
		}
	}
	apiFilterStatuses := make([]apimodel.FilterStatus, len(filter.Statuses))
	if len(apiFilterStatuses) != len(filter.Statuses) {
		// bound check eliminiation compiler-hint
		panic(gtserror.New("BCE"))
	}
	for i, filterStatus := range filter.Statuses {
		apiFilterStatuses[i] = apimodel.FilterStatus{
			ID:       filterStatus.ID,
			StatusID: filterStatus.StatusID,
		}
	}
	return apimodel.FilterV2{
		ID:           filter.ID,
		Title:        filter.Title,
		Context:      toAPIFilterContexts(filter),
		ExpiresAt:    toAPIFilterExpiresAt(filter.ExpiresAt),
		FilterAction: toAPIFilterAction(filter.Action),
		Keywords:     apiFilterKeywords,
		Statuses:     apiFilterStatuses,
	}
}

func toAPIFilterExpiresAt(expiresAt time.Time) *string {
	if expiresAt.IsZero() {
		return nil
	}
	return util.Ptr(util.FormatISO8601(expiresAt))
}

func toAPIFilterContexts(filter *gtsmodel.Filter) []apimodel.FilterContext {
	apiContexts := make([]apimodel.FilterContext, 0, apimodel.FilterContextNumValues)
	if filter.Contexts.Home() {
		apiContexts = append(apiContexts, apimodel.FilterContextHome)
	}
	if filter.Contexts.Notifications() {
		apiContexts = append(apiContexts, apimodel.FilterContextNotifications)
	}
	if filter.Contexts.Public() {
		apiContexts = append(apiContexts, apimodel.FilterContextPublic)
	}
	if filter.Contexts.Thread() {
		apiContexts = append(apiContexts, apimodel.FilterContextThread)
	}
	if filter.Contexts.Account() {
		apiContexts = append(apiContexts, apimodel.FilterContextAccount)
	}
	return apiContexts
}

func toAPIFilterAction(m gtsmodel.FilterAction) apimodel.FilterAction {
	switch m {
	case gtsmodel.FilterActionWarn:
		return apimodel.FilterActionWarn
	case gtsmodel.FilterActionHide:
		return apimodel.FilterActionHide
	}
	return apimodel.FilterActionNone
}

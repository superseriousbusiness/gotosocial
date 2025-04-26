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

package v2

import (
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"github.com/gin-gonic/gin"
)

const (
	// BasePath is the base path for serving the filters API, minus the 'api' prefix
	BasePath = "/v2/filters"
	// BasePathWithID is the base path with the ID key in it, for operations on an existing filter.
	BasePathWithID = BasePath + "/:" + apiutil.IDKey
	// FilterKeywordsPathWithID is the path for operations on an existing filter's keywords.
	FilterKeywordsPathWithID = BasePathWithID + "/keywords"
	// FilterStatusesPathWithID is the path for operations on an existing filter's statuses.
	FilterStatusesPathWithID = BasePathWithID + "/statuses"

	// KeywordPath is the base path for operations on filter keywords that don't require a filter ID.
	KeywordPath = BasePath + "/keywords"
	// KeywordPathWithKeywordID is the path for operations on an existing filter keyword.
	KeywordPathWithKeywordID = KeywordPath + "/:" + apiutil.IDKey

	// StatusPath is the base path for operations on filter statuses that don't require a filter ID.
	StatusPath = BasePath + "/statuses"
	// StatusPathWithStatusID is the path for operations on an existing filter status.
	StatusPathWithStatusID = StatusPath + "/:" + apiutil.IDKey
)

// Module implements APIs for client-side aka "v1" filtering.
type Module struct {
	processor *processing.Processor
}

func New(processor *processing.Processor) *Module {
	return &Module{
		processor: processor,
	}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodGet, BasePath, m.FiltersGETHandler)

	attachHandler(http.MethodPost, BasePath, m.FilterPOSTHandler)
	attachHandler(http.MethodGet, BasePathWithID, m.FilterGETHandler)
	attachHandler(http.MethodPut, BasePathWithID, m.FilterPUTHandler)
	attachHandler(http.MethodDelete, BasePathWithID, m.FilterDELETEHandler)

	attachHandler(http.MethodGet, FilterKeywordsPathWithID, m.FilterKeywordsGETHandler)
	attachHandler(http.MethodPost, FilterKeywordsPathWithID, m.FilterKeywordPOSTHandler)

	attachHandler(http.MethodGet, KeywordPathWithKeywordID, m.FilterKeywordGETHandler)
	attachHandler(http.MethodPut, KeywordPathWithKeywordID, m.FilterKeywordPUTHandler)
	attachHandler(http.MethodDelete, KeywordPathWithKeywordID, m.FilterKeywordDELETEHandler)

	attachHandler(http.MethodGet, FilterStatusesPathWithID, m.FilterStatusesGETHandler)
	attachHandler(http.MethodPost, FilterStatusesPathWithID, m.FilterStatusPOSTHandler)

	attachHandler(http.MethodGet, StatusPathWithStatusID, m.FilterStatusGETHandler)
	attachHandler(http.MethodDelete, StatusPathWithStatusID, m.FilterStatusDELETEHandler)
}

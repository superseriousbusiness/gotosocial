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

package lists

import (
	"net/http"

	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"github.com/gin-gonic/gin"
)

const (
	IDKey = "id"
	// BasePath is the base path for serving the lists API, minus the 'api' prefix
	BasePath       = "/v1/lists"
	BasePathWithID = BasePath + "/:" + IDKey
	AccountsPath   = BasePathWithID + "/accounts"
	MaxIDKey       = "max_id"
	LimitKey       = "limit"
	SinceIDKey     = "since_id"
	MinIDKey       = "min_id"
)

type Module struct {
	processor *processing.Processor
}

func New(processor *processing.Processor) *Module {
	return &Module{
		processor: processor,
	}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	// create / get / update / delete lists
	attachHandler(http.MethodPost, BasePath, m.ListCreatePOSTHandler)
	attachHandler(http.MethodGet, BasePath, m.ListsGETHandler)
	attachHandler(http.MethodGet, BasePathWithID, m.ListGETHandler)
	attachHandler(http.MethodPut, BasePathWithID, m.ListUpdatePUTHandler)
	attachHandler(http.MethodDelete, BasePathWithID, m.ListDELETEHandler)

	// get / add / remove list accounts
	attachHandler(http.MethodGet, AccountsPath, m.ListAccountsGETHandler)
	attachHandler(http.MethodPost, AccountsPath, m.ListAccountsPOSTHandler)
	attachHandler(http.MethodDelete, AccountsPath, m.ListAccountsDELETEHandler)
}

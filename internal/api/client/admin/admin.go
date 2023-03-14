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

package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
)

const (
	BasePath               = "/v1/admin"
	EmojiPath              = BasePath + "/custom_emojis"
	EmojiPathWithID        = EmojiPath + "/:" + IDKey
	EmojiCategoriesPath    = EmojiPath + "/categories"
	DomainBlocksPath       = BasePath + "/domain_blocks"
	DomainBlocksPathWithID = DomainBlocksPath + "/:" + IDKey
	AccountsPath           = BasePath + "/accounts"
	AccountsPathWithID     = AccountsPath + "/:" + IDKey
	AccountsActionPath     = AccountsPathWithID + "/action"
	MediaCleanupPath       = BasePath + "/media_cleanup"
	MediaRefetchPath       = BasePath + "/media_refetch"
	ReportsPath            = BasePath + "/reports"
	ReportsPathWithID      = ReportsPath + "/:" + IDKey
	ReportsResolvePath     = ReportsPathWithID + "/resolve"
	EmailPath              = BasePath + "/email"
	EmailTestPath          = EmailPath + "/test"

	ExportQueryKey        = "export"
	ImportQueryKey        = "import"
	IDKey                 = "id"
	FilterQueryKey        = "filter"
	MaxShortcodeDomainKey = "max_shortcode_domain"
	MinShortcodeDomainKey = "min_shortcode_domain"
	LimitKey              = "limit"
	DomainQueryKey        = "domain"
	ResolvedKey           = "resolved"
	AccountIDKey          = "account_id"
	TargetAccountIDKey    = "target_account_id"
	MaxIDKey              = "max_id"
	SinceIDKey            = "since_id"
	MinIDKey              = "min_id"
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
	// emoji stuff
	attachHandler(http.MethodPost, EmojiPath, m.EmojiCreatePOSTHandler)
	attachHandler(http.MethodGet, EmojiPath, m.EmojisGETHandler)
	attachHandler(http.MethodDelete, EmojiPathWithID, m.EmojiDELETEHandler)
	attachHandler(http.MethodGet, EmojiPathWithID, m.EmojiGETHandler)
	attachHandler(http.MethodPatch, EmojiPathWithID, m.EmojiPATCHHandler)
	attachHandler(http.MethodGet, EmojiCategoriesPath, m.EmojiCategoriesGETHandler)

	// domain block stuff
	attachHandler(http.MethodPost, DomainBlocksPath, m.DomainBlocksPOSTHandler)
	attachHandler(http.MethodGet, DomainBlocksPath, m.DomainBlocksGETHandler)
	attachHandler(http.MethodGet, DomainBlocksPathWithID, m.DomainBlockGETHandler)
	attachHandler(http.MethodDelete, DomainBlocksPathWithID, m.DomainBlockDELETEHandler)

	// accounts stuff
	attachHandler(http.MethodPost, AccountsActionPath, m.AccountActionPOSTHandler)

	// media stuff
	attachHandler(http.MethodPost, MediaCleanupPath, m.MediaCleanupPOSTHandler)
	attachHandler(http.MethodPost, MediaRefetchPath, m.MediaRefetchPOSTHandler)

	// reports stuff
	attachHandler(http.MethodGet, ReportsPath, m.ReportsGETHandler)
	attachHandler(http.MethodGet, ReportsPathWithID, m.ReportGETHandler)
	attachHandler(http.MethodPost, ReportsResolvePath, m.ReportResolvePOSTHandler)

	// email stuff
	attachHandler(http.MethodPost, EmailTestPath, m.EmailTestPOSTHandler)
}

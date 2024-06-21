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

	"codeberg.org/gruf/go-debug"
	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

const (
	BasePath                = "/v1/admin"
	EmojiPath               = BasePath + "/custom_emojis"
	EmojiPathWithID         = EmojiPath + "/:" + apiutil.IDKey
	EmojiCategoriesPath     = EmojiPath + "/categories"
	DomainBlocksPath        = BasePath + "/domain_blocks"
	DomainBlocksPathWithID  = DomainBlocksPath + "/:" + apiutil.IDKey
	DomainAllowsPath        = BasePath + "/domain_allows"
	DomainAllowsPathWithID  = DomainAllowsPath + "/:" + apiutil.IDKey
	DomainKeysExpirePath    = BasePath + "/domain_keys_expire"
	HeaderAllowsPath        = BasePath + "/header_allows"
	HeaderAllowsPathWithID  = HeaderAllowsPath + "/:" + apiutil.IDKey
	HeaderBlocksPath        = BasePath + "/header_blocks"
	HeaderBlocksPathWithID  = HeaderBlocksPath + "/:" + apiutil.IDKey
	AccountsV1Path          = BasePath + "/accounts"
	AccountsV2Path          = "/v2/admin/accounts"
	AccountsPathWithID      = AccountsV1Path + "/:" + apiutil.IDKey
	AccountsActionPath      = AccountsPathWithID + "/action"
	AccountsApprovePath     = AccountsPathWithID + "/approve"
	AccountsRejectPath      = AccountsPathWithID + "/reject"
	MediaCleanupPath        = BasePath + "/media_cleanup"
	MediaRefetchPath        = BasePath + "/media_refetch"
	ReportsPath             = BasePath + "/reports"
	ReportsPathWithID       = ReportsPath + "/:" + apiutil.IDKey
	ReportsResolvePath      = ReportsPathWithID + "/resolve"
	EmailPath               = BasePath + "/email"
	EmailTestPath           = EmailPath + "/test"
	InstanceRulesPath       = BasePath + "/instance/rules"
	InstanceRulesPathWithID = InstanceRulesPath + "/:" + apiutil.IDKey
	DebugPath               = BasePath + "/debug"
	DebugAPUrlPath          = DebugPath + "/apurl"
	DebugClearCachesPath    = DebugPath + "/caches/clear"

	FilterQueryKey        = "filter"
	MaxShortcodeDomainKey = "max_shortcode_domain"
	MinShortcodeDomainKey = "min_shortcode_domain"
	DomainQueryKey        = "domain"
)

type Module struct {
	processor *processing.Processor
	state     *state.State
}

func New(state *state.State, processor *processing.Processor) *Module {
	return &Module{
		processor: processor,
		state:     state,
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

	// domain allow stuff
	attachHandler(http.MethodPost, DomainAllowsPath, m.DomainAllowsPOSTHandler)
	attachHandler(http.MethodGet, DomainAllowsPath, m.DomainAllowsGETHandler)
	attachHandler(http.MethodGet, DomainAllowsPathWithID, m.DomainAllowGETHandler)
	attachHandler(http.MethodDelete, DomainAllowsPathWithID, m.DomainAllowDELETEHandler)

	// header filtering administration routes
	attachHandler(http.MethodGet, HeaderAllowsPathWithID, m.HeaderFilterAllowGET)
	attachHandler(http.MethodGet, HeaderBlocksPathWithID, m.HeaderFilterBlockGET)
	attachHandler(http.MethodGet, HeaderAllowsPath, m.HeaderFilterAllowsGET)
	attachHandler(http.MethodGet, HeaderBlocksPath, m.HeaderFilterBlocksGET)
	attachHandler(http.MethodPost, HeaderAllowsPath, m.HeaderFilterAllowPOST)
	attachHandler(http.MethodPost, HeaderBlocksPath, m.HeaderFilterBlockPOST)
	attachHandler(http.MethodDelete, HeaderAllowsPathWithID, m.HeaderFilterAllowDELETE)
	attachHandler(http.MethodDelete, HeaderBlocksPathWithID, m.HeaderFilterBlockDELETE)

	// domain maintenance stuff
	attachHandler(http.MethodPost, DomainKeysExpirePath, m.DomainKeysExpirePOSTHandler)

	// accounts stuff
	attachHandler(http.MethodGet, AccountsV1Path, m.AccountsGETV1Handler)
	attachHandler(http.MethodGet, AccountsV2Path, m.AccountsGETV2Handler)
	attachHandler(http.MethodGet, AccountsPathWithID, m.AccountGETHandler)
	attachHandler(http.MethodPost, AccountsActionPath, m.AccountActionPOSTHandler)
	attachHandler(http.MethodPost, AccountsApprovePath, m.AccountApprovePOSTHandler)
	attachHandler(http.MethodPost, AccountsRejectPath, m.AccountRejectPOSTHandler)

	// media stuff
	attachHandler(http.MethodPost, MediaCleanupPath, m.MediaCleanupPOSTHandler)
	attachHandler(http.MethodPost, MediaRefetchPath, m.MediaRefetchPOSTHandler)

	// reports stuff
	attachHandler(http.MethodGet, ReportsPath, m.ReportsGETHandler)
	attachHandler(http.MethodGet, ReportsPathWithID, m.ReportGETHandler)
	attachHandler(http.MethodPost, ReportsResolvePath, m.ReportResolvePOSTHandler)

	// email stuff
	attachHandler(http.MethodPost, EmailTestPath, m.EmailTestPOSTHandler)

	// instance rules stuff
	attachHandler(http.MethodGet, InstanceRulesPath, m.RulesGETHandler)
	attachHandler(http.MethodGet, InstanceRulesPathWithID, m.RuleGETHandler)
	attachHandler(http.MethodPost, InstanceRulesPath, m.RulePOSTHandler)
	attachHandler(http.MethodPatch, InstanceRulesPathWithID, m.RulePATCHHandler)
	attachHandler(http.MethodDelete, InstanceRulesPathWithID, m.RuleDELETEHandler)

	// debug stuff
	if debug.DEBUG {
		attachHandler(http.MethodGet, DebugAPUrlPath, m.DebugAPUrlHandler)
		attachHandler(http.MethodPost, DebugClearCachesPath, m.DebugClearCachesHandler)
	}
}

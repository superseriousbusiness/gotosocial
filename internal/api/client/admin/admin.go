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

package admin

import (
	"net/http"

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	// BasePath is the base API path for this module.
	BasePath = "/api/v1/admin"
	// EmojiPath is used for posting/deleting custom emojis.
	EmojiPath = BasePath + "/custom_emojis"
	// EmojiPathWithID is used for interacting with a single emoji.
	EmojiPathWithID = EmojiPath + "/:" + IDKey
	// EmojiCategoriesPath is used for interacting with emoji categories.
	EmojiCategoriesPath = EmojiPath + "/categories"
	// DomainBlocksPath is used for posting domain blocks.
	DomainBlocksPath = BasePath + "/domain_blocks"
	// DomainBlocksPathWithID is used for interacting with a single domain block.
	DomainBlocksPathWithID = DomainBlocksPath + "/:" + IDKey
	// AccountsPath is used for listing + acting on accounts.
	AccountsPath = BasePath + "/accounts"
	// AccountsPathWithID is used for interacting with a single account.
	AccountsPathWithID = AccountsPath + "/:" + IDKey
	// AccountsActionPath is used for taking action on a single account.
	AccountsActionPath = AccountsPathWithID + "/action"
	MediaCleanupPath   = BasePath + "/media_cleanup"

	// ExportQueryKey is for requesting a public export of some data.
	ExportQueryKey = "export"
	// ImportQueryKey is for submitting an import of some data.
	ImportQueryKey = "import"
	// IDKey specifies the ID of a single item being interacted with.
	IDKey = "id"
	// FilterKey is for applying filters to admin views of accounts, emojis, etc.
	FilterQueryKey = "filter"
	// MaxShortcodeDomainKey is the url query for returning emoji results lower (alphabetically)
	// than the given `[shortcode]@[domain]` parameter.
	MaxShortcodeDomainKey = "max_shortcode_domain"
	// MaxShortcodeDomainKey is the url query for returning emoji results higher (alphabetically)
	// than the given `[shortcode]@[domain]` parameter.
	MinShortcodeDomainKey = "min_shortcode_domain"
	// LimitKey is for specifying maximum number of results to return.
	LimitKey = "limit"
)

// Module implements the ClientAPIModule interface for admin-related actions (reports, emojis, etc)
type Module struct {
	processor processing.Processor
}

// New returns a new admin module
func New(processor processing.Processor) api.ClientModule {
	return &Module{
		processor: processor,
	}
}

// Route attaches all routes from this module to the given router
func (m *Module) Route(r router.Router) error {
	r.AttachHandler(http.MethodPost, EmojiPath, m.EmojiCreatePOSTHandler)
	r.AttachHandler(http.MethodGet, EmojiPath, m.EmojisGETHandler)
	r.AttachHandler(http.MethodDelete, EmojiPathWithID, m.EmojiDELETEHandler)
	r.AttachHandler(http.MethodGet, EmojiPathWithID, m.EmojiGETHandler)
	r.AttachHandler(http.MethodPost, DomainBlocksPath, m.DomainBlocksPOSTHandler)
	r.AttachHandler(http.MethodGet, DomainBlocksPath, m.DomainBlocksGETHandler)
	r.AttachHandler(http.MethodGet, DomainBlocksPathWithID, m.DomainBlockGETHandler)
	r.AttachHandler(http.MethodDelete, DomainBlocksPathWithID, m.DomainBlockDELETEHandler)
	r.AttachHandler(http.MethodPost, AccountsActionPath, m.AccountActionPOSTHandler)
	r.AttachHandler(http.MethodPost, MediaCleanupPath, m.MediaCleanupPOSTHandler)
	r.AttachHandler(http.MethodGet, EmojiCategoriesPath, m.EmojiCategoriesGETHandler)
	return nil
}

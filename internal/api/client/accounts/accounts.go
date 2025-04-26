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

package accounts

import (
	"net/http"

	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"github.com/gin-gonic/gin"
)

const (
	ExcludeReblogsKey = "exclude_reblogs"
	ExcludeRepliesKey = "exclude_replies"
	LimitKey          = "limit"
	MaxIDKey          = "max_id"
	MinIDKey          = "min_id"
	OnlyMediaKey      = "only_media"
	OnlyPublicKey     = "only_public"
	PinnedKey         = "pinned"

	BasePath       = "/v1/accounts"
	IDKey          = "id"
	BasePathWithID = BasePath + "/:" + IDKey

	BlockPath         = BasePathWithID + "/block"
	DeletePath        = BasePath + "/delete"
	FeaturedTagsPath  = BasePathWithID + "/featured_tags"
	FollowersPath     = BasePathWithID + "/followers"
	FollowingPath     = BasePathWithID + "/following"
	FollowPath        = BasePathWithID + "/follow"
	ListsPath         = BasePathWithID + "/lists"
	LookupPath        = BasePath + "/lookup"
	MutePath          = BasePathWithID + "/mute"
	NotePath          = BasePathWithID + "/note"
	RelationshipsPath = BasePath + "/relationships"
	SearchPath        = BasePath + "/search"
	StatusesPath      = BasePathWithID + "/statuses"
	UnblockPath       = BasePathWithID + "/unblock"
	UnfollowPath      = BasePathWithID + "/unfollow"
	UnmutePath        = BasePathWithID + "/unmute"
	UpdatePath        = BasePath + "/update_credentials"
	VerifyPath        = BasePath + "/verify_credentials"
	MovePath          = BasePath + "/move"
	AliasPath         = BasePath + "/alias"
	ThemesPath        = BasePath + "/themes"

	// ProfileBasePath for the profile API, an extension of the account update API with a different path.
	ProfileBasePath = "/v1/profile"
	AvatarPath      = ProfileBasePath + "/avatar"
	HeaderPath      = ProfileBasePath + "/header"
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
	// create account
	attachHandler(http.MethodPost, BasePath, m.AccountCreatePOSTHandler)

	// get account
	attachHandler(http.MethodGet, BasePathWithID, m.AccountGETHandler)

	// delete account
	attachHandler(http.MethodPost, DeletePath, m.AccountDeletePOSTHandler)

	// verify account
	attachHandler(http.MethodGet, VerifyPath, m.AccountVerifyGETHandler)

	// modify account
	attachHandler(http.MethodPatch, UpdatePath, m.AccountUpdateCredentialsPATCHHandler)

	// modify account profile media
	attachHandler(http.MethodDelete, AvatarPath, m.AccountAvatarDELETEHandler)
	attachHandler(http.MethodDelete, HeaderPath, m.AccountHeaderDELETEHandler)

	// get account's statuses
	attachHandler(http.MethodGet, StatusesPath, m.AccountStatusesGETHandler)

	// get account's featured tags
	attachHandler(http.MethodGet, FeaturedTagsPath, m.AccountFeaturedTagsGETHandler)

	// get following or followers
	attachHandler(http.MethodGet, FollowersPath, m.AccountFollowersGETHandler)
	attachHandler(http.MethodGet, FollowingPath, m.AccountFollowingGETHandler)

	// get relationship with account
	attachHandler(http.MethodGet, RelationshipsPath, m.AccountRelationshipsGETHandler)

	// follow or unfollow account
	attachHandler(http.MethodPost, FollowPath, m.AccountFollowPOSTHandler)
	attachHandler(http.MethodPost, UnfollowPath, m.AccountUnfollowPOSTHandler)

	// block or unblock account
	attachHandler(http.MethodPost, BlockPath, m.AccountBlockPOSTHandler)
	attachHandler(http.MethodPost, UnblockPath, m.AccountUnblockPOSTHandler)

	// account lists
	attachHandler(http.MethodGet, ListsPath, m.AccountListsGETHandler)

	// account note
	attachHandler(http.MethodPost, NotePath, m.AccountNotePOSTHandler)

	// mute or unmute account
	attachHandler(http.MethodPost, MutePath, m.AccountMutePOSTHandler)
	attachHandler(http.MethodPost, UnmutePath, m.AccountUnmutePOSTHandler)

	// search for accounts
	attachHandler(http.MethodGet, SearchPath, m.AccountSearchGETHandler)
	attachHandler(http.MethodGet, LookupPath, m.AccountLookupGETHandler)

	// migration handlers
	attachHandler(http.MethodPost, AliasPath, m.AccountAliasPOSTHandler)
	attachHandler(http.MethodPost, MovePath, m.AccountMovePOSTHandler)

	// account themes
	attachHandler(http.MethodGet, ThemesPath, m.AccountThemesGETHandler)
}

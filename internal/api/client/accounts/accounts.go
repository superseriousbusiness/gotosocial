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

package accounts

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
)

const (
	// LimitKey is for setting the return amount limit for eg., requesting an account's statuses
	LimitKey = "limit"
	// ExcludeRepliesKey is for specifying whether to exclude replies in a list of returned statuses by an account.
	ExcludeRepliesKey = "exclude_replies"
	// ExcludeReblogsKey is for specifying whether to exclude reblogs in a list of returned statuses by an account.
	ExcludeReblogsKey = "exclude_reblogs"
	// PinnedKey is for specifying whether to include pinned statuses in a list of returned statuses by an account.
	PinnedKey = "pinned"
	// MaxIDKey is for specifying the maximum ID of the status to retrieve.
	MaxIDKey = "max_id"
	// MinIDKey is for specifying the minimum ID of the status to retrieve.
	MinIDKey = "min_id"
	// OnlyMediaKey is for specifying that only statuses with media should be returned in a list of returned statuses by an account.
	OnlyMediaKey = "only_media"
	// OnlyPublicKey is for specifying that only statuses with visibility public should be returned in a list of returned statuses by account.
	OnlyPublicKey = "only_public"

	// IDKey is the key to use for retrieving account ID in requests
	IDKey = "id"
	// BasePath is the base API path for this module, excluding the 'api' prefix
	BasePath = "/v1/accounts"
	// BasePathWithID is the base path for this module with the ID key
	BasePathWithID = BasePath + "/:" + IDKey
	// VerifyPath is for verifying account credentials
	VerifyPath = BasePath + "/verify_credentials"
	// UpdateCredentialsPath is for updating account credentials
	UpdateCredentialsPath = BasePath + "/update_credentials"
	// GetStatusesPath is for showing an account's statuses
	GetStatusesPath = BasePathWithID + "/statuses"
	// GetFollowersPath is for showing an account's followers
	GetFollowersPath = BasePathWithID + "/followers"
	// GetFollowingPath is for showing account's that an account follows.
	GetFollowingPath = BasePathWithID + "/following"
	// GetRelationshipsPath is for showing an account's relationship with other accounts
	GetRelationshipsPath = BasePath + "/relationships"
	// FollowPath is for POSTing new follows to, and updating existing follows
	FollowPath = BasePathWithID + "/follow"
	// UnfollowPath is for POSTing an unfollow
	UnfollowPath = BasePathWithID + "/unfollow"
	// BlockPath is for creating a block of an account
	BlockPath = BasePathWithID + "/block"
	// UnblockPath is for removing a block of an account
	UnblockPath = BasePathWithID + "/unblock"
	// DeleteAccountPath is for deleting one's account via the API
	DeleteAccountPath = BasePath + "/delete"
)

type Module struct {
	processor processing.Processor
}

func New(processor processing.Processor) *Module {
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
	attachHandler(http.MethodPost, DeleteAccountPath, m.AccountDeletePOSTHandler)

	// verify account
	attachHandler(http.MethodGet, VerifyPath, m.AccountVerifyGETHandler)

	// modify account
	attachHandler(http.MethodPatch, UpdateCredentialsPath, m.AccountUpdateCredentialsPATCHHandler)

	// get account's statuses
	attachHandler(http.MethodGet, GetStatusesPath, m.AccountStatusesGETHandler)

	// get following or followers
	attachHandler(http.MethodGet, GetFollowersPath, m.AccountFollowersGETHandler)
	attachHandler(http.MethodGet, GetFollowingPath, m.AccountFollowingGETHandler)

	// get relationship with account
	attachHandler(http.MethodGet, GetRelationshipsPath, m.AccountRelationshipsGETHandler)

	// follow or unfollow account
	attachHandler(http.MethodPost, FollowPath, m.AccountFollowPOSTHandler)
	attachHandler(http.MethodPost, UnfollowPath, m.AccountUnfollowPOSTHandler)

	// block or unblock account
	attachHandler(http.MethodPost, BlockPath, m.AccountBlockPOSTHandler)
	attachHandler(http.MethodPost, UnblockPath, m.AccountUnblockPOSTHandler)
}

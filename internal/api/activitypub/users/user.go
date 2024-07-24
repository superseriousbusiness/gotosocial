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

package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

const (
	// UsernameKey is for account usernames.
	UsernameKey = "username"
	// StatusIDKey is for status IDs
	StatusIDKey = "status"
	// OnlyOtherAccountsKey is for filtering status responses.
	OnlyOtherAccountsKey = "only_other_accounts"
	// MinIDKey is for filtering status responses.
	MinIDKey = "min_id"
	// MaxIDKey is for filtering status responses.
	MaxIDKey = "max_id"
	// PageKey is for filtering status responses.
	PageKey = "page"

	// BasePath is the base path for serving AP 'users' requests, minus the 'users' prefix.
	BasePath = "/:" + UsernameKey
	// InboxPath is for serving POST requests to a user's inbox with the given username key.
	InboxPath = BasePath + "/" + uris.InboxPath
	// OutboxPath is for serving GET requests to a user's outbox with the given username key.
	OutboxPath = BasePath + "/" + uris.OutboxPath
	// FollowersPath is for serving GET request's to a user's followers list, with the given username key.
	FollowersPath = BasePath + "/" + uris.FollowersPath
	// FollowingPath is for serving GET request's to a user's following list, with the given username key.
	FollowingPath = BasePath + "/" + uris.FollowingPath
	// FeaturedCollectionPath is for serving GET requests to a user's list of featured (pinned) statuses.
	FeaturedCollectionPath = BasePath + "/" + uris.CollectionsPath + "/" + uris.FeaturedPath
	// StatusPath is for serving GET requests to a particular status by a user, with the given username key and status ID
	StatusPath = BasePath + "/" + uris.StatusesPath + "/:" + StatusIDKey
	// StatusRepliesPath is for serving the replies collection of a status.
	StatusRepliesPath = StatusPath + "/replies"
	// AcceptPath is for serving accepts of a status.
	AcceptPath = BasePath + "/" + uris.AcceptsPath + "/:" + apiutil.IDKey
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
	attachHandler(http.MethodGet, BasePath, m.UsersGETHandler)
	attachHandler(http.MethodPost, InboxPath, m.InboxPOSTHandler)
	attachHandler(http.MethodGet, FollowersPath, m.FollowersGETHandler)
	attachHandler(http.MethodGet, FollowingPath, m.FollowingGETHandler)
	attachHandler(http.MethodGet, FeaturedCollectionPath, m.FeaturedCollectionGETHandler)
	attachHandler(http.MethodGet, StatusPath, m.StatusGETHandler)
	attachHandler(http.MethodGet, StatusRepliesPath, m.StatusRepliesGETHandler)
	attachHandler(http.MethodGet, OutboxPath, m.OutboxGETHandler)
	attachHandler(http.MethodGet, AcceptPath, m.AcceptGETHandler)
}

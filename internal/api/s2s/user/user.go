/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package user

import (
	"net/http"

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/util"
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

	// UsersBasePath is the base path for serving information about Users eg https://example.org/users
	UsersBasePath = "/" + util.UsersPath
	// UsersBasePathWithUsername is just the users base path with the Username key in it.
	// Use this anywhere you need to know the username of the user being queried.
	// Eg https://example.org/users/:username
	UsersBasePathWithUsername = UsersBasePath + "/:" + UsernameKey
	// UsersPublicKeyPath is a path to a user's public key, for serving bare minimum AP representations.
	UsersPublicKeyPath = UsersBasePathWithUsername + "/" + util.PublicKeyPath
	// UsersInboxPath is for serving POST requests to a user's inbox with the given username key.
	UsersInboxPath = UsersBasePathWithUsername + "/" + util.InboxPath
	// UsersOutboxPath is for serving GET requests to a user's outbox with the given username key.
	UsersOutboxPath = UsersBasePathWithUsername + "/" + util.OutboxPath
	// UsersFollowersPath is for serving GET request's to a user's followers list, with the given username key.
	UsersFollowersPath = UsersBasePathWithUsername + "/" + util.FollowersPath
	// UsersFollowingPath is for serving GET request's to a user's following list, with the given username key.
	UsersFollowingPath = UsersBasePathWithUsername + "/" + util.FollowingPath
	// UsersStatusPath is for serving GET requests to a particular status by a user, with the given username key and status ID
	UsersStatusPath = UsersBasePathWithUsername + "/" + util.StatusesPath + "/:" + StatusIDKey
	// UsersStatusRepliesPath is for serving the replies collection of a status.
	UsersStatusRepliesPath = UsersStatusPath + "/replies"
)

// Module implements the FederationAPIModule interface
type Module struct {
	config    *config.Config
	processor processing.Processor
}

// New returns a new auth module
func New(config *config.Config, processor processing.Processor) api.FederationModule {
	return &Module{
		config:    config,
		processor: processor,
	}
}

// Route satisfies the RESTAPIModule interface
func (m *Module) Route(s router.Router) error {
	s.AttachHandler(http.MethodGet, UsersBasePathWithUsername, m.UsersGETHandler)
	s.AttachHandler(http.MethodPost, UsersInboxPath, m.InboxPOSTHandler)
	s.AttachHandler(http.MethodGet, UsersFollowersPath, m.FollowersGETHandler)
	s.AttachHandler(http.MethodGet, UsersFollowingPath, m.FollowingGETHandler)
	s.AttachHandler(http.MethodGet, UsersStatusPath, m.StatusGETHandler)
	s.AttachHandler(http.MethodGet, UsersPublicKeyPath, m.PublicKeyGETHandler)
	s.AttachHandler(http.MethodGet, UsersStatusRepliesPath, m.StatusRepliesGETHandler)
	s.AttachHandler(http.MethodGet, UsersOutboxPath, m.OutboxGETHandler)
	return nil
}

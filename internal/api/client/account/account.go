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

package account

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/message"

	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	// LimitKey is for setting the return amount limit for eg., requesting an account's statuses
	LimitKey = "limit"
	// ExcludeRepliesKey is for specifying whether to exclude replies in a list of returned statuses by an account.
	ExcludeRepliesKey = "exclude_replies"
	// PinnedKey is for specifying whether to include pinned statuses in a list of returned statuses by an account.
	PinnedKey = "pinned"
	// MaxIDKey is for specifying the maximum ID of the status to retrieve.
	MaxIDKey = "max_id"
	// MediaOnlyKey is for specifying that only statuses with media should be returned in a list of returned statuses by an account.
	MediaOnlyKey = "only_media"

	// IDKey is the key to use for retrieving account ID in requests
	IDKey = "id"
	// BasePath is the base API path for this module
	BasePath = "/api/v1/accounts"
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
	// GetRelationshipsPath is for showing an account's relationship with other accounts
	GetRelationshipsPath = BasePath + "/relationships"
)

// Module implements the ClientAPIModule interface for account-related actions
type Module struct {
	config    *config.Config
	processor message.Processor
	log       *logrus.Logger
}

// New returns a new account module
func New(config *config.Config, processor message.Processor, log *logrus.Logger) api.ClientModule {
	return &Module{
		config:    config,
		processor: processor,
		log:       log,
	}
}

// Route attaches all routes from this module to the given router
func (m *Module) Route(r router.Router) error {
	r.AttachHandler(http.MethodPost, BasePath, m.AccountCreatePOSTHandler)
	r.AttachHandler(http.MethodGet, BasePathWithID, m.muxHandler)
	r.AttachHandler(http.MethodPatch, BasePathWithID, m.muxHandler)
	r.AttachHandler(http.MethodGet, GetStatusesPath, m.AccountStatusesGETHandler)
	r.AttachHandler(http.MethodGet, GetFollowersPath, m.AccountFollowersGETHandler)
	r.AttachHandler(http.MethodGet, GetRelationshipsPath, m.AccountRelationshipsGETHandler)
	return nil
}

func (m *Module) muxHandler(c *gin.Context) {
	ru := c.Request.RequestURI
	switch c.Request.Method {
	case http.MethodGet:
		if strings.HasPrefix(ru, VerifyPath) {
			m.AccountVerifyGETHandler(c)
		} else {
			m.AccountGETHandler(c)
		}
	case http.MethodPatch:
		if strings.HasPrefix(ru, UpdateCredentialsPath) {
			m.AccountUpdateCredentialsPATCHHandler(c)
		}
	}
}

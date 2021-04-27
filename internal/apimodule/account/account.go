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
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"

	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
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
)

// Module implements the ClientAPIModule interface for account-related actions
type Module struct {
	config       *config.Config
	db           db.DB
	oauthServer  oauth.Server
	mediaHandler media.Handler
	tc           typeutils.TypeConverter
	log          *logrus.Logger
}

// New returns a new account module
func New(config *config.Config, db db.DB, oauthServer oauth.Server, mediaHandler media.Handler, tc typeutils.TypeConverter, log *logrus.Logger) apimodule.ClientAPIModule {
	return &Module{
		config:       config,
		db:           db,
		oauthServer:  oauthServer,
		mediaHandler: mediaHandler,
		tc:           tc,
		log:          log,
	}
}

// Route attaches all routes from this module to the given router
func (m *Module) Route(r router.Router) error {
	r.AttachHandler(http.MethodPost, BasePath, m.AccountCreatePOSTHandler)
	r.AttachHandler(http.MethodGet, BasePathWithID, m.muxHandler)
	r.AttachHandler(http.MethodPatch, BasePathWithID, m.muxHandler)
	return nil
}

// CreateTables creates the required tables for this module in the given database
func (m *Module) CreateTables(db db.DB) error {
	models := []interface{}{
		&gtsmodel.User{},
		&gtsmodel.Account{},
		&gtsmodel.Follow{},
		&gtsmodel.FollowRequest{},
		&gtsmodel.Status{},
		&gtsmodel.Application{},
		&gtsmodel.EmailDomainBlock{},
		&gtsmodel.MediaAttachment{},
	}

	for _, m := range models {
		if err := db.CreateTable(m); err != nil {
			return fmt.Errorf("error creating table: %s", err)
		}
	}
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

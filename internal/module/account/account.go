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

	"github.com/gin-gonic/gin"
	"github.com/gotosocial/gotosocial/internal/config"
	"github.com/gotosocial/gotosocial/internal/db"
	"github.com/gotosocial/gotosocial/internal/db/model"
	"github.com/gotosocial/gotosocial/internal/module"
	"github.com/gotosocial/gotosocial/internal/module/oauth"
	"github.com/gotosocial/gotosocial/internal/router"
)

const (
	basePath       = "/api/v1/accounts"
	basePathWithID = basePath + "/:id"
	verifyPath     = basePath + "/verify_credentials"
)

type accountModule struct {
	config *config.Config
	db     db.DB
}

// New returns a new account module
func New(config *config.Config, db db.DB) module.ClientAPIModule {
	return &accountModule{
		config: config,
		db:     db,
	}
}

// Route attaches all routes from this module to the given router
func (m *accountModule) Route(r router.Router) error {
	r.AttachHandler(http.MethodGet, verifyPath, m.AccountVerifyGETHandler)
	return nil
}

// AccountVerifyGETHandler serves a user's account details to them IF they reached this
// handler while in possession of a valid token, according to the oauth middleware.
func (m *accountModule) AccountVerifyGETHandler(c *gin.Context) {
	i, ok := c.Get(oauth.SessionAuthorizedUser)
	fmt.Println(i)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "The access token is invalid"})
		return
	}

	userID, ok := (i).(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "The access token is invalid"})
		return
	}

	acct := &model.Account{}
	if err := m.db.GetAccountByUserID(userID, acct); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	acctSensitive, err := m.db.AccountToMastoSensitive(acct)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, acctSensitive)
}

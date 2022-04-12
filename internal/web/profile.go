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

package web

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (m *Module) profileTemplateHandler(c *gin.Context) {
	l := logrus.WithField("func", "profileTemplateHandler")
	l.Trace("rendering profile template")
	ctx := c.Request.Context()

	username := c.Param(usernameKey)
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no account username specified"})
		return
	}

	authed, err := oauth.Authed(c, false, false, false, false)
	if err != nil {
		l.Errorf("error authing profile GET request: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	instance, errWithCode := m.processor.InstanceGet(ctx, viper.GetString(config.Keys.Host))
	if errWithCode != nil {
		l.Debugf("error getting instance from processor: %s", errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	account, errWithCode := m.processor.AccountGetLocalByUsername(ctx, authed, username)
	if errWithCode != nil {
		l.Debugf("error getting account from processor: %s", errWithCode.Error())
		if errWithCode.Code() == http.StatusNotFound {
			m.NotFoundHandler(c)
			return
		}
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	// if we're getting an AP request on this endpoint we should render the account's AP representation instead
	accept := c.NegotiateFormat(string(api.TextHTML), string(api.AppActivityJSON), string(api.AppActivityLDJSON))
	if accept == string(api.AppActivityJSON) || accept == string(api.AppActivityLDJSON) {
		m.returnAPRepresentation(ctx, c, username, accept)
		return
	}

	// get latest 10 top-level public statuses;
	// ie., exclude replies and boosts, public only,
	// with or without media
	statuses, errWithCode := m.processor.AccountStatusesGet(ctx, authed, account.ID, 10, true, true, "", "", false, false, true)
	if errWithCode != nil {
		l.Debugf("error getting statuses from processor: %s", errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	// pick a random dummy avatar if this account avatar isn't set yet
	if account.Avatar == "" && len(m.defaultAvatars) > 0 {
		//nolint:gosec
		randomIndex := rand.Intn(len(m.defaultAvatars))
		dummyAvatar := m.defaultAvatars[randomIndex]
		account.Avatar = dummyAvatar
		for _, s := range statuses {
			s.Account.Avatar = dummyAvatar
		}
	}

	c.HTML(http.StatusOK, "profile.tmpl", gin.H{
		"instance": instance,
		"account":  account,
		"statuses": statuses,
		"stylesheets": []string{
			"/assets/Fork-Awesome/css/fork-awesome.min.css",
			"/assets/status.css",
			"/assets/profile.css",
		},
	})
}

func (m *Module) returnAPRepresentation(ctx context.Context, c *gin.Context, username string, accept string) {
	verifier, signed := c.Get(string(ap.ContextRequestingPublicKeyVerifier))
	if signed {
		ctx = context.WithValue(ctx, ap.ContextRequestingPublicKeyVerifier, verifier)
	}

	signature, signed := c.Get(string(ap.ContextRequestingPublicKeySignature))
	if signed {
		ctx = context.WithValue(ctx, ap.ContextRequestingPublicKeySignature, signature)
	}

	user, errWithCode := m.processor.GetFediUser(ctx, username, c.Request.URL) // GetFediUser handles auth as well
	if errWithCode != nil {
		logrus.Infof(errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	b, mErr := json.Marshal(user)
	if mErr != nil {
		err := fmt.Errorf("could not marshal json: %s", mErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, accept, b)
}

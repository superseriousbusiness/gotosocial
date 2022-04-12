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
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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

	// if we're getting an AP request on this endpoint we should redirect to the account's AP uri
	accept := c.NegotiateFormat(string(api.TextHTML), string(api.AppActivityJSON), string(api.AppActivityLDJSON))
	if accept == string(api.AppActivityJSON) || accept == string(api.AppActivityLDJSON) {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/users/%s", username))
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

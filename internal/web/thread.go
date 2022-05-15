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
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (m *Module) threadTemplateHandler(c *gin.Context) {
	l := logrus.WithField("func", "threadTemplateGET")
	l.Trace("rendering thread template")

	ctx := c.Request.Context()

	// usernames on our instance will always be lowercase
	username := strings.ToLower(c.Param(usernameKey))
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no account username specified"})
		return
	}

	// status ids will always be uppercase
	statusID := strings.ToUpper(c.Param(statusIDKey))
	if statusID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no status id specified"})
		return
	}

	authed, err := oauth.Authed(c, false, false, false, false)
	if err != nil {
		l.Errorf("error authing status GET request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "status not found"})
		return
	}

	host := viper.GetString(config.Keys.Host)
	instance, err := m.processor.InstanceGet(ctx, host)
	if err != nil {
		l.Debugf("error getting instance from processor: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	status, err := m.processor.StatusGet(ctx, authed, statusID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status not found"})
		return
	}

	if !strings.EqualFold(username, status.Account.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status not found"})
		return
	}

	context, err := m.processor.StatusGetContext(ctx, authed, statusID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status not found"})
		return
	}

	c.HTML(http.StatusOK, "thread.tmpl", gin.H{
		"instance":    instance,
		"status":      status,
		"context":     context,
		"stylesheets": []string{"/assets/Fork-Awesome/css/fork-awesome.min.css", "/assets/status.css"},
	})
}

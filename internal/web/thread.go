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

package web

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

type statusLink struct {
	User string `uri:"user" binding:"required"`
	ID   string `uri:"id"   binding:"required"`
}

func (m *Module) threadTemplateHandler(c *gin.Context) {
	l := logrus.WithField("func", "threadTemplateGET")
	l.Trace("rendering thread template")

	ctx := c.Request.Context()

	var uriParts statusLink

	if err := c.ShouldBindUri(&uriParts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status not found"})
		return
	}

	authed, err := oauth.Authed(c, false, false, false, false)
	if err != nil {
		l.Errorf("error authing status GET request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "status not found"})
		return
	}

	instance, err := m.processor.InstanceGet(ctx, m.config.Host)
	if err != nil {
		l.Debugf("error getting instance from processor: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	status, err := m.processor.StatusGet(ctx, authed, uriParts.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status not found"})
		return
	}

	if uriParts.User[:1] != "@" || uriParts.User[1:] != status.Account.Username {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status not found"})
		return
	}

	context, err := m.processor.StatusGetContext(ctx, authed, uriParts.ID)
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

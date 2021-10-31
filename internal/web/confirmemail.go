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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (m *Module) confirmEmailGETHandler(c *gin.Context) {
	// if there's no token in the query, just serve the 404 web handler
	token := c.Query(tokenParam)
	if token == "" {
		m.NotFoundHandler(c)
		return
	}

	ctx := c.Request.Context()

	user, errWithCode := m.processor.UserConfirmEmail(ctx, token)
	if errWithCode != nil {
		logrus.Debugf("error confirming email: %s", errWithCode.Error())
		// if something goes wrong, just log it and direct to the 404 handler to not give anything away
		m.NotFoundHandler(c)
		return
	}

	instance, err := m.processor.InstanceGet(ctx, m.config.Host)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "confirmed.tmpl", gin.H{
		"instance": instance,
		"email":    user.Email,
		"username": user.Account.Username,
	})
}

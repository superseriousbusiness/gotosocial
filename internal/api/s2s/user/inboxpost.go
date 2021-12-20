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

package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror" //nolint:typecheck
)

// InboxPOSTHandler deals with incoming POST requests to an actor's inbox.
// Eg., POST to https://example.org/users/whatever/inbox.
func (m *Module) InboxPOSTHandler(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func": "InboxPOSTHandler",
		"url":  c.Request.RequestURI,
	})

	requestedUsername := c.Param(UsernameKey)
	if requestedUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no username specified in request"})
		return
	}

	ctx := transferContext(c)

	posted, err := m.processor.InboxPost(ctx, c.Writer, c.Request)
	if err != nil {
		if withCode, ok := err.(gtserror.WithCode); ok {
			l.Debugf("InboxPOSTHandler: %s", withCode.Error())
			c.JSON(withCode.Code(), withCode.Safe())
			return
		}
		l.Debugf("InboxPOSTHandler: error processing request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to process request"})
		return
	}

	if !posted {
		l.Debugf("InboxPOSTHandler: request could not be handled as an AP request; headers were: %+v", c.Request.Header)
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to process request"})
	}
}

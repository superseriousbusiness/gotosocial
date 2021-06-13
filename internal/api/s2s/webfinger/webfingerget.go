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

package webfinger

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// WebfingerGETRequest handles requests to, for example, https://example.org/.well-known/webfinger?resource=acct:some_user@example.org
func (m *Module) WebfingerGETRequest(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
		"func":       "WebfingerGETRequest",
		"user-agent": c.Request.UserAgent(),
	})

	q, set := c.GetQuery("resource")
	if !set || q == "" {
		l.Debug("aborting request because no resource was set in query")
		c.JSON(http.StatusBadRequest, gin.H{"error": "no 'resource' in request query"})
		return
	}

	withAcct := strings.Split(q, "acct:")
	if len(withAcct) != 2 {
		l.Debugf("aborting request because resource query %s could not be split by 'acct:'", q)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	usernameDomain := strings.Split(withAcct[1], "@")
	if len(usernameDomain) != 2 {
		l.Debugf("aborting request because username and domain could not be parsed from %s", withAcct[1])
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	username := strings.ToLower(usernameDomain[0])
	domain := strings.ToLower(usernameDomain[1])
	if username == "" || domain == "" {
		l.Debug("aborting request because username or domain was empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	if domain != m.config.Host {
		l.Debugf("aborting request because domain %s does not belong to this instance", domain)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("domain %s does not belong to this instance", domain)})
		return
	}

	resp, err := m.processor.GetWebfingerAccount(username, c.Request)
	if err != nil {
		l.Debugf("aborting request with an error: %s", err.Error())
		c.JSON(err.Code(), gin.H{"error": err.Safe()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

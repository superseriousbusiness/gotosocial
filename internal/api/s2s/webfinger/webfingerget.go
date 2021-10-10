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
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// WebfingerGETRequest handles requests to, for example, https://example.org/.well-known/webfinger?resource=acct:some_user@example.org
func (m *Module) WebfingerGETRequest(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func":       "WebfingerGETRequest",
		"user-agent": c.Request.UserAgent(),
	})

	q, set := c.GetQuery("resource")
	if !set || q == "" {
		l.Debug("aborting request because no resource was set in query")
		c.JSON(http.StatusBadRequest, gin.H{"error": "no 'resource' in request query"})
		return
	}

	// remove the acct: prefix if it's present
	trimAcct := strings.TrimPrefix(q, "acct:")
	// remove the first @ in @whatever@example.org if it's present
	namestring := strings.TrimPrefix(trimAcct, "@")

	// at this point we should have a string like some_user@example.org
	l.Debugf("got finger request for '%s'", namestring)

	usernameAndAccountDomain := strings.Split(namestring, "@")
	if len(usernameAndAccountDomain) != 2 {
		l.Debugf("aborting request because username and domain could not be parsed from %s", namestring)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	username := strings.ToLower(usernameAndAccountDomain[0])
	accountDomain := strings.ToLower(usernameAndAccountDomain[1])
	if username == "" || accountDomain == "" {
		l.Debug("aborting request because username or domain was empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	if accountDomain != m.config.AccountDomain && accountDomain != m.config.Host {
		l.Debugf("aborting request because accountDomain %s does not belong to this instance", accountDomain)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("accountDomain %s does not belong to this instance", accountDomain)})
		return
	}

	// transfer the signature verifier from the gin context to the request context
	ctx := c.Request.Context()
	verifier, signed := c.Get(string(util.APRequestingPublicKeyVerifier))
	if signed {
		ctx = context.WithValue(ctx, util.APRequestingPublicKeyVerifier, verifier)
	}

	resp, err := m.processor.GetWebfingerAccount(ctx, username)
	if err != nil {
		l.Debugf("aborting request with an error: %s", err.Error())
		c.JSON(err.Code(), gin.H{"error": err.Safe()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

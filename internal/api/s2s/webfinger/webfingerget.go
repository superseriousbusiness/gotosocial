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

package webfinger

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// WebfingerGETRequest swagger:operation GET /.well-known/webfinger webfingerGet
//
// Handles webfinger account lookup requests.
//
// For example, a GET to `https://goblin.technology/.well-known/webfinger?resource=acct:tobi@goblin.technology` would return:
//
// ```
//  {"subject":"acct:tobi@goblin.technology","aliases":["https://goblin.technology/users/tobi","https://goblin.technology/@tobi"],"links":[{"rel":"http://webfinger.net/rel/profile-page","type":"text/html","href":"https://goblin.technology/@tobi"},{"rel":"self","type":"application/activity+json","href":"https://goblin.technology/users/tobi"}]}
// ```
//
// See: https://webfinger.net/
//
// ---
// tags:
// - webfinger
//
// produces:
// - application/json
//
// responses:
//   '200':
//     schema:
//       "$ref": "#/definitions/wellKnownResponse"
func (m *Module) WebfingerGETRequest(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func":       "WebfingerGETRequest",
		"user-agent": c.Request.UserAgent(),
	})

	resourceQuery, set := c.GetQuery("resource")
	if !set || resourceQuery == "" {
		l.Debug("aborting request because no resource was set in query")
		c.JSON(http.StatusBadRequest, gin.H{"error": "no 'resource' in request query"})
		return
	}

	// remove the acct: prefix if it's present
	trimAcct := strings.TrimPrefix(resourceQuery, "acct:")
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
	requestedAccountDomain := strings.ToLower(usernameAndAccountDomain[1])
	if username == "" || requestedAccountDomain == "" {
		l.Debug("aborting request because username or domain was empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	accountDomain := config.GetAccountDomain()
	host := config.GetHost()

	if requestedAccountDomain != accountDomain && requestedAccountDomain != host {
		l.Debugf("aborting request because accountDomain %s does not belong to this instance", requestedAccountDomain)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("accountDomain %s does not belong to this instance", requestedAccountDomain)})
		return
	}

	// transfer the signature verifier from the gin context to the request context
	ctx := c.Request.Context()
	verifier, signed := c.Get(string(ap.ContextRequestingPublicKeyVerifier))
	if signed {
		ctx = context.WithValue(ctx, ap.ContextRequestingPublicKeyVerifier, verifier)
	}

	resp, errWithCode := m.processor.GetWebfingerAccount(ctx, username)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, resp)
}

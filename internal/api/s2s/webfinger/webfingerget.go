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

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/util"
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

	requestedUsername, requestedHost, err := util.ExtractWebfingerParts(resourceQuery)
	if err != nil {
		l.Debugf("bad webfinger request with resource query %s: %s", resourceQuery, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("bad webfinger request with resource query %s", resourceQuery)})
		return
	}

	accountDomain := config.GetAccountDomain()
	host := config.GetHost()

	if requestedHost != host && requestedHost != accountDomain {
		l.Debugf("aborting request because requestedHost %s does not belong to this instance", requestedHost)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("requested host %s does not belong to this instance", requestedHost)})
		return
	}

	// transfer the signature verifier from the gin context to the request context
	ctx := c.Request.Context()
	verifier, signed := c.Get(string(ap.ContextRequestingPublicKeyVerifier))
	if signed {
		ctx = context.WithValue(ctx, ap.ContextRequestingPublicKeyVerifier, verifier)
	}

	resp, errWithCode := m.processor.GetWebfingerAccount(ctx, requestedUsername)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, resp)
}

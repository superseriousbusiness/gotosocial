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

package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
)

// CallbackGETHandler parses a token from an external auth provider.
func (m *Module) CallbackGETHandler(c *gin.Context) {
   s := sessions.Default(c)

	// first make sure the state set in the cookie is the same as the state returned from the external provider
	state := c.Query(callbackStateParam)
	if state == "" {
      m.clearSession(s)
		c.JSON(http.StatusForbidden, gin.H{"error": "state query not found on callback"})
		return
	}

	savedStateI := s.Get(sessionState)
	savedState, ok := savedStateI.(string)
	if !ok {
      m.clearSession(s)
		c.JSON(http.StatusForbidden, gin.H{"error": "state not found in session"})
		return
	}

	if state != savedState {
      m.clearSession(s)
		c.JSON(http.StatusForbidden, gin.H{"error": "state mismatch"})
		return
	}

	code := c.Query(callbackCodeParam)

	claims, err := m.idp.HandleCallback(c.Request.Context(), code)
	if err != nil {
      m.clearSession(s)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	user, err := m.parseUserFromClaims(claims)
	if err != nil {
      m.clearSession(s)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	s.Set(sessionUserID, user.ID)
	if err := s.Save(); err != nil {
      m.clearSession(s)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, OauthAuthorizePath)
}

func (m *Module) parseUserFromClaims(claims *oidc.Claims) (*gtsmodel.User, error) {
   if claims.Email == "" {
      return nil, errors.New("no email returned in claims")
   }
   // see if we already have a user for this email address


   if claims.Name == "" {
		return nil, errors.New("no name returned in claims")
	}
   username := ""
	nameParts := strings.Split(claims.Name, " ")
   for i, n := range nameParts {
aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
   }
}

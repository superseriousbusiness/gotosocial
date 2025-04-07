// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package auth

import (
	"errors"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (m *Module) mustClearSession(s sessions.Session) {
	s.Clear()
	m.mustSaveSession(s)
}

func (m *Module) mustSaveSession(s sessions.Session) {
	if err := s.Save(); err != nil {
		panic(err)
	}
}

// mustUserFromSession returns a *gtsmodel.User by checking the
// session for a user id and fetching the user from the database.
//
// On failure, the function clears session state, writes an internal
// error to the response writer, and returns nil. Callers should always
// return immediately if receiving nil back from this function!
func (m *Module) mustUserFromSession(
	c *gin.Context,
	s sessions.Session,
) *gtsmodel.User {
	// Try "userid" key first, fall
	// back to "userid_awaiting_2fa".
	var userID string
	for _, key := range [2]string{
		sessionUserID,
		sessionUserIDAwaiting2FA,
	} {
		var ok bool
		userID, ok = s.Get(key).(string)
		if ok && userID != "" {
			// Got it.
			break
		}
	}

	if userID == "" {
		const safe = "neither userid nor userid_awaiting_2fa keys found in session"
		m.clearSessionWithInternalError(c, s, errors.New(safe), safe, oauth.HelpfulAdvice)
		return nil
	}

	user, err := m.state.DB.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		safe := "db error getting user " + userID
		m.clearSessionWithInternalError(c, s, err, safe, oauth.HelpfulAdvice)
		return nil
	}

	return user
}

// mustAppFromSession returns a *gtsmodel.Application by checking the
// session for an application keyid and fetching the app from the database.
//
// On failure, the function clears session state, writes an internal
// error to the response writer, and returns nil. Callers should always
// return immediately if receiving nil back from this function!
func (m *Module) mustAppFromSession(
	c *gin.Context,
	s sessions.Session,
) *gtsmodel.Application {
	clientID, ok := s.Get(sessionClientID).(string)
	if !ok {
		const safe = "key client_id not found in session"
		m.clearSessionWithInternalError(c, s, errors.New(safe), safe, oauth.HelpfulAdvice)
		return nil
	}

	app, err := m.state.DB.GetApplicationByClientID(c.Request.Context(), clientID)
	if err != nil {
		safe := "db error getting app for clientID " + clientID
		m.clearSessionWithInternalError(c, s, err, safe, oauth.HelpfulAdvice)
		return nil
	}

	return app
}

// mustStringFromSession returns the string value
// corresponding to the given session key, if any is set.
//
// On failure (nothing set), the function clears session
// state, writes an internal error to the response writer,
// and returns nil. Callers should always return immediately
// if receiving nil back from this function!
func (m *Module) mustStringFromSession(
	c *gin.Context,
	s sessions.Session,
	key string,
) string {
	v, ok := s.Get(key).(string)
	if !ok {
		safe := "key " + key + " not found in session"
		m.clearSessionWithInternalError(c, s, errors.New(safe), safe, oauth.HelpfulAdvice)
		return ""
	}

	return v
}

func (m *Module) clearSessionWithInternalError(
	c *gin.Context,
	s sessions.Session,
	err error,
	helpText ...string,
) {
	m.mustClearSession(s)
	errWithCode := gtserror.NewErrorInternalError(err, helpText...)
	apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
}

func (m *Module) clearSessionWithBadRequest(
	c *gin.Context,
	s sessions.Session,
	err error,
	helpText ...string,
) {
	m.mustClearSession(s)
	errWithCode := gtserror.NewErrorBadRequest(err, helpText...)
	apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
}

/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (m *Module) OobHandler(c *gin.Context) {
	instance, errWithCode := m.processor.InstanceGetV1(c.Request.Context())
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	instanceGet := func(ctx context.Context) (*apimodel.InstanceV1, gtserror.WithCode) {
		return instance, nil
	}

	oobToken := c.Query("code")
	if oobToken == "" {
		err := errors.New("no 'code' query value provided in callback redirect")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error(), oauth.HelpfulAdvice), instanceGet)
		return
	}

	s := sessions.Default(c)

	errs := []string{}

	scope, ok := s.Get(sessionScope).(string)
	if !ok {
		errs = append(errs, fmt.Sprintf("key %s was not found in session", sessionScope))
	}

	userID, ok := s.Get(sessionUserID).(string)
	if !ok {
		errs = append(errs, fmt.Sprintf("key %s was not found in session", sessionUserID))
	}

	if len(errs) != 0 {
		errs = append(errs, oauth.HelpfulAdvice)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(errors.New("one or more missing keys on session during OobHandler"), errs...), m.processor.InstanceGetV1)
		return
	}

	user, err := m.db.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		m.clearSession(s)
		safe := fmt.Sprintf("user with id %s could not be retrieved", userID)
		var errWithCode gtserror.WithCode
		if err == db.ErrNoEntries {
			errWithCode = gtserror.NewErrorBadRequest(err, safe, oauth.HelpfulAdvice)
		} else {
			errWithCode = gtserror.NewErrorInternalError(err, safe, oauth.HelpfulAdvice)
		}
		apiutil.ErrorHandler(c, errWithCode, instanceGet)
		return
	}

	acct, err := m.db.GetAccountByID(c.Request.Context(), user.AccountID)
	if err != nil {
		m.clearSession(s)
		safe := fmt.Sprintf("account with id %s could not be retrieved", user.AccountID)
		var errWithCode gtserror.WithCode
		if err == db.ErrNoEntries {
			errWithCode = gtserror.NewErrorBadRequest(err, safe, oauth.HelpfulAdvice)
		} else {
			errWithCode = gtserror.NewErrorInternalError(err, safe, oauth.HelpfulAdvice)
		}
		apiutil.ErrorHandler(c, errWithCode, instanceGet)
		return
	}

	// we're done with the session now, so just clear it out
	m.clearSession(s)

	c.HTML(http.StatusOK, "oob.tmpl", gin.H{
		"instance": instance,
		"user":     acct.Username,
		"oobToken": oobToken,
		"scope":    scope,
	})
}

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
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"golang.org/x/crypto/bcrypt"
)

// login just wraps a form-submitted username (we want an email) and password
type login struct {
	Email    string `form:"username"`
	Password string `form:"password"`
}

// SignInGETHandler should be served at https://example.org/auth/sign_in.
// The idea is to present a sign in page to the user, where they can enter their username and password.
// The form will then POST to the sign in page, which will be handled by SignInPOSTHandler.
// If an idp provider is set, then the user will be redirected to that to do their sign in.
func (m *Module) SignInGETHandler(c *gin.Context) {
	if _, err := apiutil.NegotiateAccept(c, apiutil.HTMLAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if !config.GetOIDCEnabled() {
		instance, errWithCode := m.processor.InstanceGetV1(c.Request.Context())
		if errWithCode != nil {
			apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
			return
		}

		// no idp provider, use our own funky little sign in page
		c.HTML(http.StatusOK, "sign-in.tmpl", gin.H{
			"instance": instance,
		})
		return
	}

	// idp provider is in use, so redirect to it
	s := sessions.Default(c)

	internalStateI := s.Get(sessionInternalState)
	internalState, ok := internalStateI.(string)
	if !ok {
		m.clearSession(s)
		err := fmt.Errorf("key %s was not found in session", sessionInternalState)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	c.Redirect(http.StatusSeeOther, m.idp.AuthCodeURL(internalState))
}

// SignInPOSTHandler should be served at https://example.org/auth/sign_in.
// The idea is to present a sign in page to the user, where they can enter their username and password.
// The handler will then redirect to the auth handler served at /auth
func (m *Module) SignInPOSTHandler(c *gin.Context) {
	s := sessions.Default(c)

	form := &login{}
	if err := c.ShouldBind(form); err != nil {
		m.clearSession(s)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, oauth.HelpfulAdvice), m.processor.InstanceGetV1)
		return
	}

	userid, errWithCode := m.ValidatePassword(c.Request.Context(), form.Email, form.Password)
	if errWithCode != nil {
		// don't clear session here, so the user can just press back and try again
		// if they accidentally gave the wrong password or something
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	s.Set(sessionUserID, userid)
	if err := s.Save(); err != nil {
		err := fmt.Errorf("error saving user id onto session: %s", err)
		apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err, oauth.HelpfulAdvice), m.processor.InstanceGetV1)
	}

	c.Redirect(http.StatusFound, "/oauth"+OauthAuthorizePath)
}

// ValidatePassword takes an email address and a password.
// The goal is to authenticate the password against the one for that email
// address stored in the database. If OK, we return the userid (a ulid) for that user,
// so that it can be used in further Oauth flows to generate a token/retreieve an oauth client from the db.
func (m *Module) ValidatePassword(ctx context.Context, email string, password string) (string, gtserror.WithCode) {
	if email == "" || password == "" {
		err := errors.New("email or password was not provided")
		return incorrectPassword(err)
	}

	user, err := m.db.GetUserByEmailAddress(ctx, email)
	if err != nil {
		err := fmt.Errorf("user %s was not retrievable from db during oauth authorization attempt: %s", email, err)
		return incorrectPassword(err)
	}

	if user.EncryptedPassword == "" {
		err := fmt.Errorf("encrypted password for user %s was empty for some reason", user.Email)
		return incorrectPassword(err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword), []byte(password)); err != nil {
		err := fmt.Errorf("password hash didn't match for user %s during login attempt: %s", user.Email, err)
		return incorrectPassword(err)
	}

	return user.ID, nil
}

// incorrectPassword wraps the given error in a gtserror.WithCode, and returns
// only a generic 'safe' error message to the user, to not give any info away.
func incorrectPassword(err error) (string, gtserror.WithCode) {
	safeErr := fmt.Errorf("password/email combination was incorrect")
	return "", gtserror.NewErrorUnauthorized(err, safeErr.Error(), oauth.HelpfulAdvice)
}

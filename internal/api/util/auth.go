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

package util

import (
	"errors"
	"slices"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/oauth2/v4"
	"github.com/gin-gonic/gin"
)

// Auth wraps an authorized token, application, user, and account.
// It is used in the functions GetAuthed and MustAuth.
// Because the user might *not* be authed, any of the fields in this struct
// might be nil, so make sure to check that when you're using this struct anywhere.
type Auth struct {
	Token       oauth2.TokenInfo
	Application *gtsmodel.Application
	User        *gtsmodel.User
	Account     *gtsmodel.Account
}

// TokenAuth is a convenience function for returning an TokenAuth struct from a gin context.
// In essence, it tries to extract a token, application, user, and account from the context,
// and then sets them on a struct for convenience.
//
// If any are not present in the context, they will be set to nil on the returned TokenAuth struct.
//
// If *ALL* are not present, then nil and an error will be returned.
//
// If something goes wrong during parsing, then nil and an error will be returned (consider this not authed).
// TokenAuth is like GetAuthed, but will fail if one of the requirements is not met.
func TokenAuth(
	c *gin.Context,
	requireToken bool,
	requireApp bool,
	requireUser bool,
	requireAccount bool,
	requireScope ...Scope,
) (*Auth, gtserror.WithCode) {
	var (
		ctx = c.Copy()
		a   = &Auth{}
		i   interface{}
		ok  bool
	)

	i, ok = ctx.Get(oauth.SessionAuthorizedToken)
	if ok {
		parsed, ok := i.(oauth2.TokenInfo)
		if !ok {
			const errText = "could not parse token from session context"
			return nil, gtserror.NewErrorUnauthorized(errors.New(errText), errText)
		}
		a.Token = parsed
	}

	i, ok = ctx.Get(oauth.SessionAuthorizedApplication)
	if ok {
		parsed, ok := i.(*gtsmodel.Application)
		if !ok {
			const errText = "could not parse application from session context"
			return nil, gtserror.NewErrorUnauthorized(errors.New(errText), errText)
		}
		a.Application = parsed
	}

	i, ok = ctx.Get(oauth.SessionAuthorizedUser)
	if ok {
		parsed, ok := i.(*gtsmodel.User)
		if !ok {
			const errText = "could not parse user from session context"
			return nil, gtserror.NewErrorUnauthorized(errors.New(errText), errText)
		}
		a.User = parsed
	}

	i, ok = ctx.Get(oauth.SessionAuthorizedAccount)
	if ok {
		parsed, ok := i.(*gtsmodel.Account)
		if !ok {
			const errText = "could not parse account from session context"
			return nil, gtserror.NewErrorUnauthorized(errors.New(errText), errText)
		}
		a.Account = parsed
	}

	if requireToken && a.Token == nil {
		const errText = "token not supplied"
		return nil, gtserror.NewErrorUnauthorized(errors.New(errText), errText)
	}

	if requireApp && a.Application == nil {
		const errText = "application not supplied"
		return nil, gtserror.NewErrorUnauthorized(errors.New(errText), errText)
	}

	if requireUser && a.User == nil {
		const errText = "user not supplied or not authorized"
		return nil, gtserror.NewErrorUnauthorized(errors.New(errText), errText)
	}

	if requireAccount && a.Account == nil {
		const errText = "account not supplied or not authorized"
		return nil, gtserror.NewErrorUnauthorized(errors.New(errText), errText)
	}

	if len(requireScope) != 0 {
		// We need to match one of the
		// required scopes, check if we can.
		hasScopes := strings.Split(a.Token.GetScope(), " ")
		scopeOK := slices.ContainsFunc(
			hasScopes,
			func(hasScope string) bool {
				for _, requiredScope := range requireScope {
					if Scope(hasScope).Permits(requiredScope) {
						// Got it.
						return true
					}
				}
				return false
			},
		)

		if !scopeOK {
			const errText = "token has insufficient scope permission"
			return nil, gtserror.NewErrorForbidden(errors.New(errText), errText)
		}
	}

	return a, nil
}

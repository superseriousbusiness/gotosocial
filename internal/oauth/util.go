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

package oauth

import (
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/oauth2/v4"
	"github.com/superseriousbusiness/oauth2/v4/errors"
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

// Authed is a convenience function for returning an Authed struct from a gin context.
// In essence, it tries to extract a token, application, user, and account from the context,
// and then sets them on a struct for convenience.
//
// If any are not present in the context, they will be set to nil on the returned Authed struct.
//
// If *ALL* are not present, then nil and an error will be returned.
//
// If something goes wrong during parsing, then nil and an error will be returned (consider this not authed).
// Authed is like GetAuthed, but will fail if one of the requirements is not met.
func Authed(c *gin.Context, requireToken bool, requireApp bool, requireUser bool, requireAccount bool) (*Auth, error) {
	ctx := c.Copy()
	a := &Auth{}
	var i interface{}
	var ok bool

	i, ok = ctx.Get(SessionAuthorizedToken)
	if ok {
		parsed, ok := i.(oauth2.TokenInfo)
		if !ok {
			return nil, errors.New("could not parse token from session context")
		}
		a.Token = parsed
	}

	i, ok = ctx.Get(SessionAuthorizedApplication)
	if ok {
		parsed, ok := i.(*gtsmodel.Application)
		if !ok {
			return nil, errors.New("could not parse application from session context")
		}
		a.Application = parsed
	}

	i, ok = ctx.Get(SessionAuthorizedUser)
	if ok {
		parsed, ok := i.(*gtsmodel.User)
		if !ok {
			return nil, errors.New("could not parse user from session context")
		}
		a.User = parsed
	}

	i, ok = ctx.Get(SessionAuthorizedAccount)
	if ok {
		parsed, ok := i.(*gtsmodel.Account)
		if !ok {
			return nil, errors.New("could not parse account from session context")
		}
		a.Account = parsed
	}

	if requireToken && a.Token == nil {
		return nil, errors.New("token not supplied")
	}

	if requireApp && a.Application == nil {
		return nil, errors.New("application not supplied")
	}

	if requireUser && a.User == nil {
		return nil, errors.New("user not supplied or not authorized")
	}

	if requireAccount && a.Account == nil {
		return nil, errors.New("account not supplied or not authorized")
	}

	return a, nil
}

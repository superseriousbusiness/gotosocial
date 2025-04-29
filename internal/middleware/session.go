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

package middleware

import (
	"fmt"
	"net/url"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/idna"
)

// SessionOptions returns the standard set of options to use for each session.
func SessionOptions(cookiePolicy apiutil.CookiePolicy) sessions.Options {
	return sessions.Options{
		Path:   "/",
		Domain: cookiePolicy.Domain,
		// 2 minutes
		MaxAge:   120,
		Secure:   cookiePolicy.Secure,
		HttpOnly: cookiePolicy.HTTPOnly,
		SameSite: cookiePolicy.SameSite,
	}
}

// SessionName is a utility function that derives an appropriate session name from the hostname.
func SessionName() (string, error) {
	// parse the protocol + host
	protocol := config.GetProtocol()
	host := config.GetHost()
	u, err := url.Parse(fmt.Sprintf("%s://%s", protocol, host))
	if err != nil {
		return "", err
	}

	// take the hostname without any port attached
	strippedHostname := u.Hostname()
	if strippedHostname == "" {
		return "", fmt.Errorf("could not derive hostname without port from %s://%s", protocol, host)
	}

	// make sure IDNs are converted to punycode or the cookie library breaks:
	// see https://en.wikipedia.org/wiki/Punycode
	punyHostname, err := idna.New().ToASCII(strippedHostname)
	if err != nil {
		return "", fmt.Errorf("could not convert %s to punycode: %s", strippedHostname, err)
	}

	return fmt.Sprintf("gotosocial-%s", punyHostname), nil
}

// Session returns a new gin middleware that implements session cookies using the given sessionName, authentication
// key, and encryption key. Session name can be derived from the SessionName utility function in this package.
func Session(sessionName string, auth []byte, crypt []byte, cookiePolicy apiutil.CookiePolicy) gin.HandlerFunc {
	store := memstore.NewStore(auth, crypt)
	store.Options(SessionOptions(cookiePolicy))
	return sessions.Sessions(sessionName, store)
}

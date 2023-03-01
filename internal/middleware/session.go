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

package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"golang.org/x/net/idna"
)

// SessionOptions returns the standard set of options to use for each session.
func SessionOptions() sessions.Options {
	var samesite http.SameSite
	switch strings.TrimSpace(strings.ToLower(config.GetAdvancedCookiesSamesite())) {
	case "lax":
		samesite = http.SameSiteLaxMode
	case "strict":
		samesite = http.SameSiteStrictMode
	default:
		log.Warnf(nil, "%s set to %s which is not recognized, defaulting to 'lax'", config.AdvancedCookiesSamesiteFlag(), config.GetAdvancedCookiesSamesite())
		samesite = http.SameSiteLaxMode
	}

	return sessions.Options{
		Path:   "/",
		Domain: config.GetHost(),
		// 2 minutes
		MaxAge: 120,
		// only set secure over https
		Secure: config.GetProtocol() == "https",
		// forbid javascript from inspecting cookie
		HttpOnly: true,
		// https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-cookie-same-site-00#section-4.1.1
		SameSite: samesite,
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

// Session returns a new gin middleware that implements session cookies using the given
// sessionName, authentication key, and encryption key. Session name can be derived from the
// SessionName utility function in this package.
func Session(sessionName string, auth []byte, crypt []byte) gin.HandlerFunc {
	store := memstore.NewStore(auth, crypt)
	store.Options(SessionOptions())
	return sessions.Sessions(sessionName, store)
}

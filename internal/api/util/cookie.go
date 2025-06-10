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
	"net/http"
	"net/url"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/gin-gonic/gin"
)

// CookiePolicy encompasses a number
// of security related cookie directives
// of which we want to be set consistently
// on all cookies administered by us.
type CookiePolicy struct {
	Domain   string
	SameSite http.SameSite
	HTTPOnly bool
	Secure   bool
}

// NewCookiePolicy will return a new CookiePolicy{}
// object setup according to current instance config.
func NewCookiePolicy() CookiePolicy {
	var sameSite http.SameSite
	switch s := config.GetAdvancedCookiesSamesite(); s {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "lax":
		sameSite = http.SameSiteLaxMode
	default:
		log.Warnf(nil, "%s set to %s which is not recognized, defaulting to 'lax'", config.AdvancedCookiesSamesiteFlag, s)
		sameSite = http.SameSiteLaxMode
	}
	return CookiePolicy{
		Domain: config.GetHost(),

		// https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-cookie-same-site-00#section-4.1.1
		SameSite: sameSite,

		// forbid javascript from
		// inspecting cookie
		HTTPOnly: true,

		// only set secure cookie directive over https
		Secure: (config.GetProtocol() == "https"),
	}
}

// SetCookie will set the given cookie details according to currently configured CookiePolicy{}.
func (p *CookiePolicy) SetCookie(c *gin.Context, name, value string, maxAge int, path string) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   p.Domain,
		SameSite: p.SameSite,
		Secure:   p.Secure,
		HttpOnly: p.HTTPOnly,
	})
}

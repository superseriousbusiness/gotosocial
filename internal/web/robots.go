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

package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	robotsPath          = "/robots.txt"
	robotsMetaAllowSome = "nofollow, noarchive, nositelinkssearchbox, max-image-preview:standard" // https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag#robotsmeta
	robotsTxt           = `# GoToSocial robots.txt -- to edit, see internal/web/robots.go
# more info @ https://developers.google.com/search/docs/crawling-indexing/robots/intro
User-agent: *
Crawl-delay: 500
# api stuff
Disallow: /api/
# auth/login stuff
Disallow: /auth/
Disallow: /oauth/
Disallow: /check_your_email
Disallow: /wait_for_approval
Disallow: /account_disabled
# well known stuff
Disallow: /.well-known/
# files
Disallow: /fileserver/
# s2s AP stuff
Disallow: /users/
Disallow: /emoji/
# panels
Disallow: /admin
Disallow: /user
Disallow: /settings/
# domain blocklist
Disallow: /about/suspended`
)

// robotsGETHandler returns a decent robots.txt that prevents crawling
// the api, auth pages, settings pages, etc.
//
// More granular robots meta tags are then applied for web pages
// depending on user preferences (see internal/web).
func (m *Module) robotsGETHandler(c *gin.Context) {
	c.String(http.StatusOK, robotsTxt)
}

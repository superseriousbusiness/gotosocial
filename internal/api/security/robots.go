package security

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const robotsString = `User-agent: *
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
`

// RobotsGETHandler returns a decent robots.txt that prevents crawling
// the api, auth pages, settings pages, etc.
//
// More granular robots meta tags are then applied for web pages
// depending on user preferences (see internal/web).
func (m *Module) RobotsGETHandler(c *gin.Context) {
	c.String(http.StatusOK, robotsString)
}

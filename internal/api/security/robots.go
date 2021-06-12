package security

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const robotsString = `User-agent: *
Disallow: /
`

// RobotsGETHandler returns the most restrictive possible robots.txt file in response to a call to /robots.txt.
// The response instructs bots with *any* user agent not to index the instance at all.
func (m *Module) RobotsGETHandler(c *gin.Context) {
	c.String(http.StatusOK, robotsString)
}

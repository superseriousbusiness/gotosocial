package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// DomainBlocksGETHandler swagger:operation GET /api/v1/admin/domain_blocks domainBlocksGet
//
// View all domain blocks currently in place.
//
// ---
// tags:
// - admin
//
// produces:
// - application/json
//
// parameters:
// - name: export
//   type: boolean
//   description: |-
//     If set to true, then each entry in the returned list of domain blocks will only consist of
//     the fields 'domain' and 'public_comment'. This is perfect for when you want to save and share
//     a list of all the domains you have blocked on your instance, so that someone else can easily import them,
//     but you don't need them to see the database IDs of your blocks, or private comments etc.
//   in: query
//   required: false
//
// security:
// - OAuth2 Bearer:
//   - admin
//
// responses:
//   '200':
//     description: All domain blocks currently in place.
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/domainBlock"
//   '403':
//      description: forbidden
//   '400':
//      description: bad request
//   '404':
//      description: not found
func (m *Module) DomainBlocksGETHandler(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func":        "DomainBlocksGETHandler",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})

	// make sure we're authed with an admin account
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	if !authed.User.Admin {
		l.Debugf("user %s not an admin", authed.User.ID)
		c.JSON(http.StatusForbidden, gin.H{"error": "not an admin"})
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	export := false
	exportString := c.Query(ExportQueryKey)
	if exportString != "" {
		i, err := strconv.ParseBool(exportString)
		if err != nil {
			l.Debugf("error parsing export string: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't parse export query param"})
			return
		}
		export = i
	}

	domainBlocks, err := m.processor.AdminDomainBlocksGet(c.Request.Context(), authed, export)
	if err != nil {
		l.Debugf("error getting domain blocks: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, domainBlocks)
}

package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// DomainBlockGETHandler swagger:operation GET /api/v1/admin/domain_blocks/{id} domainBlockGet
//
// View domain block with the given ID.
//
// ---
// tags:
// - admin
//
// produces:
// - application/json
//
// parameters:
// - name: id
//   type: string
//   description: The id of the domain block.
//   in: path
//   required: true
//
// security:
// - OAuth2 Bearer:
//   - admin
//
// responses:
//   '200':
//     description: The requested domain block.
//     schema:
//       "$ref": "#/definitions/domainBlock"
//   '403':
//      description: forbidden
//   '400':
//      description: bad request
//   '404':
//      description: not found
func (m *Module) DomainBlockGETHandler(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func":        "DomainBlockGETHandler",
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

	domainBlockID := c.Param(IDKey)
	if domainBlockID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no domain block id provided"})
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

	domainBlock, err := m.processor.AdminDomainBlockGet(c.Request.Context(), authed, domainBlockID, export)
	if err != nil {
		l.Debugf("error getting domain block: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, domainBlock)
}

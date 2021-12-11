package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// DomainBlockDELETEHandler swagger:operation DELETE /api/v1/admin/domain_blocks/{id} domainBlockDelete
//
// Delete domain block with the given ID.
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
//     description: The domain block that was just deleted.
//     schema:
//       "$ref": "#/definitions/domainBlock"
//   '403':
//      description: forbidden
//   '400':
//      description: bad request
//   '404':
//      description: not found
func (m *Module) DomainBlockDELETEHandler(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func":        "DomainBlockDELETEHandler",
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

	domainBlock, errWithCode := m.processor.AdminDomainBlockDelete(c.Request.Context(), authed, domainBlockID)
	if errWithCode != nil {
		l.Debugf("error deleting domain block: %s", errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, domainBlock)
}

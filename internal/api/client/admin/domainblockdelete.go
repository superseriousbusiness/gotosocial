package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// DomainBlockDELETEHandler deals with the delete of an existing domain block.
func (m *Module) DomainBlockDELETEHandler(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
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

	domainBlockID := c.Param(IDKey)
	if domainBlockID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no domain block id provided"})
		return
	}

	domainBlock, errWithCode := m.processor.AdminDomainBlockDelete(authed, domainBlockID)
	if errWithCode != nil {
		l.Debugf("error deleting domain block: %s", errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, domainBlock)
}

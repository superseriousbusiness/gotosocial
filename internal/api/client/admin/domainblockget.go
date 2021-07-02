package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (m *Module) DomainBlocksGETHandler(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
		"func":        "DomainBlocksPOSTHandler",
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

	domainBlocks, err := m.processor.AdminDomainBlocksGet(authed, export)
	if err != nil {
		l.Debugf("error getting domain blocks: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, domainBlocks)
}

package account

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// AccountRelationshipsGETHandler serves the relationship of the requesting account with one or more requested account IDs.
func (m *Module) AccountRelationshipsGETHandler(c *gin.Context) {
	l := m.log.WithField("func", "AccountRelationshipsGETHandler")

	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("error authing: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	targetAccountIDs := c.QueryArray("id[]")
	if len(targetAccountIDs) == 0 {
		l.Debug("no account id specified in query")
		c.JSON(http.StatusBadRequest, gin.H{"error": "no account id specified"})
		return
	}

	relationships := []model.Relationship{}

	for _, targetAccountID := range targetAccountIDs {
		r, errWithCode := m.processor.AccountRelationshipGet(authed, targetAccountID)
		if err != nil {
			c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
			return
		}
		relationships = append(relationships, *r)
	}

	c.JSON(http.StatusOK, relationships)
}

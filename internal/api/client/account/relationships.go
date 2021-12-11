package account

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// AccountRelationshipsGETHandler swagger:operation GET /api/v1/accounts/relationships accountRelationships
//
// See your account's relationships with the given account IDs.
//
// ---
// tags:
// - accounts
//
// produces:
// - application/json
//
// parameters:
// - name: id
//   type: array
//   items:
//     type: string
//   description: Account IDs.
//   in: query
//   required: true
//
// security:
// - OAuth2 Bearer:
//   - read:accounts
//
// responses:
//   '200':
//     name: account relationships
//     description: Array of account relationships.
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/accountRelationship"
//   '401':
//      description: unauthorized
//   '400':
//      description: bad request
//   '404':
//      description: not found
func (m *Module) AccountRelationshipsGETHandler(c *gin.Context) {
	l := logrus.WithField("func", "AccountRelationshipsGETHandler")

	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("error authing: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	targetAccountIDs := c.QueryArray("id[]")
	if len(targetAccountIDs) == 0 {
		// check fallback -- let's be generous and see if maybe it's just set as 'id'?
		id := c.Query("id")
		if id == "" {
			l.Debug("no account id specified in query")
			c.JSON(http.StatusBadRequest, gin.H{"error": "no account id specified"})
			return
		}
		targetAccountIDs = append(targetAccountIDs, id)
	}

	relationships := []model.Relationship{}

	for _, targetAccountID := range targetAccountIDs {
		r, errWithCode := m.processor.AccountRelationshipGet(c.Request.Context(), authed, targetAccountID)
		if err != nil {
			c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
			return
		}
		relationships = append(relationships, *r)
	}

	c.JSON(http.StatusOK, relationships)
}

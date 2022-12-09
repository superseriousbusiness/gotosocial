package accounts

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// AccountRelationshipsGETHandler swagger:operation GET /api/v1/accounts/relationships accountRelationships
//
// See your account's relationships with the given account IDs.
//
//	---
//	tags:
//	- accounts
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id
//		type: array
//		items:
//			type: string
//		description: Account IDs.
//		in: query
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- read:accounts
//
//	responses:
//		'200':
//			name: account relationships
//			description: Array of account relationships.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/accountRelationship"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) AccountRelationshipsGETHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	targetAccountIDs := c.QueryArray("id[]")
	if len(targetAccountIDs) == 0 {
		// check fallback -- let's be generous and see if maybe it's just set as 'id'?
		id := c.Query("id")
		if id == "" {
			err = errors.New("no account id(s) specified in query")
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
			return
		}
		targetAccountIDs = append(targetAccountIDs, id)
	}

	relationships := []apimodel.Relationship{}

	for _, targetAccountID := range targetAccountIDs {
		r, errWithCode := m.processor.AccountRelationshipGet(c.Request.Context(), authed, targetAccountID)
		if errWithCode != nil {
			apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
			return
		}
		relationships = append(relationships, *r)
	}

	c.JSON(http.StatusOK, relationships)
}

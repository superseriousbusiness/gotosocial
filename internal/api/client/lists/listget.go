package lists

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// ListGETHandler swagger:operation GET /api/v1/list/{id} list
//
// Get a single list with the given ID.
//
//	---
//	tags:
//	- lists
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- read:lists
//
//	responses:
//		'200':
//			name: list
//			description: Requested list.
//			schema:
//				"$ref": "#/definitions/list"
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
func (m *Module) ListGETHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	targetListID := c.Param(IDKey)
	if targetListID == "" {
		err := errors.New("no list id specified")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.List().Get(c.Request.Context(), authed.Account, targetListID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	c.JSON(http.StatusOK, resp)
}

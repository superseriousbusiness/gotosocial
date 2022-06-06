package admin

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
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
//   '400':
//      description: bad request
//   '401':
//      description: unauthorized
//   '403':
//      description: forbidden
//   '404':
//      description: not found
//   '406':
//      description: not acceptable
//   '500':
//      description: internal server error
func (m *Module) DomainBlockDELETEHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if !authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		api.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	domainBlockID := c.Param(IDKey)
	if domainBlockID == "" {
		err := errors.New("no domain block id specified")
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	domainBlock, errWithCode := m.processor.AdminDomainBlockDelete(c.Request.Context(), authed, domainBlockID)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, domainBlock)
}

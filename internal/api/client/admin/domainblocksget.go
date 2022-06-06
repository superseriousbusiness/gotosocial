package admin

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
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
func (m *Module) DomainBlocksGETHandler(c *gin.Context) {
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

	export := false
	exportString := c.Query(ExportQueryKey)
	if exportString != "" {
		i, err := strconv.ParseBool(exportString)
		if err != nil {
			err := fmt.Errorf("error parsing %s: %s", ExportQueryKey, err)
			api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
			return
		}
		export = i
	}

	domainBlocks, errWithCode := m.processor.AdminDomainBlocksGet(c.Request.Context(), authed, export)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, domainBlocks)
}

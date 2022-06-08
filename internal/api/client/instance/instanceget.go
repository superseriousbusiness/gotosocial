package instance

import (
	"net/http"

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"

	"github.com/gin-gonic/gin"
)

// InstanceInformationGETHandler swagger:operation GET /api/v1/instance instanceGet
//
// View instance information.
//
// This is mostly provided for Mastodon application compatibility, since many apps that work with Mastodon use `/api/v1/instance` to inform their connection parameters.
//
// However, it can also be used by other instances for gathering instance information and representing instances in some UI or other.
//
// ---
// tags:
// - instance
//
// produces:
// - application/json
//
// responses:
//   '200':
//     description: "Instance information."
//     schema:
//       "$ref": "#/definitions/instance"
//   '406':
//      description: not acceptable
//   '500':
//      description: internal error
func (m *Module) InstanceInformationGETHandler(c *gin.Context) {
	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	instance, errWithCode := m.processor.InstanceGet(c.Request.Context(), config.GetHost())
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, instance)
}

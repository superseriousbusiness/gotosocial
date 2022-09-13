package emoji

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// EmojisGETHandler swagger:operation GET /api/v1/custom_emojis customEmojisGet
//
// Get an array of custom emojis available on the instance.
//
// ---
// tags:
// - custom_emojis
//
// produces:
// - application/json
//
// security:
// - OAuth2 Bearer:
//   - read:custom_emojis
//
// responses:
//   '200':
//     description: Array of custom emojis.
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/emoji"
//   '401':
//      description: unauthorized
//   '406':
//      description: not acceptable
//   '500':
//      description: internal server error
func (m *Module) EmojisGETHandler(c *gin.Context) {
	if _, err := oauth.Authed(c, true, true, true, true); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	emojis, errWithCode := m.processor.CustomEmojisGet(c)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, emojis)
}

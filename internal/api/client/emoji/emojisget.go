package emoji

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// EmojisGETHandler returns a list of custom emojis enabled on the instance
func (m *Module) EmojisGETHandler(c *gin.Context) {
	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	emojis, errWithCode := m.processor.CustomEmojisGet(c)
	if errWithCode != nil {
		util.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, emojis)
}

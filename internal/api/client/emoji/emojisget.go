package emoji

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
)

// EmojisGETHandler returns a list of custom emojis enabled on the instance
func (m *Module) EmojisGETHandler(c *gin.Context) {
	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	emojis, errWithCode := m.processor.CustomEmojisGet(c)
	if errWithCode != nil {
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, emojis)
}

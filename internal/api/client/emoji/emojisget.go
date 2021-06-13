package emoji

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// EmojisGETHandler returns a list of custom emojis enabled on the instance
func (m *Module) EmojisGETHandler(c *gin.Context) {
	c.JSON(http.StatusOK, []string{})
}

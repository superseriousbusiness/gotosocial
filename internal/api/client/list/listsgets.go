package list

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListsGETHandler returns a list of lists created by/for the authed account
func (m *Module) ListsGETHandler(c *gin.Context) {
	c.JSON(http.StatusOK, []string{})
}

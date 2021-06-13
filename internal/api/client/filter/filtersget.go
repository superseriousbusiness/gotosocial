package filter

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// FiltersGETHandler returns a list of filters set by/for the authed account
func (m *Module) FiltersGETHandler(c *gin.Context) {
	c.JSON(http.StatusOK, []string{})
}

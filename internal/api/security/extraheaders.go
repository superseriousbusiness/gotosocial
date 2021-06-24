package security

import "github.com/gin-gonic/gin"

// ExtraHeaders adds any additional required headers to the response
func (m *Module) ExtraHeaders(c *gin.Context) {
	c.Header("Server", "gotosocial")
}

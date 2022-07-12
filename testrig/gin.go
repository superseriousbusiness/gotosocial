package testrig

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

// CreateGinTextContext creates a new gin.Context suitable for a test, with an instantiated gin.Engine.
func CreateGinTestContext(rw http.ResponseWriter, r *http.Request) (*gin.Context, *gin.Engine) {
	ctx, eng := gin.CreateTestContext(rw)
	router.LoadTemplateFunctions(eng)
	if err := router.LoadTemplates(eng); err != nil {
		panic(err)
	}
	ctx.Request = r
	return ctx, eng
}

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func statusGet(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title": "Posts",
	})
}

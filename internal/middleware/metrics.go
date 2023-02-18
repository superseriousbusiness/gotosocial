package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/metrics"
)

func Instrument(m *metrics.EndpointMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == metrics.HandlerPath {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		path := c.FullPath()
		method := c.Request.Method
		if method == "" {
			method = http.MethodGet
		}
		elapsed := time.Since(start)
		code := c.Writer.Status()
		if code == 0 {
			code = http.StatusInternalServerError
		}
		m.Inc(method, path, code)
		m.Observe(method, path, code, elapsed)
	}
}

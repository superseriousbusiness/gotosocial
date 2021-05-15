package instance

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// InstanceInformationGETHandler is for serving instance information at /api/v1/instance
func (m *Module) InstanceInformationGETHandler(c *gin.Context) {
	l := m.log.WithField("func", "InstanceInformationGETHandler")

	instance, err := m.processor.InstanceGet(m.config.Host)
	if err != nil {
		l.Debugf("error getting instance from processor: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, instance)
}

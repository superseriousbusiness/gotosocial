package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// StatusGETHandler serves the target status as an activitystreams NOTE so that other AP servers can parse it.
func (m *Module) StatusGETHandler(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
		"func": "StatusGETHandler",
		"url":  c.Request.RequestURI,
	})

	requestedUsername := c.Param(UsernameKey)
	if requestedUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no username specified in request"})
		return
	}

	requestedStatusID := c.Param(StatusIDKey)
	if requestedStatusID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no status id specified in request"})
		return
	}

	format, err := negotiateFormat(c)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": fmt.Sprintf("could not negotiate format with given Accept header(s): %s", err)})
		return
	}
	l.Tracef("negotiated format: %s", format)

	ctx := populateContext(c)

	status, errWithCode := m.processor.GetFediStatus(ctx, requestedUsername, requestedStatusID, c.Request.URL)
	if errWithCode != nil {
		l.Info(errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	b, mErr := json.Marshal(status)
	if mErr != nil {
		err := fmt.Errorf("could not marshal json: %s", mErr)
		l.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, format, b)
}

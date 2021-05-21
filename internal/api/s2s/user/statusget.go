package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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

	// make sure this actually an AP request
	format := c.NegotiateFormat(ActivityPubAcceptHeaders...)
	if format == "" {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "could not negotiate format with given Accept header(s)"})
		return
	}
	l.Tracef("negotiated format: %s", format)

	// make a copy of the context to pass along so we don't break anything
	cp := c.Copy()
	status, err := m.processor.GetFediStatus(requestedUsername, requestedStatusID, cp.Request) // handles auth as well
	if err != nil {
		l.Info(err.Error())
		c.JSON(err.Code(), gin.H{"error": err.Safe()})
		return
	}

	c.JSON(http.StatusOK, status)
}

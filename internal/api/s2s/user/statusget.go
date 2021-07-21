package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/util"
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

	// make sure this actually an AP request
	format := c.NegotiateFormat(ActivityPubAcceptHeaders...)
	if format == "" {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "could not negotiate format with given Accept header(s)"})
		return
	}
	l.Tracef("negotiated format: %s", format)

	// transfer the signature verifier from the gin context to the request context
	ctx := c.Request.Context()
	verifier, signed := c.Get(string(util.APRequestingPublicKeyVerifier))
	if signed {
		ctx = context.WithValue(ctx, util.APRequestingPublicKeyVerifier, verifier)
	}

	status, err := m.processor.GetFediStatus(ctx, requestedUsername, requestedStatusID, c.Request.URL) // handles auth as well
	if err != nil {
		l.Info(err.Error())
		c.JSON(err.Code(), gin.H{"error": err.Safe()})
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

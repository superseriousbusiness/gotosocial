package user

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// PublicKeyGETHandler should be served at eg https://example.org/users/:username/main-key.
//
// The goal here is to return a MINIMAL activitypub representation of an account
// in the form of a vocab.ActivityStreamsPerson. The account will only contain the id,
// public key, username, and type of the account.
func (m *Module) PublicKeyGETHandler(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
		"func": "PublicKeyGETHandler",
		"url":  c.Request.RequestURI,
	})

	requestedUsername := c.Param(UsernameKey)
	if requestedUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no username specified in request"})
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

	user, err := m.processor.GetFediUser(ctx, requestedUsername, c.Request.URL) // GetFediUser handles auth as well
	if err != nil {
		l.Info(err.Error())
		c.JSON(err.Code(), gin.H{"error": err.Safe()})
		return
	}

	c.JSON(http.StatusOK, user)
}

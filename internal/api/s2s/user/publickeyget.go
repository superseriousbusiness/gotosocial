package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

	format, err := negotiateFormat(c)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": fmt.Sprintf("could not negotiate format with given Accept header(s): %s", err)})
		return
	}
	l.Tracef("negotiated format: %s", format)

	ctx := populateContext(c)

	user, errWithCode := m.processor.GetFediUser(ctx, requestedUsername, c.Request.URL)
	if errWithCode != nil {
		l.Info(errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	b, mErr := json.Marshal(user)
	if mErr != nil {
		err := fmt.Errorf("could not marshal json: %s", mErr)
		l.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, format, b)
}

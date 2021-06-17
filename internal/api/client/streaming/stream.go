package streaming

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func (m *Module) StreamGETHandler(c *gin.Context) {
	l := m.log.WithField("func", "StreamGETHandler")

	streamType := c.Query(StreamQueryKey)
	if streamType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("no stream type provided under query key %s", StreamQueryKey)})
		return
	}

	accessToken := c.Query(AccessTokenQueryKey)
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("no access token provided under query key %s", AccessTokenQueryKey)})
		return
	}

	account, err := m.processor.AuthorizeStreamingRequest(accessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "could not authorize with given token"})
		return
	}

	upgrader := websocket.Upgrader{
		HandshakeTimeout: 5 * time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		Subprotocols:     []string{"wss"},
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		l.Infof("error upgrading websocket connection: %s", err)
		return
	}

	if errWithCode := m.processor.OpenStreamForAccount(conn, account, streamType); errWithCode != nil {
		c.JSON(errWithCode.Code(), errWithCode.Safe())
	}
}

package streaming

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func (m *Module) StreamGETHandler(c *gin.Context) {
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
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	if errWithCode := m.processor.StreamForAccount(conn, account, streamType); errWithCode != nil {
		c.JSON(errWithCode.Code(), errWithCode.Safe())
	}
}

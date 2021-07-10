package streaming

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// StreamGETHandler handles the creation of a new websocket streaming request.
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

	// make sure a valid token has been provided and obtain the associated account
	account, err := m.processor.AuthorizeStreamingRequest(accessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "could not authorize with given token"})
		return
	}

	// prepare to upgrade the connection to a websocket connection
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// we fully expect cors requests (via something like pinafore.social) so we should be lenient here
			return true
		},
	}

	// do the actual upgrade here
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		l.Infof("error upgrading websocket connection: %s", err)
		return
	}
	defer conn.Close() // whatever happens, when we leave this function we want to close the websocket connection

	// inform the processor that we have a new connection and want a stream for it
	stream, errWithCode := m.processor.OpenStreamForAccount(account, streamType)
	if errWithCode != nil {
		c.JSON(errWithCode.Code(), errWithCode.Safe())
		return
	}
	defer close(stream.Hangup) // closing stream.Hangup indicates that we've finished with the connection (the client has gone), so we want to do this on exiting this handler

	// spawn a new ticker for pinging the connection periodically
	t := time.NewTicker(30 * time.Second)

	// we want to stay in the sendloop as long as possible while the client is connected -- the only thing that should break the loop is if the client leaves or something else goes wrong
sendLoop:
	for {
		select {
		case m := <-stream.Messages:
			// we've got a streaming message!!
			l.Trace("received message from stream")
			if err := conn.WriteJSON(m); err != nil {
				l.Debugf("error writing json to websocket connection: %s", err)
				// if something is wrong we want to bail and drop the connection -- the client will create a new one
				break sendLoop
			}
			l.Trace("wrote message into websocket connection")
		case <-t.C:
			l.Trace("received TICK from ticker")
			if err := conn.WriteMessage(websocket.PingMessage, []byte(": ping")); err != nil {
				l.Debugf("error writing ping to websocket connection: %s", err)
				// if something is wrong we want to bail and drop the connection -- the client will create a new one
				break sendLoop
			}
			l.Trace("wrote ping message into websocket connection")
		}
	}

	l.Trace("leaving StreamGETHandler")
}

package streaming

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// StreamGETHandler swagger:operation GET /api/v1/streaming streamGet
//
// Initiate a websocket connection for live streaming of statuses and notifications.
//
// The scheme used should *always* be `wss`. The streaming basepath can be viewed at `/api/v1/instance`.
//
// On a successful connection, a code `101` will be returned, which indicates that the connection is being upgraded to a secure websocket connection.
//
// As long as the connection is open, various message types will be streamed into it.
//
// GoToSocial will ping the connection every 30 seconds to check whether the client is still receiving.
//
// If the ping fails, or something else goes wrong during transmission, then the connection will be dropped, and the client will be expected to start it again.
//
// ---
// tags:
// - streaming
//
// produces:
// - application/json
//
// schemes:
// - wss
//
// parameters:
// - name: access_token
//   type: string
//   description: Access token for the requesting account.
//   in: query
//   required: true
// - name: stream
//   type: string
//   description: |-
//     Type of stream to request.
//
//     Options are:
//
//     `user`: receive updates for the account's home timeline.
//     `public`: receive updates for the public timeline.
//     `public:local`: receive updates for the local timeline.
//     `hashtag`: receive updates for a given hashtag.
//     `hashtag:local`: receive local updates for a given hashtag.
//     `list`: receive updates for a certain list of accounts.
//     `direct`: receive updates for direct messages.
//   in: query
//   required: true
// security:
// - OAuth2 Bearer:
//   - read:streaming
//
// responses:
//   '101':
//     schema:
//       type: object
//       properties:
//         stream:
//           type: array
//           items:
//             type: string
//             enum:
//             - user
//             - public
//             - public:local
//             - hashtag
//             - hashtag:local
//             - list
//             - direct
//         event:
//           description: |-
//             The type of event being received.
//
//             `update`: a new status has been received.
//             `notification`: a new notification has been received.
//             `delete`: a status has been deleted.
//             `filters_changed`: not implemented.
//           type: string
//           enum:
//           - update
//           - notification
//           - delete
//           - filters_changed
//         payload:
//           description: |-
//             The payload of the streamed message.
//             Different depending on the `event` type.
//
//             If present, it should be parsed as a string.
//
//             If `event` = `update`, then the payload will be a JSON string of a status.
//             If `event` = `notification`, then the payload will be a JSON string of a notification.
//             If `event` = `delete`, then the payload will be a status ID.
//           type: string
//           example: "{\"id\":\"01FC3TZ5CFG6H65GCKCJRKA669\",\"created_at\":\"2021-08-02T16:25:52Z\",\"sensitive\":false,\"spoiler_text\":\"\",\"visibility\":\"public\",\"language\":\"en\",\"uri\":\"https://gts.superseriousbusiness.org/users/dumpsterqueer/statuses/01FC3TZ5CFG6H65GCKCJRKA669\",\"url\":\"https://gts.superseriousbusiness.org/@dumpsterqueer/statuses/01FC3TZ5CFG6H65GCKCJRKA669\",\"replies_count\":0,\"reblogs_count\":0,\"favourites_count\":0,\"favourited\":false,\"reblogged\":false,\"muted\":false,\"bookmarked\":fals…//gts.superseriousbusiness.org/fileserver/01JNN207W98SGG3CBJ76R5MVDN/header/original/019036W043D8FXPJKSKCX7G965.png\",\"header_static\":\"https://gts.superseriousbusiness.org/fileserver/01JNN207W98SGG3CBJ76R5MVDN/header/small/019036W043D8FXPJKSKCX7G965.png\",\"followers_count\":33,\"following_count\":28,\"statuses_count\":126,\"last_status_at\":\"2021-08-02T16:25:52Z\",\"emojis\":[],\"fields\":[]},\"media_attachments\":[],\"mentions\":[],\"tags\":[],\"emojis\":[],\"card\":null,\"poll\":null,\"text\":\"a\"}"
//   '401':
//      description: unauthorized
//   '400':
//      description: bad request
func (m *Module) StreamGETHandler(c *gin.Context) {
	l := logrus.WithField("func", "StreamGETHandler")

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
	account, err := m.processor.AuthorizeStreamingRequest(c.Request.Context(), accessToken)
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

	// inform the processor that we have a new connection and want a s for it
	s, errWithCode := m.processor.OpenStreamForAccount(c.Request.Context(), account, streamType)
	if errWithCode != nil {
		c.JSON(errWithCode.Code(), errWithCode.Safe())
		return
	}
	defer close(s.Hangup) // closing stream.Hangup indicates that we've finished with the connection (the client has gone), so we want to do this on exiting this handler

	// spawn a new ticker for pinging the connection periodically
	t := time.NewTicker(30 * time.Second)

	// we want to stay in the sendloop as long as possible while the client is connected -- the only thing that should break the loop is if the client leaves or something else goes wrong
sendLoop:
	for {
		select {
		case m := <-s.Messages:
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

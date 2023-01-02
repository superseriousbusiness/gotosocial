/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package streaming

import (
	"fmt"
	"net/http"
	"time"

	"codeberg.org/gruf/go-kv"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// we expect cors requests (via eg., pinafore.social) so be lenient
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	errNoToken = fmt.Errorf("no access token provided under query key %s or under header %s", AccessTokenQueryKey, AccessTokenHeader)
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
//	---
//	tags:
//	- streaming
//
//	produces:
//	- application/json
//
//	schemes:
//	- wss
//
//	parameters:
//	-
//		name: access_token
//		type: string
//		description: Access token for the requesting account.
//		in: query
//		required: true
//	-
//		name: stream
//		type: string
//		description: |-
//			Type of stream to request.
//
//			Options are:
//
//			`user`: receive updates for the account's home timeline.
//			`public`: receive updates for the public timeline.
//			`public:local`: receive updates for the local timeline.
//			`hashtag`: receive updates for a given hashtag.
//			`hashtag:local`: receive local updates for a given hashtag.
//			`list`: receive updates for a certain list of accounts.
//			`direct`: receive updates for direct messages.
//		in: query
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- read:streaming
//
//	responses:
//		'101':
//			schema:
//				type: object
//				properties:
//					stream:
//						type: array
//						items:
//							type: string
//							enum:
//							- user
//							- public
//							- public:local
//							- hashtag
//							- hashtag:local
//							- list
//							- direct
//					event:
//						description: |-
//							The type of event being received.
//
//							`update`: a new status has been received.
//							`notification`: a new notification has been received.
//							`delete`: a status has been deleted.
//							`filters_changed`: not implemented.
//						type: string
//						enum:
//						- update
//						- notification
//						- delete
//						- filters_changed
//					payload:
//						description: |-
//							The payload of the streamed message.
//							Different depending on the `event` type.
//
//							If present, it should be parsed as a string.
//
//							If `event` = `update`, then the payload will be a JSON string of a status.
//							If `event` = `notification`, then the payload will be a JSON string of a notification.
//							If `event` = `delete`, then the payload will be a status ID.
//						type: string
//						example: "{\"id\":\"01FC3TZ5CFG6H65GCKCJRKA669\",\"created_at\":\"2021-08-02T16:25:52Z\",\"sensitive\":false,\"spoiler_text\":\"\",\"visibility\":\"public\",\"language\":\"en\",\"uri\":\"https://gts.superseriousbusiness.org/users/dumpsterqueer/statuses/01FC3TZ5CFG6H65GCKCJRKA669\",\"url\":\"https://gts.superseriousbusiness.org/@dumpsterqueer/statuses/01FC3TZ5CFG6H65GCKCJRKA669\",\"replies_count\":0,\"reblogs_count\":0,\"favourites_count\":0,\"favourited\":false,\"reblogged\":false,\"muted\":false,\"bookmarked\":fals…//gts.superseriousbusiness.org/fileserver/01JNN207W98SGG3CBJ76R5MVDN/header/original/019036W043D8FXPJKSKCX7G965.png\",\"header_static\":\"https://gts.superseriousbusiness.org/fileserver/01JNN207W98SGG3CBJ76R5MVDN/header/small/019036W043D8FXPJKSKCX7G965.png\",\"followers_count\":33,\"following_count\":28,\"statuses_count\":126,\"last_status_at\":\"2021-08-02T16:25:52Z\",\"emojis\":[],\"fields\":[]},\"media_attachments\":[],\"mentions\":[],\"tags\":[],\"emojis\":[],\"card\":null,\"poll\":null,\"text\":\"a\"}"
//		'401':
//			description: unauthorized
//		'400':
//			description: bad request
func (m *Module) StreamGETHandler(c *gin.Context) {
	streamType := c.Query(StreamQueryKey)
	if streamType == "" {
		err := fmt.Errorf("no stream type provided under query key %s", StreamQueryKey)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	var accessToken string
	if t := c.Query(AccessTokenQueryKey); t != "" {
		// try query param first
		accessToken = t
	} else if t := c.GetHeader(AccessTokenHeader); t != "" {
		// fall back to Sec-Websocket-Protocol
		accessToken = t
	} else {
		// no token
		err := errNoToken
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	account, errWithCode := m.processor.AuthorizeStreamingRequest(c.Request.Context(), accessToken)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	stream, errWithCode := m.processor.OpenStreamForAccount(c.Request.Context(), account, streamType)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	l := log.WithFields(kv.Fields{
		{"account", account.Username},
		{"path", BasePath},
		{"streamID", stream.ID},
		{"streamType", streamType},
	}...)

	wsConn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// If the upgrade fails, then Upgrade replies to the client with an HTTP error response.
		// Because websocket issues are a pretty common source of headaches, we should also log
		// this at Error to make this plenty visible and help admins out a bit.
		l.Errorf("error upgrading websocket connection: %s", err)
		close(stream.Hangup)
		return
	}

	defer func() {
		// cleanup
		wsConn.Close()
		close(stream.Hangup)
	}()

	streamTicker := time.NewTicker(m.tickDuration)
	defer streamTicker.Stop()

	// We want to stay in the loop as long as possible while the client is connected.
	// The only thing that should break the loop is if the client leaves or the connection becomes unhealthy.
	//
	// If the loop does break, we expect the client to reattempt connection, so it's cheap to leave + try again
wsLoop:
	for {
		select {
		case m := <-stream.Messages:
			l.Trace("received message from stream")
			if err := wsConn.WriteJSON(m); err != nil {
				l.Debugf("error writing json to websocket connection; breaking off: %s", err)
				break wsLoop
			}
			l.Trace("wrote message into websocket connection")
		case <-streamTicker.C:
			l.Trace("received TICK from ticker")
			if err := wsConn.WriteMessage(websocket.PingMessage, []byte(": ping")); err != nil {
				l.Debugf("error writing ping to websocket connection; breaking off: %s", err)
				break wsLoop
			}
			l.Trace("wrote ping message into websocket connection")
		}
	}
}

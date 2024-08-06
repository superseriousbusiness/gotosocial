// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package streaming

import (
	"context"
	"net/http"
	"slices"
	"time"

	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	streampkg "github.com/superseriousbusiness/gotosocial/internal/stream"

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
//	-
//		name: list
//		type: string
//		description: |-
//			ID of the list to subscribe to.
//			Only used if stream type is 'list'.
//		in: query
//	-
//		name: tag
//		type: string
//		description: |-
//			Name of the tag to subscribe to.
//			Only used if stream type is 'hashtag' or 'hashtag:local'.
//		in: query
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
//							`filters_changed`: filters (including keywords and statuses) have changed.
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
//							If `event` = `filters_changed`, then there is no payload.
//						type: string
//						example: "{\"id\":\"01FC3TZ5CFG6H65GCKCJRKA669\",\"created_at\":\"2021-08-02T16:25:52Z\",\"sensitive\":false,\"spoiler_text\":\"\",\"visibility\":\"public\",\"language\":\"en\",\"uri\":\"https://gts.superseriousbusiness.org/users/dumpsterqueer/statuses/01FC3TZ5CFG6H65GCKCJRKA669\",\"url\":\"https://gts.superseriousbusiness.org/@dumpsterqueer/statuses/01FC3TZ5CFG6H65GCKCJRKA669\",\"replies_count\":0,\"reblogs_count\":0,\"favourites_count\":0,\"favourited\":false,\"reblogged\":false,\"muted\":false,\"bookmarked\":fals…//gts.superseriousbusiness.org/fileserver/01JNN207W98SGG3CBJ76R5MVDN/header/original/019036W043D8FXPJKSKCX7G965.png\",\"header_static\":\"https://gts.superseriousbusiness.org/fileserver/01JNN207W98SGG3CBJ76R5MVDN/header/small/019036W043D8FXPJKSKCX7G965.png\",\"followers_count\":33,\"following_count\":28,\"statuses_count\":126,\"last_status_at\":\"2021-08-02T16:25:52Z\",\"emojis\":[],\"fields\":[]},\"media_attachments\":[],\"mentions\":[],\"tags\":[],\"emojis\":[],\"card\":null,\"poll\":null,\"text\":\"a\"}"
//		'401':
//			description: unauthorized
//		'400':
//			description: bad request
func (m *Module) StreamGETHandler(c *gin.Context) {
	var (
		token         string
		tokenInHeader bool
		account       *gtsmodel.Account
		errWithCode   gtserror.WithCode
	)

	if t := c.Query(AccessTokenQueryKey); t != "" {
		// Token was provided as
		// query param, no problem.
		token = t
	} else if t := c.GetHeader(AccessTokenHeader); t != "" {
		// Token was provided in "Sec-Websocket-Protocol" header.
		//
		// This is hacky and not technically correct but some
		// clients do it since Mastodon allows it, so we must
		// also allow it to avoid breaking expectations.
		token = t
		tokenInHeader = true
	}

	if token != "" {

		// Token was provided, use it to authorize stream.
		account, errWithCode = m.processor.Stream().Authorize(c.Request.Context(), token)
		if errWithCode != nil {
			apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
			return
		}

	} else {

		// No explicit token was provided:
		// try regular oauth as a last resort.
		authed, err := oauth.Authed(c, true, true, true, true)
		if err != nil {
			errWithCode := gtserror.NewErrorUnauthorized(err, err.Error())
			apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
			return
		}

		// Set the auth'ed account.
		account = authed.Account
	}

	if account.IsMoving() {
		// Moving accounts can't
		// use streaming endpoints.
		apiutil.NotFoundAfterMove(c)
		return
	}

	// Get the initial requested stream type, if there is one.
	streamType := c.Query(StreamQueryKey)

	// By appending other query params to the streamType, we
	// can allow streaming for specific list IDs or hashtags.
	// The streamType in this case will end up looking like
	// `hashtag:example` or `list:01H3YF48G8B7KTPQFS8D2QBVG8`.
	if list := c.Query(StreamListKey); list != "" {
		streamType += ":" + list
	} else if tag := c.Query(StreamTagKey); tag != "" {
		streamType += ":" + tag
	}

	// Open a stream with the processor; this lets processor
	// functions pass messages into a channel, which we can
	// then read from and put into a websockets connection.
	stream, errWithCode := m.processor.Stream().Open(
		c.Request.Context(), // this ctx is only used for logging
		account,
		streamType,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	l := log.
		WithContext(c.Request.Context()).
		WithField("streamID", id.NewULID()).
		WithField("username", account.Username)

	// Upgrade the incoming HTTP request. This hijacks the
	// underlying connection and reuses it for the websocket
	// (non-http) protocol.
	//
	// If the upgrade fails, then Upgrade replies to the client
	// with an HTTP error response.
	var responseHeader http.Header
	if tokenInHeader {
		// Return the token in the response,
		// else Chrome fails to connect.
		//
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Protocol_upgrade_mechanism#sec-websocket-protocol
		responseHeader = http.Header{AccessTokenHeader: {token}}
	}

	wsConn, err := m.wsUpgrade.Upgrade(c.Writer, c.Request, responseHeader)
	if err != nil {
		l.Errorf("error upgrading websocket connection: %v", err)
		stream.Close()
		return
	}

	// We perform the main websocket rw loops in a separate
	// goroutine in order to let the upgrade handler return.
	// This prevents the upgrade handler from holding open any
	// throttle / rate-limit request tokens which could become
	// problematic on instances with multiple users.
	go m.handleWSConn(&l, wsConn, stream)
}

// handleWSConn handles a two-way websocket streaming connection.
// It will both read messages from the connection, and push messages
// into the connection. If any errors are encountered while reading
// or writing (including expected errors like clients leaving), the
// connection will be closed.
func (m *Module) handleWSConn(l *log.Entry, wsConn *websocket.Conn, stream *streampkg.Stream) {
	l.Info("opened websocket connection")

	// Create new async context with cancel.
	ctx, cncl := context.WithCancel(context.Background())

	go func() {
		defer cncl()

		// Read messages from websocket to server.
		m.readFromWSConn(ctx, wsConn, stream, l)
	}()

	go func() {
		defer cncl()

		// Write messages from processor in websocket conn.
		m.writeToWSConn(ctx, wsConn, stream, m.dTicker, l)
	}()

	// Wait for ctx
	// to be closed.
	<-ctx.Done()

	// Close stream
	// straightaway.
	stream.Close()

	// Tidy up underlying websocket connection.
	if err := wsConn.Close(); err != nil {
		l.Errorf("error closing websocket connection: %v", err)
	}

	l.Info("closed websocket connection")
}

// readFromWSConn reads control messages coming in from the given
// websockets connection, and modifies the subscription StreamTypes
// of the given stream accordingly after acquiring a lock on it.
//
// This is a blocking function; will return only on read error or
// if the given context is canceled.
func (m *Module) readFromWSConn(
	ctx context.Context,
	wsConn *websocket.Conn,
	stream *streampkg.Stream,
	l *log.Entry,
) {

	for {
		var msg struct {
			Type   string `json:"type"`
			Stream string `json:"stream"`
			List   string `json:"list,omitempty"`
		}

		// Read JSON objects from the client and act on them.
		if err := wsConn.ReadJSON(&msg); err != nil {
			// Only log an error if something weird happened.
			// See: https://www.rfc-editor.org/rfc/rfc6455.html#section-11.7
			if !websocket.IsCloseError(err, []int{
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived,
			}...) {
				l.Errorf("error during websocket read: %v", err)
			}

			// The connection is gone; no
			// further streaming possible.
			break
		}

		// Messages *from* the WS connection are infrequent
		// and usually interesting, so log this at info.
		l.Infof("received websocket message: %+v", msg)

		// Ignore if the updateStreamType is unknown (or missing),
		// so a bad client can't cause extra memory allocations
		if !slices.Contains(streampkg.AllStatusTimelines, msg.Stream) {
			l.Warnf("unknown 'stream' field: %v", msg)
			continue
		}

		if msg.List != "" {
			// If a list is given, add this to
			// the stream name as this is how we
			// we track stream types internally.
			msg.Stream += ":" + msg.List
		}

		switch msg.Type {
		case "subscribe":
			stream.Subscribe(msg.Stream)
		case "unsubscribe":
			stream.Unsubscribe(msg.Stream)
		default:
			l.Warnf("invalid 'type' field: %v", msg)
		}
	}

	l.Debug("finished websocket read")
}

// writeToWSConn receives messages coming from the processor via the
// given stream, and writes them into the given websockets connection.
// This function also handles sending ping messages into the websockets
// connection to keep it alive when no other activity occurs.
//
// This is a blocking function; will return only on write error or
// if the given context is canceled.
func (m *Module) writeToWSConn(
	ctx context.Context,
	wsConn *websocket.Conn,
	stream *streampkg.Stream,
	ping time.Duration,
	l *log.Entry,
) {
	for {
		// Wrap context with timeout to send a ping.
		pingctx, cncl := context.WithTimeout(ctx, ping)

		// Block on receipt of msg.
		msg, ok := stream.Recv(pingctx)

		// Check if cancel because ping.
		pinged := (pingctx.Err() != nil)
		cncl()

		switch {
		case !ok && pinged:
			// The ping context timed out!
			l.Trace("writing websocket ping")

			// Wrapped context time-out, send a keep-alive "ping".
			if err := wsConn.WriteControl(websocket.PingMessage, nil, time.Time{}); err != nil {
				l.Debugf("error writing websocket ping: %v", err)
				break
			}

		case !ok:
			// Stream was
			// closed.
			return
		}

		l.Trace("writing websocket message: %+v", msg)

		// Received a new message from the processor.
		if err := wsConn.WriteJSON(msg); err != nil {
			l.Debugf("error writing websocket message: %v", err)
			break
		}
	}

	l.Debug("finished websocket write")
}

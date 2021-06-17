package streaming

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

func (p *processor) OpenStreamForAccount(conn *websocket.Conn, account *gtsmodel.Account, streamType string) gtserror.WithCode {
	l := p.log.WithFields(logrus.Fields{
		"func":       "OpenStreamForAccount",
		"account":    account.ID,
		"streamType": streamType,
	})
	l.Debug("received open stream request")

	streamID, err := id.NewRandomULID()
	if err != nil {
		return gtserror.NewErrorInternalError(fmt.Errorf("error generating stream id: %s", err))
	}

	thisStream := &stream{
		streamID:   streamID,
		streamType: streamType,
		conn:       conn,
	}

	v, ok := p.streamMap.Load(account.ID)
	if !ok || v == nil {
		// there is no entry in the streamMap for this account yet, so make one and store it
		streams := &streamsForAccount{
			s: []*stream{
				thisStream,
			},
		}
		p.streamMap.Store(account.ID, streams)
	} else {
		// there is an entry in the streamMap for this account
		// parse the interface as a streamsForAccount
		streams, ok := v.(*streamsForAccount)
		if !ok {
			return gtserror.NewErrorInternalError(errors.New("stream map error"))
		}

		// append this stream to it
		streams.Lock()
		streams.s = append(streams.s, thisStream)
		streams.Unlock()
	}

	// set the close handler to remove the given stream from the stream map so that messages stop getting put into it
	conn.SetCloseHandler(func(code int, text string) error {
		l.Debug("closing stream")
		v, ok := p.streamMap.Load(account.ID)
		if !ok || v == nil {
			// the map doesn't contain an entry for the account anyway, so we can just return
			// this probably should never happen but let's check anyway
			return nil
		}

		// parse the interface as a streamsForAccount
		streams, ok := v.(*streamsForAccount)
		if !ok {
			return gtserror.NewErrorInternalError(errors.New("stream map error"))
		}

		// remove thisStream from the slice of streams stored in streamsForAccount
		streams.Lock()
		newStreamSlice := []*stream{}
		for _, s := range streams.s {
			if s.streamID != thisStream.streamID {
				newStreamSlice = append(newStreamSlice, s)
			}
		}
		streams.s = newStreamSlice
		streams.Unlock()
		l.Debug("stream closed")
		return nil
	})

	defer conn.Close()
	t := time.NewTicker(60 * time.Second)
	for range t.C {
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return gtserror.NewErrorInternalError(err)
		}
	}

	return nil
}

type streamsForAccount struct {
	s []*stream
	sync.Mutex
}

type stream struct {
	streamID   string
	streamType string
	conn       *websocket.Conn
}

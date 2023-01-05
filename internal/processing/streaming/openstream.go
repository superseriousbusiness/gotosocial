/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"context"
	"errors"
	"fmt"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
)

func (p *processor) OpenStreamForAccount(ctx context.Context, account *gtsmodel.Account, streamTimeline string) (*stream.Stream, gtserror.WithCode) {
	l := log.WithFields(kv.Fields{

		{"account", account.ID},
		{"streamType", streamTimeline},
	}...)
	l.Debug("received open stream request")

	// each stream needs a unique ID so we know to close it
	streamID, err := id.NewRandomULID()
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error generating stream id: %s", err))
	}

	thisStream := &stream.Stream{
		ID:        streamID,
		Timeline:  streamTimeline,
		Messages:  make(chan *stream.Message, 100),
		Hangup:    make(chan interface{}, 1),
		Connected: true,
	}
	go p.waitToCloseStream(account, thisStream)

	v, ok := p.streamMap.Load(account.ID)
	if !ok || v == nil {
		// there is no entry in the streamMap for this account yet, so make one and store it
		streamsForAccount := &stream.StreamsForAccount{
			Streams: []*stream.Stream{
				thisStream,
			},
		}
		p.streamMap.Store(account.ID, streamsForAccount)
	} else {
		// there is an entry in the streamMap for this account
		// parse the interface as a streamsForAccount
		streamsForAccount, ok := v.(*stream.StreamsForAccount)
		if !ok {
			return nil, gtserror.NewErrorInternalError(errors.New("stream map error"))
		}

		// append this stream to it
		streamsForAccount.Lock()
		streamsForAccount.Streams = append(streamsForAccount.Streams, thisStream)
		streamsForAccount.Unlock()
	}

	return thisStream, nil
}

// waitToCloseStream waits until the hangup channel is closed for the given stream.
// It then iterates through the map of streams stored by the processor, removes the stream from it,
// and then closes the messages channel of the stream to indicate that the channel should no longer be read from.
func (p *processor) waitToCloseStream(account *gtsmodel.Account, thisStream *stream.Stream) {
	<-thisStream.Hangup // wait for a hangup message

	// lock the stream to prevent more messages being put in it while we work
	thisStream.Lock()
	defer thisStream.Unlock()

	// indicate the stream is no longer connected
	thisStream.Connected = false

	// load and parse the entry for this account from the stream map
	v, ok := p.streamMap.Load(account.ID)
	if !ok || v == nil {
		return
	}
	streamsForAccount, ok := v.(*stream.StreamsForAccount)
	if !ok {
		return
	}

	// lock the streams for account while we remove this stream from its slice
	streamsForAccount.Lock()
	defer streamsForAccount.Unlock()

	// put everything into modified streams *except* the stream we're removing
	modifiedStreams := []*stream.Stream{}
	for _, s := range streamsForAccount.Streams {
		if s.ID != thisStream.ID {
			modifiedStreams = append(modifiedStreams, s)
		}
	}
	streamsForAccount.Streams = modifiedStreams

	// finally close the messages channel so no more messages can be read from it
	close(thisStream.Messages)
}

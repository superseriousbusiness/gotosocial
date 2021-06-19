package streaming

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

func (p *processor) OpenStreamForAccount(account *gtsmodel.Account, streamType string) (*gtsmodel.Stream, gtserror.WithCode) {
	l := p.log.WithFields(logrus.Fields{
		"func":       "OpenStreamForAccount",
		"account":    account.ID,
		"streamType": streamType,
	})
	l.Debug("received open stream request")

	// each stream needs a unique ID so we know to close it
	streamID, err := id.NewRandomULID()
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error generating stream id: %s", err))
	}

	thisStream := &gtsmodel.Stream{
		ID:        streamID,
		Type:      streamType,
		Messages:  make(chan *gtsmodel.Message, 100),
		Hangup:    make(chan interface{}, 1),
		Connected: true,
	}
	go p.waitToCloseStream(account, thisStream)

	v, ok := p.streamMap.Load(account.ID)
	if !ok || v == nil {
		// there is no entry in the streamMap for this account yet, so make one and store it
		streamsForAccount := &gtsmodel.StreamsForAccount{
			Streams: []*gtsmodel.Stream{
				thisStream,
			},
		}
		p.streamMap.Store(account.ID, streamsForAccount)
	} else {
		// there is an entry in the streamMap for this account
		// parse the interface as a streamsForAccount
		streamsForAccount, ok := v.(*gtsmodel.StreamsForAccount)
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
func (p *processor) waitToCloseStream(account *gtsmodel.Account, thisStream *gtsmodel.Stream) {
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
	streamsForAccount, ok := v.(*gtsmodel.StreamsForAccount)
	if !ok {
		return
	}

	// lock the streams for account while we remove this stream from its slice
	streamsForAccount.Lock()
	defer streamsForAccount.Unlock()

	// put everything into modified streams *except* the stream we're removing
	modifiedStreams := []*gtsmodel.Stream{}
	for _, s := range streamsForAccount.Streams {
		if s.ID != thisStream.ID {
			modifiedStreams = append(modifiedStreams, s)
		}
	}
	streamsForAccount.Streams = modifiedStreams

	// finally close the messages channel so no more messages can be read from it
	close(thisStream.Messages)
}

package streaming

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) StreamStatusToAccount(s *apimodel.Status, account *gtsmodel.Account) error {
	l := p.log.WithFields(logrus.Fields{
		"func":    "StreamStatusForAccount",
		"account": account.ID,
	})
	v, ok := p.streamMap.Load(account.ID)
	if !ok {
		// no open connections so nothing to stream
		return nil
	}

	streamsForAccount, ok := v.(*gtsmodel.StreamsForAccount)
	if !ok {
		return errors.New("stream map error")
	}

	statusBytes, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("error marshalling status to json: %s", err)
	}

	streamsForAccount.Lock()
	defer streamsForAccount.Unlock()
	for _, stream := range streamsForAccount.Streams {
		stream.Lock()
		defer stream.Unlock()
		if stream.Connected {
			l.Debugf("streaming status to stream id %s", stream.ID)
			stream.Messages <- &gtsmodel.Message{
				Stream:  []string{stream.Type},
				Event:   "update",
				Payload: string(statusBytes),
			}
		}
	}

	return nil
}

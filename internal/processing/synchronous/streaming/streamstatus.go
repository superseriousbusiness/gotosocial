package streaming

import (
	"errors"

	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) StreamStatusForAccount(s *apimodel.Status, account *gtsmodel.Account) error {
	l := p.log.WithFields(logrus.Fields{
		"func":    "StreamStatusForAccount",
		"account": account.ID,
	})
	v, ok := p.streamMap.Load(account.ID)
	if !ok {
		// no open connections so nothing to stream
		return nil
	}

	streams, ok := v.(*streamsForAccount)
	if !ok {
		return errors.New("stream map error")
	}

	streams.Lock()
	defer streams.Unlock()
	for _, stream := range streams.s {
		l.Debugf("streaming status to stream id %s", stream.streamID)
		if err := stream.conn.WriteJSON(s); err != nil {
			return err
		}
	}

	return nil
}

package streaming

import (
	"fmt"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/stream"
)

func (p *processor) StreamDelete(statusID string) error {
	errs := []string{}

	// we want to range through ALL streams for ALL accounts here to make sure it's very clear to everyone that the status has been deleted
	p.streamMap.Range(func(k interface{}, v interface{}) bool {
		// the key of this map should be an accountID (string)
		accountID, ok := k.(string)
		if !ok {
			errs = append(errs, "key in streamMap was not a string!")
			return false
		}

		// the value of the map should be a buncha streams
		streamsForAccount, ok := v.(*stream.StreamsForAccount)
		if !ok {
			errs = append(errs, fmt.Sprintf("stream map error for account stream %s", accountID))
		}

		// lock the streams while we work on them
		streamsForAccount.Lock()
		defer streamsForAccount.Unlock()
		for _, s := range streamsForAccount.Streams {
			// lock each individual stream as we work on it
			s.Lock()
			defer s.Unlock()
			if s.Connected {
				s.Messages <- &stream.Message{
					Stream:  []string{s.Type},
					Event:   "delete",
					Payload: statusID,
				}
			}
		}
		return true
	})

	if len(errs) != 0 {
		return fmt.Errorf("one or more errors streaming status delete: %s", strings.Join(errs, ";"))
	}

	return nil
}

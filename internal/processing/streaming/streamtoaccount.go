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
	"errors"

	"github.com/superseriousbusiness/gotosocial/internal/stream"
)

// streamToAccount streams the given payload with the given event type to any streams currently open for the given account ID.
func (p *processor) streamToAccount(payload string, event string, timelines []string, accountID string) error {
	v, ok := p.streamMap.Load(accountID)
	if !ok {
		// no open connections so nothing to stream
		return nil
	}

	streamsForAccount, ok := v.(*stream.StreamsForAccount)
	if !ok {
		return errors.New("stream map error")
	}

	streamsForAccount.Lock()
	defer streamsForAccount.Unlock()
	for _, s := range streamsForAccount.Streams {
		s.Lock()
		defer s.Unlock()
		if !s.Connected {
			continue
		}

		for _, t := range timelines {
			if s.Timeline == string(t) {
				s.Messages <- &stream.Message{
					Stream:  []string{string(t)},
					Event:   string(event),
					Payload: payload,
				}
			}
		}
	}

	return nil
}

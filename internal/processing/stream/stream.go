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

package stream

import (
	"sync"

	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
)

type Processor struct {
	state       *state.State
	oauthServer oauth.Server
	streamMap   *sync.Map
}

func New(state *state.State, oauthServer oauth.Server) Processor {
	return Processor{
		state:       state,
		oauthServer: oauthServer,
		streamMap:   &sync.Map{},
	}
}

// toAccount streams the given payload with the given event type to any streams currently open for the given account ID.
func (p *Processor) toAccount(payload string, event string, streamTypes []string, accountID string) error {
	// Load all streams open for this account.
	v, ok := p.streamMap.Load(accountID)
	if !ok {
		return nil // No entry = nothing to stream.
	}
	streamsForAccount := v.(*stream.StreamsForAccount) //nolint:forcetypeassert

	streamsForAccount.Lock()
	defer streamsForAccount.Unlock()

	for _, s := range streamsForAccount.Streams {
		s.Lock()
		defer s.Unlock()

		if !s.Connected {
			continue
		}

	typeLoop:
		for _, streamType := range streamTypes {
			if _, found := s.StreamTypes[streamType]; found {
				s.Messages <- &stream.Message{
					Stream:  []string{streamType},
					Event:   string(event),
					Payload: payload,
				}

				// Break out to the outer loop,
				// to avoid sending duplicates of
				// the same event to the same stream.
				break typeLoop
			}
		}
	}

	return nil
}

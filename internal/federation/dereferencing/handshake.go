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

package dereferencing

import (
	"net/url"
)

func (d *Dereferencer) Handshaking(username string, remoteAccountID *url.URL) bool {
	d.handshakesMu.Lock()
	defer d.handshakesMu.Unlock()

	if d.handshakes == nil {
		// Handshakes isn't even initialized yet,
		// so we can't be handshaking with anyone.
		return false
	}

	remoteIDs, ok := d.handshakes[username]
	if !ok {
		// Given username isn't
		// handshaking with anyone.
		return false
	}

	// Calculate remote account ID str once.
	remoteIDStr := remoteAccountID.String()

	for _, id := range remoteIDs {
		if id.String() == remoteIDStr {
			// We are currently handshaking
			// with the remote account.
			return true
		}
	}

	// No results: we're not handshaking
	// with the remote account.
	return false
}

func (d *Dereferencer) startHandshake(username string, remoteAccountID *url.URL) {
	d.handshakesMu.Lock()
	defer d.handshakesMu.Unlock()

	remoteIDs, ok := d.handshakes[username]
	if !ok {
		// No handshakes were stored yet,
		// so just add this entry and return.
		d.handshakes[username] = []*url.URL{remoteAccountID}
		return
	}

	// Add the remote account ID to the slice.
	remoteIDs = append(remoteIDs, remoteAccountID)
	d.handshakes[username] = remoteIDs
}

func (d *Dereferencer) stopHandshake(username string, remoteAccountID *url.URL) {
	d.handshakesMu.Lock()
	defer d.handshakesMu.Unlock()

	remoteIDs, ok := d.handshakes[username]
	if !ok {
		// No handshake was in progress,
		// so there's nothing to stop.
		return
	}

	// Generate a new remoteIDs slice that
	// doesn't contain the removed entry.
	var (
		remoteAccountIDStr = remoteAccountID.String()
		newRemoteIDs       = make([]*url.URL, 0, len(remoteIDs)-1)
	)

	for _, id := range remoteIDs {
		if id.String() != remoteAccountIDStr {
			newRemoteIDs = append(newRemoteIDs, id)
		}
	}

	if len(newRemoteIDs) == 0 {
		// There are no handshakes remaining,
		// so just remove this username's slice
		// from the map and save a few bytes.
		delete(d.handshakes, username)
	} else {
		// There are still other handshakes ongoing.
		d.handshakes[username] = newRemoteIDs
	}
}

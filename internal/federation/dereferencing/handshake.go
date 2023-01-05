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

package dereferencing

import (
	"context"
	"net/url"
)

func (d *deref) Handshaking(ctx context.Context, username string, remoteAccountID *url.URL) bool {
	d.handshakeSync.Lock()
	defer d.handshakeSync.Unlock()

	if d.handshakes == nil {
		// handshakes isn't even initialized yet so we can't be handshaking with anyone
		return false
	}

	remoteIDs, ok := d.handshakes[username]
	if !ok {
		// user isn't handshaking with anyone, bail
		return false
	}

	for _, id := range remoteIDs {
		if id.String() == remoteAccountID.String() {
			// we are currently handshaking with the remote account, yep
			return true
		}
	}

	// didn't find it which means we're not handshaking
	return false
}

func (d *deref) startHandshake(username string, remoteAccountID *url.URL) {
	d.handshakeSync.Lock()
	defer d.handshakeSync.Unlock()

	// lazily initialize handshakes
	if d.handshakes == nil {
		d.handshakes = make(map[string][]*url.URL)
	}

	remoteIDs, ok := d.handshakes[username]
	if !ok {
		// there was nothing in there yet, so just add this entry and return
		d.handshakes[username] = []*url.URL{remoteAccountID}
		return
	}

	// add the remote ID to the slice
	remoteIDs = append(remoteIDs, remoteAccountID)
	d.handshakes[username] = remoteIDs
}

func (d *deref) stopHandshake(username string, remoteAccountID *url.URL) {
	d.handshakeSync.Lock()
	defer d.handshakeSync.Unlock()

	if d.handshakes == nil {
		return
	}

	remoteIDs, ok := d.handshakes[username]
	if !ok {
		// there was nothing in there yet anyway so just bail
		return
	}

	newRemoteIDs := []*url.URL{}
	for _, id := range remoteIDs {
		if id.String() != remoteAccountID.String() {
			newRemoteIDs = append(newRemoteIDs, id)
		}
	}

	if len(newRemoteIDs) == 0 {
		// there are no handshakes so just remove this user entry from the map and save a few bytes
		delete(d.handshakes, username)
	} else {
		// there are still other handshakes ongoing
		d.handshakes[username] = newRemoteIDs
	}
}

package dereferencing

import "net/url"

func (d *deref) Handshaking(username string, remoteAccountID *url.URL) bool {
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

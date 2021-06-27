package federation

import "net/url"

func (f *federator) Handshaking(username string, remoteAccountID *url.URL) bool {
	f.handshakeSync.Lock()
	defer f.handshakeSync.Unlock()

	if f.handshakes == nil {
		// handshakes isn't even initialized yet so we can't be handshaking with anyone
		return false
	}

	remoteIDs, ok := f.handshakes[username];
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

func (f *federator) startHandshake(username string, remoteAccountID *url.URL) {
	f.handshakeSync.Lock()
	defer f.handshakeSync.Unlock()

	// lazily initialize handshakes
	if f.handshakes == nil {
		f.handshakes = make(map[string][]*url.URL)
	}

	remoteIDs, ok := f.handshakes[username]
	if !ok {
		// there was nothing in there yet, so just add this entry and return
		f.handshakes[username] = []*url.URL{remoteAccountID}
		return
	}

	// add the remote ID to the slice
	remoteIDs = append(remoteIDs, remoteAccountID)
	f.handshakes[username] = remoteIDs
}

func (f *federator) stopHandshake(username string, remoteAccountID *url.URL) {
	f.handshakeSync.Lock()
	defer f.handshakeSync.Unlock()

	if f.handshakes == nil {
		return
	}

	remoteIDs, ok := f.handshakes[username]
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
		delete(f.handshakes, username)
	} else {
		// there are still other handshakes ongoing
		f.handshakes[username] = newRemoteIDs
	}
}

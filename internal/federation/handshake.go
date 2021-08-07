package federation

import "net/url"

func (f *federator) Handshaking(username string, remoteAccountID *url.URL) bool {
	return f.dereferencer.Handshaking(username, remoteAccountID)
}

package federation

import (
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *federator) GetRemoteAccount(username string, remoteAccountID *url.URL, refresh bool) (*gtsmodel.Account, bool, error) {
	return f.dereferencer.GetRemoteAccount(username, remoteAccountID, refresh)
}

func (f *federator) GetRemoteStatus(username string, remoteStatusID *url.URL) (*gtsmodel.Status, bool, error) {
	return f.dereferencer.GetRemoteStatus(username, remoteStatusID)
}

func (f *federator) DereferenceRemoteThread(username string, statusIRI *url.URL) error {
	return f.dereferencer.DereferenceThread(username, statusIRI)
}

func (f *federator) GetRemoteInstance(username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error) {
	return f.dereferencer.GetRemoteInstance(username, remoteInstanceURI)
}

func (f *federator) DereferenceAnnounce(announce *gtsmodel.Status, requestingUsername string) error {
	return f.dereferencer.DereferenceAnnounce(announce, requestingUsername)
}

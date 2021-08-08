package federation

import (
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *federator) GetRemoteAccount(username string, remoteAccountID *url.URL, refresh bool) (*gtsmodel.Account, bool, error) {
	return f.dereferencer.GetRemoteAccount(username, remoteAccountID, refresh)
}

func (f *federator) GetRemoteStatus(username string, remoteStatusID *url.URL, refresh bool) (*gtsmodel.Status, ap.Statusable, bool, error) {
	return f.dereferencer.GetRemoteStatus(username, remoteStatusID, refresh)
}

func (f *federator) EnrichRemoteStatus(username string, status *gtsmodel.Status) (*gtsmodel.Status, error) {
	return f.dereferencer.EnrichRemoteStatus(username, status)
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

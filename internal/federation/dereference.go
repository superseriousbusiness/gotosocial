package federation

import (
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

func (f *federator) DereferenceRemoteThread(username string, statusIRI *url.URL) error {
	return f.dereferencer.DereferenceThread(username, statusIRI)
}

func (f *federator) DereferenceCollectionPage(username string, pageIRI *url.URL) (typeutils.CollectionPageable, error) {
	return f.dereferencer.DereferenceCollectionPage(username, pageIRI)
}

func (f *federator) DereferenceRemoteAccount(username string, remoteAccountID *url.URL) (typeutils.Accountable, error) {
	f.startHandshake(username, remoteAccountID)
	defer f.stopHandshake(username, remoteAccountID)

	return f.dereferencer.DereferenceAccountable(username, remoteAccountID)
}

func (f *federator) DereferenceRemoteStatus(username string, remoteStatusID *url.URL) (typeutils.Statusable, error) {
	return f.dereferencer.DereferenceStatusable(username, remoteStatusID)
}

func (f *federator) DereferenceRemoteInstance(username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error) {
	return f.dereferencer.DereferenceRemoteInstance(username, remoteInstanceURI)
}

func (f *federator) DereferenceStatusFields(status *gtsmodel.Status, requestingUsername string) error {
	return f.dereferencer.PopulateStatusFields(status, requestingUsername)
}

func (f *federator) DereferenceAccountFields(account *gtsmodel.Account, requestingUsername string, refresh bool) error {
	return f.dereferencer.PopulateAccountFields(account, requestingUsername, refresh)
}

func (f *federator) DereferenceAnnounce(announce *gtsmodel.Status, requestingUsername string) error {
	return f.dereferencer.DereferenceAnnounce(announce, requestingUsername)
}

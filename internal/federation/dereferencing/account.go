package dereferencing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

func (d *deref) DereferenceAccountable(username string, remoteAccountID *url.URL) (typeutils.Accountable, error) {
	if blocked, err := d.blockedDomain(remoteAccountID.Host); blocked || err != nil {
		return nil, fmt.Errorf("DereferenceAccountable: domain %s is blocked", remoteAccountID.Host)
	}

	transport, err := d.transportController.NewTransportForUsername(username)
	if err != nil {
		return nil, fmt.Errorf("DereferenceAccountable: transport err: %s", err)
	}

	b, err := transport.Dereference(context.Background(), remoteAccountID)
	if err != nil {
		return nil, fmt.Errorf("DereferenceAccountable: error deferencing %s: %s", remoteAccountID.String(), err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("DereferenceAccountable: error unmarshalling bytes into json: %s", err)
	}

	t, err := streams.ToType(context.Background(), m)
	if err != nil {
		return nil, fmt.Errorf("DereferenceAccountable: error resolving json into ap vocab type: %s", err)
	}

	switch t.GetTypeName() {
	case string(gtsmodel.ActivityStreamsPerson):
		p, ok := t.(vocab.ActivityStreamsPerson)
		if !ok {
			return nil, errors.New("DereferenceAccountable: error resolving type as activitystreams person")
		}
		return p, nil
	case string(gtsmodel.ActivityStreamsApplication):
		p, ok := t.(vocab.ActivityStreamsApplication)
		if !ok {
			return nil, errors.New("DereferenceAccountable: error resolving type as activitystreams application")
		}
		return p, nil
	case string(gtsmodel.ActivityStreamsService):
		p, ok := t.(vocab.ActivityStreamsService)
		if !ok {
			return nil, errors.New("DereferenceAccountable: error resolving type as activitystreams service")
		}
		return p, nil
	}

	return nil, fmt.Errorf("DereferenceAccountable: type name %s not supported", t.GetTypeName())
}

func (d *deref) PopulateAccountFields(account *gtsmodel.Account, requestingUsername string, refresh bool) error {
	l := d.log.WithFields(logrus.Fields{
		"func":               "PopulateAccountFields",
		"requestingUsername": requestingUsername,
	})

	accountURI, err := url.Parse(account.URI)
	if err != nil {
		return fmt.Errorf("PopulateAccountFields: couldn't parse account URI %s: %s", account.URI, err)
	}
	if blocked, err := d.blockedDomain(accountURI.Host); blocked || err != nil {
		return fmt.Errorf("PopulateAccountFields: domain %s is blocked", accountURI.Host)
	}

	t, err := d.transportController.NewTransportForUsername(requestingUsername)
	if err != nil {
		return fmt.Errorf("PopulateAccountFields: error getting transport for user: %s", err)
	}

	// fetch the header and avatar
	if err := d.fetchHeaderAndAviForAccount(account, t, refresh); err != nil {
		// if this doesn't work, just skip it -- we can do it later
		l.Debugf("error fetching header/avi for account: %s", err)
	}

	if err := d.db.UpdateByID(account.ID, account); err != nil {
		return fmt.Errorf("PopulateAccountFields: error updating account in database: %s", err)
	}

	return nil
}

// fetchHeaderAndAviForAccount fetches the header and avatar for a remote account, using a transport
// on behalf of requestingUsername.
//
// targetAccount's AvatarMediaAttachmentID and HeaderMediaAttachmentID will be updated as necessary.
//
// SIDE EFFECTS: remote header and avatar will be stored in local storage, and the database will be updated
// to reflect the creation of these new attachments.
func (d *deref) fetchHeaderAndAviForAccount(targetAccount *gtsmodel.Account, t transport.Transport, refresh bool) error {
	accountURI, err := url.Parse(targetAccount.URI)
	if err != nil {
		return fmt.Errorf("fetchHeaderAndAviForAccount: couldn't parse account URI %s: %s", targetAccount.URI, err)
	}
	if blocked, err := d.blockedDomain(accountURI.Host); blocked || err != nil {
		return fmt.Errorf("fetchHeaderAndAviForAccount: domain %s is blocked", accountURI.Host)
	}

	if targetAccount.AvatarRemoteURL != "" && (targetAccount.AvatarMediaAttachmentID == "" || refresh) {
		a, err := d.mediaHandler.ProcessRemoteHeaderOrAvatar(t, &gtsmodel.MediaAttachment{
			RemoteURL: targetAccount.AvatarRemoteURL,
			Avatar:    true,
		}, targetAccount.ID)
		if err != nil {
			return fmt.Errorf("error processing avatar for user: %s", err)
		}
		targetAccount.AvatarMediaAttachmentID = a.ID
	}

	if targetAccount.HeaderRemoteURL != "" && (targetAccount.HeaderMediaAttachmentID == "" || refresh) {
		a, err := d.mediaHandler.ProcessRemoteHeaderOrAvatar(t, &gtsmodel.MediaAttachment{
			RemoteURL: targetAccount.HeaderRemoteURL,
			Header:    true,
		}, targetAccount.ID)
		if err != nil {
			return fmt.Errorf("error processing header for user: %s", err)
		}
		targetAccount.HeaderMediaAttachmentID = a.ID
	}
	return nil
}

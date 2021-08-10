/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

// EnrichRemoteAccount takes an account that's already been inserted into the database in a minimal form,
// and populates it with additional fields, media, etc.
//
// EnrichRemoteAccount is mostly useful for calling after an account has been initially created by
// the federatingDB's Create function, or during the federated authorization flow.
func (d *deref) EnrichRemoteAccount(username string, account *gtsmodel.Account) (*gtsmodel.Account, error) {
	if err := d.PopulateAccountFields(account, username, false); err != nil {
		return nil, err
	}

	if err := d.db.UpdateByID(account.ID, account); err != nil {
		return nil, fmt.Errorf("EnrichRemoteAccount: error updating account: %s", err)
	}

	return account, nil
}

// GetRemoteAccount completely dereferences a remote account, converts it to a GtS model account,
// puts it in the database, and returns it to a caller. The boolean indicates whether the account is new
// to us or not. If we haven't seen the account before, bool will be true. If we have seen the account before,
// it will be false.
//
// Refresh indicates whether--if the account exists in our db already--it should be refreshed by calling
// the remote instance again.
//
// SIDE EFFECTS: remote account will be stored in the database, or updated if it already exists (and refresh is true).
func (d *deref) GetRemoteAccount(username string, remoteAccountID *url.URL, refresh bool) (*gtsmodel.Account, bool, error) {
	new := true

	// check if we already have the account in our db
	maybeAccount := &gtsmodel.Account{}
	if err := d.db.GetWhere([]db.Where{{Key: "uri", Value: remoteAccountID.String()}}, maybeAccount); err == nil {
		// we've seen this account before so it's not new
		new = false
		if !refresh {
			// we're not being asked to refresh, but just in case we don't have the avatar/header cached yet....
			maybeAccount, err = d.EnrichRemoteAccount(username, maybeAccount)
			return maybeAccount, new, err
		}
	}

	accountable, err := d.dereferenceAccountable(username, remoteAccountID)
	if err != nil {
		return nil, new, fmt.Errorf("FullyDereferenceAccount: error dereferencing accountable: %s", err)
	}

	gtsAccount, err := d.typeConverter.ASRepresentationToAccount(accountable, refresh)
	if err != nil {
		return nil, new, fmt.Errorf("FullyDereferenceAccount: error converting accountable to account: %s", err)
	}

	if new {
		// generate a new id since we haven't seen this account before, and do a put
		ulid, err := id.NewRandomULID()
		if err != nil {
			return nil, new, fmt.Errorf("FullyDereferenceAccount: error generating new id for account: %s", err)
		}
		gtsAccount.ID = ulid

		if err := d.PopulateAccountFields(gtsAccount, username, refresh); err != nil {
			return nil, new, fmt.Errorf("FullyDereferenceAccount: error populating further account fields: %s", err)
		}

		if err := d.db.Put(gtsAccount); err != nil {
			return nil, new, fmt.Errorf("FullyDereferenceAccount: error putting new account: %s", err)
		}
	} else {
		// take the id we already have and do an update
		gtsAccount.ID = maybeAccount.ID

		if err := d.PopulateAccountFields(gtsAccount, username, refresh); err != nil {
			return nil, new, fmt.Errorf("FullyDereferenceAccount: error populating further account fields: %s", err)
		}

		if err := d.db.UpdateByID(gtsAccount.ID, gtsAccount); err != nil {
			return nil, new, fmt.Errorf("FullyDereferenceAccount: error updating existing account: %s", err)
		}
	}

	return gtsAccount, new, nil
}

// dereferenceAccountable calls remoteAccountID with a GET request, and tries to parse whatever
// it finds as something that an account model can be constructed out of.
//
// Will work for Person, Application, or Service models.
func (d *deref) dereferenceAccountable(username string, remoteAccountID *url.URL) (ap.Accountable, error) {
	d.startHandshake(username, remoteAccountID)
	defer d.stopHandshake(username, remoteAccountID)

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

// PopulateAccountFields populates any fields on the given account that weren't populated by the initial
// dereferencing. This includes things like header and avatar etc.
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

	return nil
}

// fetchHeaderAndAviForAccount fetches the header and avatar for a remote account, using a transport
// on behalf of requestingUsername.
//
// targetAccount's AvatarMediaAttachmentID and HeaderMediaAttachmentID will be updated as necessary.
//
// SIDE EFFECTS: remote header and avatar will be stored in local storage.
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

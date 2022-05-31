/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"io"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

func instanceAccount(account *gtsmodel.Account) bool {
	return strings.EqualFold(account.Username, account.Domain) ||
		account.FollowersURI == "" ||
		account.FollowingURI == "" ||
		(account.Username == "internal.fetch" && strings.Contains(account.Note, "internal service actor"))
}

// GetRemoteAccountParams wraps parameters for a remote account lookup.
type GetRemoteAccountParams struct {
	// The username of the user doing the lookup request (optional).
	// If not set, then the GtS instance account will be used to do the lookup.
	RequestingUsername string
	// The ActivityPub URI of the remote account (optional).
	// If not set (nil), the ActivityPub URI of the remote account will be discovered
	// via webfinger, so you must set RemoteAccountUsername and RemoteAccountHost
	// if this parameter is not set.
	RemoteAccountID *url.URL
	// The username of the remote account (optional).
	// If RemoteAccountID is not set, then this value must be set.
	RemoteAccountUsername string
	// The host of the remote account (optional).
	// If RemoteAccountID is not set, then this value must be set.
	RemoteAccountHost string
	// Whether to do a blocking call to the remote instance. If true,
	// then the account's media and other fields will be fully dereferenced before it is returned.
	// If false, then the account's media and other fields will be dereferenced in the background,
	// so only a minimal account representation will be returned by GetRemoteAccount.
	Blocking bool
	// Whether to refresh the account by performing dereferencing all over again.
	// If true, the account will be updated and returned.
	// If false, and the account already exists in the database, then that will be returned instead.
	Refresh bool
}

func (d *deref) populateAccountBeforeReturn(ctx context.Context, params GetRemoteAccountParams, remoteAccount *gtsmodel.Account) (*gtsmodel.Account, error) {
	// make sure the account fields are populated before returning:
	// even if we're not doing a refresh, the caller might want to block
	// until everything is loaded
	changed, err := d.populateAccountFields(ctx, remoteAccount, params.RequestingUsername, params.Refresh, params.Blocking)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteAccount: error populating remoteAccount fields: %s", err)
	}

	if changed {
		remoteAccount, err = d.db.UpdateAccount(ctx, remoteAccount)
		if err != nil {
			return nil, fmt.Errorf("GetRemoteAccount: error updating remoteAccount: %s", err)
		}
	}

	return remoteAccount, nil
}

// GetRemoteAccount completely dereferences a remote account, converts it to a GtS model account,
// puts or updates it in the database (if necessary), and returns it to a caller.
//
// It will try to make as few remote calls as possible, unless 'Refresh' is set to true in params.
func (d *deref) GetRemoteAccount(ctx context.Context, params GetRemoteAccountParams) (remoteAccount *gtsmodel.Account, err error) {
	var new = true
	var accountDomain string
	var accountable ap.Accountable

	if params.RemoteAccountID == nil {
		if params.RemoteAccountUsername == "" || params.RemoteAccountHost == "" {
			err = errors.New("GetRemoteAccount: RemoteAccountID wasn't set, and RemoteAccountUsername/RemoteAccountHost weren't set either, so a lookup couldn't be performed")
			return
		}

		accountDomain, params.RemoteAccountID, err = d.fingerRemoteAccount(ctx, params.RequestingUsername, params.RemoteAccountUsername, params.RemoteAccountHost)
		if err != nil {
			err = fmt.Errorf("GetRemoteAccount: error while fingering: %s", err)
			return
		}
	}

	// if we reach here then params.RemoteAccountID must be set,
	// either in original params or through fingering

	// see if we have this account stored already
	if a, dbErr := d.db.GetAccountByURI(ctx, params.RemoteAccountID.String()); dbErr == nil {
		new = false
		remoteAccount = a
		if !params.Refresh {
			// found it and no need to go any further
			remoteAccount, err = d.populateAccountBeforeReturn(ctx, params, remoteAccount)
			return
		}
	}

	// if we reach here then

	if params.RemoteAccountID != nil && (params.RemoteAccountUsername != "" && params.RemoteAccountHost != "") {
		// scenario 2: all three things are defined, do a webfinger lookup just for the accountDomain
		accountDomain, _, err = d.fingerRemoteAccount(ctx, params.RequestingUsername, params.RemoteAccountUsername, params.RemoteAccountHost)
		if err != nil {
			err = fmt.Errorf("GetRemoteAccount: error while fingering: %s", err)
			return
		}
	} else if params.RemoteAccountID != nil && (params.RemoteAccountUsername == "" || params.RemoteAccountHost == "") {
		// scenario 3: remote account ID is defined but nothing else is
		return nil, errors.New("GetRemoteAccount: RemoteAccountID wasn't set, and RemoteAccountUsername/RemoteAccountHost weren't set either, so a lookup couldn't be performed")
	} else

	// check if we already have the account in our db, and just return it unless we'd doing a refresh
	if params.RemoteAccountID != nil {

	}

	if new {
		// we haven't seen this account before: dereference it from remote and store it in the database
		if accountable == nil {
			accountable, err = d.dereferenceAccountable(ctx, params.RequestingUsername, params.RemoteAccountID)
			if err != nil {
				return nil, fmt.Errorf("GetRemoteAccount: error dereferencing accountable: %s", err)
			}
		}

		newAccount, err := d.typeConverter.ASRepresentationToAccount(ctx, accountable, accountDomain, params.Refresh)
		if err != nil {
			return nil, fmt.Errorf("GetRemoteAccount: error converting accountable to account: %s", err)
		}

		ulid, err := id.NewRandomULID()
		if err != nil {
			return nil, fmt.Errorf("GetRemoteAccount: error generating new id for account: %s", err)
		}
		newAccount.ID = ulid

		if _, err := d.populateAccountFields(ctx, newAccount, params.RequestingUsername, params.Refresh, params.Blocking); err != nil {
			return nil, fmt.Errorf("GetRemoteAccount: error populating further account fields: %s", err)
		}

		if err := d.db.Put(ctx, newAccount); err != nil {
			return nil, fmt.Errorf("GetRemoteAccount: error putting new account: %s", err)
		}

		return newAccount, nil
	}

	// we have seen this account before, but we have to refresh it
	if accountable == nil {
		accountable, err = d.dereferenceAccountable(ctx, username, remoteAccountID)
		if err != nil {
			return nil, fmt.Errorf("GetRemoteAccount: error dereferencing refreshedAccountable: %s", err)
		}
	}

	refreshedAccount, err := d.typeConverter.ASRepresentationToAccount(ctx, accountable, accountDomain, refresh)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteAccount: error converting refreshedAccountable to refreshedAccount: %s", err)
	}
	refreshedAccount.ID = remoteAccount.ID

	remoteAccount, err = d.populateAccountBeforeReturn(ctx, params, remoteAccount)
	return
}

// dereferenceAccountable calls remoteAccountID with a GET request, and tries to parse whatever
// it finds as something that an account model can be constructed out of.
//
// Will work for Person, Application, or Service models.
func (d *deref) dereferenceAccountable(ctx context.Context, username string, remoteAccountID *url.URL) (ap.Accountable, error) {
	d.startHandshake(username, remoteAccountID)
	defer d.stopHandshake(username, remoteAccountID)

	if blocked, err := d.db.IsDomainBlocked(ctx, remoteAccountID.Host); blocked || err != nil {
		return nil, fmt.Errorf("DereferenceAccountable: domain %s is blocked", remoteAccountID.Host)
	}

	transport, err := d.transportController.NewTransportForUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("DereferenceAccountable: transport err: %s", err)
	}

	b, err := transport.Dereference(ctx, remoteAccountID)
	if err != nil {
		return nil, fmt.Errorf("DereferenceAccountable: error deferencing %s: %s", remoteAccountID.String(), err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("DereferenceAccountable: error unmarshalling bytes into json: %s", err)
	}

	t, err := streams.ToType(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("DereferenceAccountable: error resolving json into ap vocab type: %s", err)
	}

	switch t.GetTypeName() {
	case ap.ActorApplication:
		p, ok := t.(vocab.ActivityStreamsApplication)
		if !ok {
			return nil, errors.New("DereferenceAccountable: error resolving type as activitystreams application")
		}
		return p, nil
	case ap.ActorGroup:
		p, ok := t.(vocab.ActivityStreamsGroup)
		if !ok {
			return nil, errors.New("DereferenceAccountable: error resolving type as activitystreams group")
		}
		return p, nil
	case ap.ActorOrganization:
		p, ok := t.(vocab.ActivityStreamsOrganization)
		if !ok {
			return nil, errors.New("DereferenceAccountable: error resolving type as activitystreams organization")
		}
		return p, nil
	case ap.ActorPerson:
		p, ok := t.(vocab.ActivityStreamsPerson)
		if !ok {
			return nil, errors.New("DereferenceAccountable: error resolving type as activitystreams person")
		}
		return p, nil
	case ap.ActorService:
		p, ok := t.(vocab.ActivityStreamsService)
		if !ok {
			return nil, errors.New("DereferenceAccountable: error resolving type as activitystreams service")
		}
		return p, nil
	}

	return nil, fmt.Errorf("DereferenceAccountable: type name %s not supported", t.GetTypeName())
}

// populateAccountFields populates any fields on the given account that weren't populated by the initial
// dereferencing. This includes things like header and avatar etc.
func (d *deref) populateAccountFields(ctx context.Context, account *gtsmodel.Account, requestingUsername string, blocking bool, refresh bool) (bool, error) {
	// if we're dealing with an instance account, just bail, we don't need to do anything
	if instanceAccount(account) {
		return false, nil
	}

	accountURI, err := url.Parse(account.URI)
	if err != nil {
		return false, fmt.Errorf("populateAccountFields: couldn't parse account URI %s: %s", account.URI, err)
	}

	if blocked, err := d.db.IsDomainBlocked(ctx, accountURI.Host); blocked || err != nil {
		return false, fmt.Errorf("populateAccountFields: domain %s is blocked", accountURI.Host)
	}

	t, err := d.transportController.NewTransportForUsername(ctx, requestingUsername)
	if err != nil {
		return false, fmt.Errorf("populateAccountFields: error getting transport for user: %s", err)
	}

	// fetch the header and avatar
	changed, err := d.fetchRemoteAccountMedia(ctx, account, t, refresh, blocking)
	if err != nil {
		return false, fmt.Errorf("populateAccountFields: error fetching header/avi for account: %s", err)
	}

	return changed, nil
}

// fetchRemoteAccountMedia fetches and stores the header and avatar for a remote account,
// using a transport on behalf of requestingUsername.
//
// The returned boolean indicates whether anything changed -- in other words, whether the
// account should be updated in the database.
//
// targetAccount's AvatarMediaAttachmentID and HeaderMediaAttachmentID will be updated as necessary.
//
// If refresh is true, then the media will be fetched again even if it's already been fetched before.
//
// If blocking is true, then the calls to the media manager made by this function will be blocking:
// in other words, the function won't return until the header and the avatar have been fully processed.
func (d *deref) fetchRemoteAccountMedia(ctx context.Context, targetAccount *gtsmodel.Account, t transport.Transport, blocking bool, refresh bool) (bool, error) {
	changed := false

	accountURI, err := url.Parse(targetAccount.URI)
	if err != nil {
		return changed, fmt.Errorf("fetchRemoteAccountMedia: couldn't parse account URI %s: %s", targetAccount.URI, err)
	}

	if blocked, err := d.db.IsDomainBlocked(ctx, accountURI.Host); blocked || err != nil {
		return changed, fmt.Errorf("fetchRemoteAccountMedia: domain %s is blocked", accountURI.Host)
	}

	if targetAccount.AvatarRemoteURL != "" && (targetAccount.AvatarMediaAttachmentID == "" || refresh) {
		var processingMedia *media.ProcessingMedia

		d.dereferencingAvatarsLock.Lock() // LOCK HERE
		// first check if we're already processing this media
		if alreadyProcessing, ok := d.dereferencingAvatars[targetAccount.ID]; ok {
			// we're already on it, no worries
			processingMedia = alreadyProcessing
		}

		if processingMedia == nil {
			// we're not already processing it so start now
			avatarIRI, err := url.Parse(targetAccount.AvatarRemoteURL)
			if err != nil {
				d.dereferencingAvatarsLock.Unlock()
				return changed, err
			}

			data := func(innerCtx context.Context) (io.Reader, int, error) {
				return t.DereferenceMedia(innerCtx, avatarIRI)
			}

			avatar := true
			newProcessing, err := d.mediaManager.ProcessMedia(ctx, data, nil, targetAccount.ID, &media.AdditionalMediaInfo{
				RemoteURL: &targetAccount.AvatarRemoteURL,
				Avatar:    &avatar,
			})
			if err != nil {
				d.dereferencingAvatarsLock.Unlock()
				return changed, err
			}

			// store it in our map to indicate it's in process
			d.dereferencingAvatars[targetAccount.ID] = newProcessing
			processingMedia = newProcessing
		}
		d.dereferencingAvatarsLock.Unlock() // UNLOCK HERE

		// block until loaded if required...
		if blocking {
			if err := lockAndLoad(ctx, d.dereferencingAvatarsLock, processingMedia, d.dereferencingAvatars, targetAccount.ID); err != nil {
				return changed, err
			}
		} else {
			// ...otherwise do it async
			go func() {
				dlCtx, done := context.WithDeadline(context.Background(), time.Now().Add(1*time.Minute))
				if err := lockAndLoad(dlCtx, d.dereferencingAvatarsLock, processingMedia, d.dereferencingAvatars, targetAccount.ID); err != nil {
					logrus.Errorf("fetchRemoteAccountMedia: error during async lock and load of avatar: %s", err)
				}
				done()
			}()
		}

		targetAccount.AvatarMediaAttachmentID = processingMedia.AttachmentID()
		changed = true
	}

	if targetAccount.HeaderRemoteURL != "" && (targetAccount.HeaderMediaAttachmentID == "" || refresh) {
		var processingMedia *media.ProcessingMedia

		d.dereferencingHeadersLock.Lock() // LOCK HERE
		// first check if we're already processing this media
		if alreadyProcessing, ok := d.dereferencingHeaders[targetAccount.ID]; ok {
			// we're already on it, no worries
			processingMedia = alreadyProcessing
		}

		if processingMedia == nil {
			// we're not already processing it so start now
			headerIRI, err := url.Parse(targetAccount.HeaderRemoteURL)
			if err != nil {
				d.dereferencingAvatarsLock.Unlock()
				return changed, err
			}

			data := func(innerCtx context.Context) (io.Reader, int, error) {
				return t.DereferenceMedia(innerCtx, headerIRI)
			}

			header := true
			newProcessing, err := d.mediaManager.ProcessMedia(ctx, data, nil, targetAccount.ID, &media.AdditionalMediaInfo{
				RemoteURL: &targetAccount.HeaderRemoteURL,
				Header:    &header,
			})
			if err != nil {
				d.dereferencingAvatarsLock.Unlock()
				return changed, err
			}

			// store it in our map to indicate it's in process
			d.dereferencingHeaders[targetAccount.ID] = newProcessing
			processingMedia = newProcessing
		}
		d.dereferencingHeadersLock.Unlock() // UNLOCK HERE

		// block until loaded if required...
		if blocking {
			if err := lockAndLoad(ctx, d.dereferencingHeadersLock, processingMedia, d.dereferencingHeaders, targetAccount.ID); err != nil {
				return changed, err
			}
		} else {
			// ...otherwise do it async
			go func() {
				dlCtx, done := context.WithDeadline(context.Background(), time.Now().Add(1*time.Minute))
				if err := lockAndLoad(dlCtx, d.dereferencingHeadersLock, processingMedia, d.dereferencingHeaders, targetAccount.ID); err != nil {
					logrus.Errorf("fetchRemoteAccountMedia: error during async lock and load of header: %s", err)
				}
				done()
			}()
		}

		targetAccount.HeaderMediaAttachmentID = processingMedia.AttachmentID()
		changed = true
	}

	return changed, nil
}

func lockAndLoad(ctx context.Context, lock *sync.Mutex, processing *media.ProcessingMedia, processingMap map[string]*media.ProcessingMedia, accountID string) error {
	// whatever happens, remove the in-process media from the map
	defer func() {
		lock.Lock()
		delete(processingMap, accountID)
		lock.Unlock()
	}()

	// try and load it
	_, err := processing.LoadAttachment(ctx)
	return err
}

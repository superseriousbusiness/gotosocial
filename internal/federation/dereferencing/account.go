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

// GetRemoteAccount completely dereferences a remote account, converts it to a GtS model account,
// puts it in the database, and returns it to a caller.
//
// Refresh indicates whether--if the account exists in our db already--it should be refreshed by calling
// the remote instance again. Blocking indicates whether the function should block until processing of
// the fetched account is complete.
//
// SIDE EFFECTS: remote account will be stored in the database, or updated if it already exists (and refresh is true).
func (d *deref) GetRemoteAccount(ctx context.Context, username string, remoteAccountID *url.URL, blocking bool, refresh bool) (*gtsmodel.Account, error) {
	new := true

	// check if we already have the account in our db, and just return it unless we'd doing a refresh
	remoteAccount, err := d.db.GetAccountByURI(ctx, remoteAccountID.String())
	if err == nil {
		new = false
		if !refresh {
			// make sure the account fields are populated before returning:
			// even if we're not doing a refresh, the caller might want to block
			// until everything is loaded
			changed, err := d.populateAccountFields(ctx, remoteAccount, username, refresh, blocking)
			if err != nil {
				return nil, fmt.Errorf("GetRemoteAccount: error populating remoteAccount fields: %s", err)
			}

			if changed {
				updatedAccount, err := d.db.UpdateAccount(ctx, remoteAccount)
				if err != nil {
					return nil, fmt.Errorf("GetRemoteAccount: error updating remoteAccount: %s", err)
				}
				return updatedAccount, err
			}

			return remoteAccount, nil
		}
	}

	if new {
		// we haven't seen this account before: dereference it from remote
		accountable, err := d.dereferenceAccountable(ctx, username, remoteAccountID)
		if err != nil {
			return nil, fmt.Errorf("GetRemoteAccount: error dereferencing accountable: %s", err)
		}

		newAccount, err := d.typeConverter.ASRepresentationToAccount(ctx, accountable, refresh)
		if err != nil {
			return nil, fmt.Errorf("GetRemoteAccount: error converting accountable to account: %s", err)
		}

		ulid, err := id.NewRandomULID()
		if err != nil {
			return nil, fmt.Errorf("GetRemoteAccount: error generating new id for account: %s", err)
		}
		newAccount.ID = ulid

		if _, err := d.populateAccountFields(ctx, newAccount, username, refresh, blocking); err != nil {
			return nil, fmt.Errorf("GetRemoteAccount: error populating further account fields: %s", err)
		}

		if err := d.db.Put(ctx, newAccount); err != nil {
			return nil, fmt.Errorf("GetRemoteAccount: error putting new account: %s", err)
		}

		return newAccount, nil
	}

	// we have seen this account before, but we have to refresh it
	refreshedAccountable, err := d.dereferenceAccountable(ctx, username, remoteAccountID)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteAccount: error dereferencing refreshedAccountable: %s", err)
	}

	refreshedAccount, err := d.typeConverter.ASRepresentationToAccount(ctx, refreshedAccountable, refresh)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteAccount: error converting refreshedAccountable to refreshedAccount: %s", err)
	}
	refreshedAccount.ID = remoteAccount.ID

	changed, err := d.populateAccountFields(ctx, refreshedAccount, username, refresh, blocking)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteAccount: error populating further refreshedAccount fields: %s", err)
	}

	if changed {
		updatedAccount, err := d.db.UpdateAccount(ctx, refreshedAccount)
		if err != nil {
			return nil, fmt.Errorf("GetRemoteAccount: error updating refreshedAccount: %s", err)
		}
		return updatedAccount, nil
	}

	return refreshedAccount, nil
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

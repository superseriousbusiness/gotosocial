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
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

var webfingerInterval = -48 * time.Hour // 2 days in the past

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
	// Whether to skip making calls to remote instances. This is useful when you want to
	// quickly fetch a remote account from the database or fail, and don't want to cause
	// http requests to go flying around.
	SkipResolve bool
}

// GetRemoteAccount completely dereferences a remote account, converts it to a GtS model account,
// puts or updates it in the database (if necessary), and returns it to a caller.
func (d *deref) GetRemoteAccount(ctx context.Context, params GetRemoteAccountParams) (remoteAccount *gtsmodel.Account, err error) {

	/*
		In this function we want to retrieve a gtsmodel representation of a remote account, with its proper
		accountDomain set, while making as few calls to remote instances as possible to save time and bandwidth.

		There are a few different paths through this function, and the path taken depends on how much
		initial information we are provided with via parameters, how much information we already have stored,
		and what we're allowed to do according to the parameters we've been passed.

		Scenario 1: We're not allowed to resolve remotely, but we've got either the account URI or the
		            account username + host, so we can check in our database and return if possible.

		Scenario 2: We are allowed to resolve remotely, and we have an account URI but no username or host.
		            In this case, we can use the URI to resolve the remote account and find the username,
					and then we can webfinger the account to discover the accountDomain if necessary.

		Scenario 3: We are allowed to resolve remotely, and we have the username and host but no URI.
		            In this case, we can webfinger the account to discover the URI, and then dereference
					from that.
	*/

	// first check if we can retrieve the account locally just with what we've been given
	switch {
	case params.RemoteAccountID != nil:
		// try with uri
		if a, dbErr := d.db.GetAccountByURI(ctx, params.RemoteAccountID.String()); dbErr == nil {
			remoteAccount = a
		} else if dbErr != db.ErrNoEntries {
			err = fmt.Errorf("GetRemoteAccount: database error looking for account %s: %s", params.RemoteAccountID, err)
		}
	case params.RemoteAccountUsername != "" && params.RemoteAccountHost != "":
		// try with username/host
		a := &gtsmodel.Account{}
		where := []db.Where{{Key: "username", Value: params.RemoteAccountUsername}, {Key: "domain", Value: params.RemoteAccountHost}}
		if dbErr := d.db.GetWhere(ctx, where, a); dbErr == nil {
			remoteAccount = a
		} else if dbErr != db.ErrNoEntries {
			err = fmt.Errorf("GetRemoteAccount: database error looking for account with username %s and host %s: %s", params.RemoteAccountUsername, params.RemoteAccountHost, err)
		}
	default:
		err = errors.New("GetRemoteAccount: no identifying parameters were set so we cannot get account")
	}

	if err != nil {
		return
	}

	if params.SkipResolve {
		// if we can't resolve, return already since there's nothing more we can do
		if remoteAccount == nil {
			err = errors.New("GetRemoteAccount: error retrieving account with skipResolve set true")
		}
		return
	}

	var accountable ap.Accountable
	if params.RemoteAccountUsername == "" || params.RemoteAccountHost == "" {
		// try to populate the missing params
		// the first one is easy ...
		params.RemoteAccountHost = params.RemoteAccountID.Host
		// ... but we still need the username so we can do a finger for the accountDomain

		// check if we had the account stored already and got it earlier
		if remoteAccount != nil {
			params.RemoteAccountUsername = remoteAccount.Username
		} else {
			// if we didn't already have it, we have dereference it from remote and just...
			accountable, err = d.dereferenceAccountable(ctx, params.RequestingUsername, params.RemoteAccountID)
			if err != nil {
				err = fmt.Errorf("GetRemoteAccount: error dereferencing accountable: %s", err)
				return
			}

			// ... take the username (for now)
			params.RemoteAccountUsername, err = ap.ExtractPreferredUsername(accountable)
			if err != nil {
				err = fmt.Errorf("GetRemoteAccount: error extracting accountable username: %s", err)
				return
			}
		}
	}

	// if we reach this point, params.RemoteAccountHost and params.RemoteAccountUsername must be set
	// params.RemoteAccountID may or may not be set, but we have enough information to fetch it if we need it

	// we finger to fetch the account domain but just in case we're not fingering, make a best guess
	// already about what the account domain might be; this var will be overwritten later if necessary
	var accountDomain string
	switch {
	case remoteAccount != nil:
		accountDomain = remoteAccount.Domain
	case params.RemoteAccountID != nil:
		accountDomain = params.RemoteAccountID.Host
	default:
		accountDomain = params.RemoteAccountHost
	}

	// to save on remote calls: only webfinger if we don't have a remoteAccount yet, or if we haven't
	// fingered the remote account for at least 2 days; don't finger instance accounts
	var fingered time.Time
	if remoteAccount == nil || (remoteAccount.LastWebfingeredAt.Before(time.Now().Add(webfingerInterval)) && !instanceAccount(remoteAccount)) {
		accountDomain, params.RemoteAccountID, err = d.fingerRemoteAccount(ctx, params.RequestingUsername, params.RemoteAccountUsername, params.RemoteAccountHost)
		if err != nil {
			err = fmt.Errorf("GetRemoteAccount: error while fingering: %s", err)
			return
		}
		fingered = time.Now()
	}

	if !fingered.IsZero() && remoteAccount == nil {
		// if we just fingered and now have a discovered account domain but still no account,
		// we should do a final lookup in the database with the discovered username + accountDomain
		// to make absolutely sure we don't already have this account
		a := &gtsmodel.Account{}
		where := []db.Where{{Key: "username", Value: params.RemoteAccountUsername}, {Key: "domain", Value: accountDomain}}
		if dbErr := d.db.GetWhere(ctx, where, a); dbErr == nil {
			remoteAccount = a
		} else if dbErr != db.ErrNoEntries {
			err = fmt.Errorf("GetRemoteAccount: database error looking for account with username %s and host %s: %s", params.RemoteAccountUsername, params.RemoteAccountHost, err)
			return
		}
	}

	// we may also have some extra information already, like the account we had in the db, or the
	// accountable representation that we dereferenced from remote
	if remoteAccount == nil {
		// we still don't have the account, so deference it if we didn't earlier
		if accountable == nil {
			accountable, err = d.dereferenceAccountable(ctx, params.RequestingUsername, params.RemoteAccountID)
			if err != nil {
				err = fmt.Errorf("GetRemoteAccount: error dereferencing accountable: %s", err)
				return
			}
		}

		// then convert
		remoteAccount, err = d.typeConverter.ASRepresentationToAccount(ctx, accountable, accountDomain, false)
		if err != nil {
			err = fmt.Errorf("GetRemoteAccount: error converting accountable to account: %s", err)
			return
		}

		// this is a new account so we need to generate a new ID for it
		var ulid string
		ulid, err = id.NewRandomULID()
		if err != nil {
			err = fmt.Errorf("GetRemoteAccount: error generating new id for account: %s", err)
			return
		}
		remoteAccount.ID = ulid

		_, err = d.populateAccountFields(ctx, remoteAccount, params.RequestingUsername, params.Blocking)
		if err != nil {
			err = fmt.Errorf("GetRemoteAccount: error populating further account fields: %s", err)
			return
		}

		remoteAccount.LastWebfingeredAt = fingered
		remoteAccount.UpdatedAt = time.Now()

		err = d.db.Put(ctx, remoteAccount)
		if err != nil {
			err = fmt.Errorf("GetRemoteAccount: error putting new account: %s", err)
			return
		}

		return // the new account
	}

	// we had the account already, but now we know the account domain, so update it if it's different
	if !strings.EqualFold(remoteAccount.Domain, accountDomain) {
		remoteAccount.Domain = accountDomain
		remoteAccount, err = d.db.UpdateAccount(ctx, remoteAccount)
		if err != nil {
			err = fmt.Errorf("GetRemoteAccount: error updating account: %s", err)
			return
		}
	}

	// make sure the account fields are populated before returning:
	// the caller might want to block until everything is loaded
	var fieldsChanged bool
	fieldsChanged, err = d.populateAccountFields(ctx, remoteAccount, params.RequestingUsername, params.Blocking)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteAccount: error populating remoteAccount fields: %s", err)
	}

	var fingeredChanged bool
	if !fingered.IsZero() {
		fingeredChanged = true
		remoteAccount.LastWebfingeredAt = fingered
	}

	if fieldsChanged || fingeredChanged {
		remoteAccount.UpdatedAt = time.Now()
		remoteAccount, err = d.db.UpdateAccount(ctx, remoteAccount)
		if err != nil {
			return nil, fmt.Errorf("GetRemoteAccount: error updating remoteAccount: %s", err)
		}
	}

	return // the account we already had + possibly updated
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
func (d *deref) populateAccountFields(ctx context.Context, account *gtsmodel.Account, requestingUsername string, blocking bool) (bool, error) {
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
	changed, err := d.fetchRemoteAccountMedia(ctx, account, t, blocking)
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
func (d *deref) fetchRemoteAccountMedia(ctx context.Context, targetAccount *gtsmodel.Account, t transport.Transport, blocking bool) (bool, error) {
	changed := false

	accountURI, err := url.Parse(targetAccount.URI)
	if err != nil {
		return changed, fmt.Errorf("fetchRemoteAccountMedia: couldn't parse account URI %s: %s", targetAccount.URI, err)
	}

	if blocked, err := d.db.IsDomainBlocked(ctx, accountURI.Host); blocked || err != nil {
		return changed, fmt.Errorf("fetchRemoteAccountMedia: domain %s is blocked", accountURI.Host)
	}

	if targetAccount.AvatarRemoteURL != "" && (targetAccount.AvatarMediaAttachmentID == "") {
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

	if targetAccount.HeaderRemoteURL != "" && (targetAccount.HeaderMediaAttachmentID == "") {
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

/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"time"

	"github.com/miekg/dns"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
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

// GetAccountParams wraps parameters for an account lookup.
type GetAccountParams struct {
	// The username of the user doing the lookup request (optional).
	// If not set, then the GtS instance account will be used to do the lookup.
	RequestingUsername string
	// The ActivityPub URI of the account (optional).
	// If not set (nil), the ActivityPub URI of the account will be discovered
	// via webfinger, so you must set RemoteAccountUsername and RemoteAccountHost
	// if this parameter is not set.
	RemoteAccountID *url.URL
	// The username of the account (optional).
	// If RemoteAccountID is not set, then this value must be set.
	RemoteAccountUsername string
	// The host of the account (optional).
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
	// PartialAccount can be used if the GetRemoteAccount call results from a federated/ap
	// account update. In this case, we will already have a partial representation of the account,
	// derived from converting the AP representation to a gtsmodel representation. If this field
	// is provided, then GetRemoteAccount will use this as a basis for building the full account.
	PartialAccount *gtsmodel.Account
}

type lookupType int

const (
	lookupPartialLocal lookupType = iota
	lookupPartial
	lookupURILocal
	lookupURI
	lookupMentionLocal
	lookupMention
	lookupBad
)

func getLookupType(params GetAccountParams) lookupType {
	switch {
	case params.PartialAccount != nil:
		if params.PartialAccount.Domain == "" || params.PartialAccount.Domain == config.GetHost() || params.PartialAccount.Domain == config.GetAccountDomain() {
			return lookupPartialLocal
		}
		return lookupPartial
	case params.RemoteAccountID != nil:
		if host := params.RemoteAccountID.Host; host == config.GetHost() || host == config.GetAccountDomain() {
			return lookupURILocal
		}
		return lookupURI
	case params.RemoteAccountUsername != "":
		if params.RemoteAccountHost == "" || params.RemoteAccountHost == config.GetHost() || params.RemoteAccountHost == config.GetAccountDomain() {
			return lookupMentionLocal
		}
		return lookupMention
	default:
		return lookupBad
	}
}

// GetAccount completely dereferences an account, converts it to a GtS model account,
// puts or updates it in the database (if necessary), and returns it to a caller.
//
// GetAccount will guard against trying to do http calls to fetch an account that belongs to this instance.
// Instead of making calls, it will just return the account early if it finds it, or return an error.
//
// Even if a fastfail context is used, and something goes wrong, an account might still be returned instead
// of an error, if we already had the account in our database (in other words, if we just needed to try
// fingering/refreshing the account again). The rationale for this is that it's more useful to be able
// to provide *something* to the caller, even if that something is not necessarily 100% up to date.
func (d *deref) GetAccount(ctx context.Context, params GetAccountParams) (foundAccount *gtsmodel.Account, err error) {
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

	// this first step checks if we have the
	// account in the database somewhere already,
	// or if we've been provided it as a partial
	switch getLookupType(params) {
	case lookupPartialLocal:
		params.SkipResolve = true
		fallthrough
	case lookupPartial:
		foundAccount = params.PartialAccount
	case lookupURILocal:
		params.SkipResolve = true
		fallthrough
	case lookupURI:
		// see if we have this in the db already with this uri/url
		uri := params.RemoteAccountID.String()

		if a, dbErr := d.db.GetAccountByURI(ctx, uri); dbErr == nil {
			// got it, break here to leave early
			foundAccount = a
			break
		} else if !errors.Is(dbErr, db.ErrNoEntries) {
			// a real error
			err = newErrDB(fmt.Errorf("GetRemoteAccount: unexpected error while looking for account with uri %s: %w", uri, dbErr))
			break
		}

		// dbErr was just db.ErrNoEntries so search by url instead
		if a, dbErr := d.db.GetAccountByURL(ctx, uri); dbErr == nil {
			// got it
			foundAccount = a
			break
		} else if !errors.Is(dbErr, db.ErrNoEntries) {
			// a real error
			err = newErrDB(fmt.Errorf("GetRemoteAccount: unexpected error while looking for account with url %s: %w", uri, dbErr))
			break
		}
	case lookupMentionLocal:
		params.SkipResolve = true
		params.RemoteAccountHost = ""
		fallthrough
	case lookupMention:
		// see if we have this in the db already with this username/host
		if a, dbErr := d.db.GetAccountByUsernameDomain(ctx, params.RemoteAccountUsername, params.RemoteAccountHost); dbErr == nil {
			foundAccount = a
		} else if !errors.Is(dbErr, db.ErrNoEntries) {
			// a real error
			err = newErrDB(fmt.Errorf("GetRemoteAccount: unexpected error while looking for account %s: %w", params.RemoteAccountUsername, dbErr))
		}
	default:
		err = newErrBadRequest(errors.New("GetRemoteAccount: no identifying parameters were set so we cannot get account"))
	}

	// bail if we've set a real error, and not just no entries in the db
	if err != nil {
		return
	}

	if params.SkipResolve {
		// if we can't resolve, return already since there's nothing more we can do
		if foundAccount == nil {
			err = newErrNotRetrievable(errors.New("GetRemoteAccount: couldn't retrieve account locally and not allowed to resolve it"))
		}
		return
	}

	// if we reach this point, we have some remote calls to make

	var accountable ap.Accountable
	if params.RemoteAccountUsername == "" && params.RemoteAccountHost == "" {
		// if we're still missing some params, try to populate them now
		params.RemoteAccountHost = params.RemoteAccountID.Host
		if foundAccount != nil {
			// username is easy if we found something already
			params.RemoteAccountUsername = foundAccount.Username
		} else {
			// if we didn't already have it, we have to dereference it from remote
			var derefErr error
			accountable, derefErr = d.dereferenceAccountable(ctx, params.RequestingUsername, params.RemoteAccountID)
			if derefErr != nil {
				err = wrapDerefError(derefErr, "GetRemoteAccount: error dereferencing Accountable")
				return
			}

			var apError error
			params.RemoteAccountUsername, apError = ap.ExtractPreferredUsername(accountable)
			if apError != nil {
				err = newErrOther(fmt.Errorf("GetRemoteAccount: error extracting Accountable username: %w", apError))
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
	case foundAccount != nil:
		accountDomain = foundAccount.Domain
	case params.RemoteAccountID != nil:
		accountDomain = params.RemoteAccountID.Host
	default:
		accountDomain = params.RemoteAccountHost
	}

	// to save on remote calls, only webfinger if:
	// - we don't know the remote account ActivityPub ID yet OR
	// - we haven't found the account yet in some other way OR
	// - we were passed a partial account in params OR
	// - we haven't webfingered the account for two days AND the account isn't an instance account
	var fingered time.Time
	var refreshFinger bool
	if foundAccount != nil {
		refreshFinger = foundAccount.LastWebfingeredAt.Before(time.Now().Add(webfingerInterval)) && !instanceAccount(foundAccount)
	}

	if params.RemoteAccountID == nil || foundAccount == nil || params.PartialAccount != nil || refreshFinger {
		if ad, accountURI, fingerError := d.fingerRemoteAccount(ctx, params.RequestingUsername, params.RemoteAccountUsername, params.RemoteAccountHost); fingerError != nil {
			if !refreshFinger {
				// only return with an error if this wasn't just a refresh finger;
				// that is, if we actually *needed* to finger in order to get the account,
				// otherwise we can just continue and we'll try again in 2 days
				err = newErrNotRetrievable(fmt.Errorf("GetRemoteAccount: error while fingering: %w", fingerError))
				return
			}
			log.Infof("error doing non-vital webfinger refresh call to %s: %s", params.RemoteAccountHost, err)
		} else {
			accountDomain = ad
			params.RemoteAccountID = accountURI
		}
		fingered = time.Now()
	}

	if !fingered.IsZero() && foundAccount == nil {
		// if we just fingered and now have a discovered account domain but still no account,
		// we should do a final lookup in the database with the discovered username + accountDomain
		// to make absolutely sure we don't already have this account
		if a, dbErr := d.db.GetAccountByUsernameDomain(ctx, params.RemoteAccountUsername, accountDomain); dbErr == nil {
			foundAccount = a
		} else if !errors.Is(dbErr, db.ErrNoEntries) {
			// a real error
			err = newErrDB(fmt.Errorf("GetRemoteAccount: unexpected error while looking for account %s: %w", params.RemoteAccountUsername, dbErr))
			return
		}
	}

	// we may have some extra information already, like the account we had in the db, or the
	// accountable representation that we dereferenced from remote
	if foundAccount == nil {
		// if we still don't have a remoteAccountID here we're boned
		if params.RemoteAccountID == nil {
			err = newErrNotRetrievable(errors.New("GetRemoteAccount: could not populate find an account nor populate params.RemoteAccountID"))
			return
		}

		// deference accountable if we didn't earlier
		if accountable == nil {
			var derefErr error
			accountable, derefErr = d.dereferenceAccountable(ctx, params.RequestingUsername, params.RemoteAccountID)
			if derefErr != nil {
				err = wrapDerefError(derefErr, "GetRemoteAccount: error dereferencing Accountable")
				return
			}
		}

		// then convert
		foundAccount, err = d.typeConverter.ASRepresentationToAccount(ctx, accountable, accountDomain, false)
		if err != nil {
			err = newErrOther(fmt.Errorf("GetRemoteAccount: error converting Accountable to account: %w", err))
			return
		}

		// this is a new account so we need to generate a new ID for it
		var ulid string
		ulid, err = id.NewRandomULID()
		if err != nil {
			err = newErrOther(fmt.Errorf("GetRemoteAccount: error generating new id for account: %w", err))
			return
		}
		foundAccount.ID = ulid

		if _, populateErr := d.populateAccountFields(ctx, foundAccount, params.RequestingUsername, params.Blocking); populateErr != nil {
			// it's not the end of the world if we can't populate account fields, but we do want to log it
			log.Errorf("GetRemoteAccount: error populating further account fields: %s", populateErr)
		}

		foundAccount.LastWebfingeredAt = fingered
		foundAccount.UpdatedAt = time.Now()

		if dbErr := d.db.PutAccount(ctx, foundAccount); dbErr != nil {
			err = newErrDB(fmt.Errorf("GetRemoteAccount: error putting new account: %w", dbErr))
			return
		}

		return // the new account
	}

	// we had the account already, but now we know the account domain, so update it if it's different
	var accountDomainChanged bool
	if !strings.EqualFold(foundAccount.Domain, accountDomain) {
		accountDomainChanged = true
		foundAccount.Domain = accountDomain
	}

	// if SharedInboxURI is nil, that means we don't know yet if this account has
	// a shared inbox available for it, so we need to check this here
	var sharedInboxChanged bool
	if foundAccount.SharedInboxURI == nil {
		// we need the accountable for this, so get it if we don't have it yet
		if accountable == nil {
			var derefErr error
			accountable, derefErr = d.dereferenceAccountable(ctx, params.RequestingUsername, params.RemoteAccountID)
			if derefErr != nil {
				err = wrapDerefError(derefErr, "GetRemoteAccount: error dereferencing Accountable")
				return
			}
		}

		// This can be:
		// - an empty string (we know it doesn't have a shared inbox) OR
		// - a string URL (we know it does a shared inbox).
		// Set it either way!
		var sharedInbox string

		if sharedInboxURI := ap.ExtractSharedInbox(accountable); sharedInboxURI != nil {
			// only trust shared inbox if it has at least two domains,
			// from the right, in common with the domain of the account
			if dns.CompareDomainName(foundAccount.Domain, sharedInboxURI.Host) >= 2 {
				sharedInbox = sharedInboxURI.String()
			}
		}

		sharedInboxChanged = true
		foundAccount.SharedInboxURI = &sharedInbox
	}

	// make sure the account fields are populated before returning:
	// the caller might want to block until everything is loaded
	fieldsChanged, populateErr := d.populateAccountFields(ctx, foundAccount, params.RequestingUsername, params.Blocking)
	if populateErr != nil {
		// it's not the end of the world if we can't populate account fields, but we do want to log it
		log.Errorf("GetRemoteAccount: error populating further account fields: %s", populateErr)
	}

	var fingeredChanged bool
	if !fingered.IsZero() {
		fingeredChanged = true
		foundAccount.LastWebfingeredAt = fingered
	}

	if accountDomainChanged || sharedInboxChanged || fieldsChanged || fingeredChanged {
		if dbErr := d.db.UpdateAccount(ctx, foundAccount); dbErr != nil {
			err = newErrDB(fmt.Errorf("GetRemoteAccount: error updating remoteAccount: %w", dbErr))
			return
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
		return nil, fmt.Errorf("DereferenceAccountable: transport err: %w", err)
	}

	b, err := transport.Dereference(ctx, remoteAccountID)
	if err != nil {
		return nil, fmt.Errorf("DereferenceAccountable: error deferencing %s: %w", remoteAccountID.String(), err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("DereferenceAccountable: error unmarshalling bytes into json: %w", err)
	}

	t, err := streams.ToType(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("DereferenceAccountable: error resolving json into ap vocab type: %w", err)
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

	return nil, newErrWrongType(fmt.Errorf("DereferenceAccountable: type name %s not supported as Accountable", t.GetTypeName()))
}

// populateAccountFields makes a best effort to populate fields on an account such as emojis, avatar, header.
// Will return true if one of these things changed on the passed-in account.
func (d *deref) populateAccountFields(ctx context.Context, account *gtsmodel.Account, requestingUsername string, blocking bool) (bool, error) {
	// if we're dealing with an instance account, just bail, we don't need to do anything
	if instanceAccount(account) {
		return false, nil
	}

	accountURI, err := url.Parse(account.URI)
	if err != nil {
		return false, fmt.Errorf("populateAccountFields: couldn't parse account URI %s: %w", account.URI, err)
	}

	blocked, dbErr := d.db.IsDomainBlocked(ctx, accountURI.Host)
	if dbErr != nil {
		return false, fmt.Errorf("populateAccountFields: eror checking for block of domain %s: %w", accountURI.Host, err)
	}

	if blocked {
		return false, fmt.Errorf("populateAccountFields: domain %s is blocked", accountURI.Host)
	}

	var changed bool

	// fetch the header and avatar
	if mediaChanged, err := d.fetchRemoteAccountMedia(ctx, account, requestingUsername, blocking); err != nil {
		return false, fmt.Errorf("populateAccountFields: error fetching header/avi for account: %w", err)
	} else if mediaChanged {
		changed = mediaChanged
	}

	// fetch any emojis used in note, fields, display name, etc
	if emojisChanged, err := d.fetchRemoteAccountEmojis(ctx, account, requestingUsername); err != nil {
		return false, fmt.Errorf("populateAccountFields: error fetching emojis for account: %w", err)
	} else if emojisChanged {
		changed = emojisChanged
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
func (d *deref) fetchRemoteAccountMedia(ctx context.Context, targetAccount *gtsmodel.Account, requestingUsername string, blocking bool) (bool, error) {
	var (
		changed bool
		t       transport.Transport
	)

	if targetAccount.AvatarRemoteURL != "" && (targetAccount.AvatarMediaAttachmentID == "") {
		var processingMedia *media.ProcessingMedia

		d.dereferencingAvatarsLock.Lock() // LOCK HERE
		// first check if we're already processing this media
		if alreadyProcessing, ok := d.dereferencingAvatars[targetAccount.ID]; ok {
			// we're already on it, no worries
			processingMedia = alreadyProcessing
		} else {
			// we're not already processing it so start now
			avatarIRI, err := url.Parse(targetAccount.AvatarRemoteURL)
			if err != nil {
				d.dereferencingAvatarsLock.Unlock()
				return changed, err
			}

			if t == nil {
				var err error
				t, err = d.transportController.NewTransportForUsername(ctx, requestingUsername)
				if err != nil {
					d.dereferencingAvatarsLock.Unlock()
					return false, fmt.Errorf("fetchRemoteAccountMedia: error getting transport for user: %s", err)
				}
			}

			data := func(innerCtx context.Context) (io.ReadCloser, int64, error) {
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

		load := func(innerCtx context.Context) error {
			_, err := processingMedia.LoadAttachment(innerCtx)
			return err
		}

		cleanup := func() {
			d.dereferencingAvatarsLock.Lock()
			delete(d.dereferencingAvatars, targetAccount.ID)
			d.dereferencingAvatarsLock.Unlock()
		}

		// block until loaded if required...
		if blocking {
			if err := loadAndCleanup(ctx, load, cleanup); err != nil {
				return changed, err
			}
		} else {
			// ...otherwise do it async
			go func() {
				dlCtx, done := context.WithDeadline(context.Background(), time.Now().Add(1*time.Minute))
				if err := loadAndCleanup(dlCtx, load, cleanup); err != nil {
					log.Errorf("fetchRemoteAccountMedia: error during async lock and load of avatar: %s", err)
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
		} else {
			// we're not already processing it so start now
			headerIRI, err := url.Parse(targetAccount.HeaderRemoteURL)
			if err != nil {
				d.dereferencingAvatarsLock.Unlock()
				return changed, err
			}

			if t == nil {
				var err error
				t, err = d.transportController.NewTransportForUsername(ctx, requestingUsername)
				if err != nil {
					d.dereferencingAvatarsLock.Unlock()
					return false, fmt.Errorf("fetchRemoteAccountMedia: error getting transport for user: %s", err)
				}
			}

			data := func(innerCtx context.Context) (io.ReadCloser, int64, error) {
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

		load := func(innerCtx context.Context) error {
			_, err := processingMedia.LoadAttachment(innerCtx)
			return err
		}

		cleanup := func() {
			d.dereferencingHeadersLock.Lock()
			delete(d.dereferencingHeaders, targetAccount.ID)
			d.dereferencingHeadersLock.Unlock()
		}

		// block until loaded if required...
		if blocking {
			if err := loadAndCleanup(ctx, load, cleanup); err != nil {
				return changed, err
			}
		} else {
			// ...otherwise do it async
			go func() {
				dlCtx, done := context.WithDeadline(context.Background(), time.Now().Add(1*time.Minute))
				if err := loadAndCleanup(dlCtx, load, cleanup); err != nil {
					log.Errorf("fetchRemoteAccountMedia: error during async lock and load of header: %s", err)
				}
				done()
			}()
		}

		targetAccount.HeaderMediaAttachmentID = processingMedia.AttachmentID()
		changed = true
	}

	return changed, nil
}

func (d *deref) fetchRemoteAccountEmojis(ctx context.Context, targetAccount *gtsmodel.Account, requestingUsername string) (bool, error) {
	maybeEmojis := targetAccount.Emojis
	maybeEmojiIDs := targetAccount.EmojiIDs

	// It's possible that the account had emoji IDs set on it, but not Emojis
	// themselves, depending on how it was fetched before being passed to us.
	//
	// If we only have IDs, fetch the emojis from the db. We know they're in
	// there or else they wouldn't have IDs.
	if len(maybeEmojiIDs) > len(maybeEmojis) {
		maybeEmojis = make([]*gtsmodel.Emoji, 0, len(maybeEmojiIDs))
		for _, emojiID := range maybeEmojiIDs {
			maybeEmoji, err := d.db.GetEmojiByID(ctx, emojiID)
			if err != nil {
				return false, err
			}
			maybeEmojis = append(maybeEmojis, maybeEmoji)
		}
	}

	// For all the maybe emojis we have, we either fetch them from the database
	// (if we haven't already), or dereference them from the remote instance.
	gotEmojis, err := d.populateEmojis(ctx, maybeEmojis, requestingUsername)
	if err != nil {
		return false, err
	}

	// Extract the ID of each fetched or dereferenced emoji, so we can attach
	// this to the account if necessary.
	gotEmojiIDs := make([]string, 0, len(gotEmojis))
	for _, e := range gotEmojis {
		gotEmojiIDs = append(gotEmojiIDs, e.ID)
	}

	var (
		changed  = false // have the emojis for this account changed?
		maybeLen = len(maybeEmojis)
		gotLen   = len(gotEmojis)
	)

	// if the length of everything is zero, this is simple:
	// nothing has changed and there's nothing to do
	if maybeLen == 0 && gotLen == 0 {
		return changed, nil
	}

	// if the *amount* of emojis on the account has changed, then the got emojis
	// are definitely different from the previous ones (if there were any) --
	// the account has either more or fewer emojis set on it now, so take the
	// discovered emojis as the new correct ones.
	if maybeLen != gotLen {
		changed = true
		targetAccount.Emojis = gotEmojis
		targetAccount.EmojiIDs = gotEmojiIDs
		return changed, nil
	}

	// if the lengths are the same but not all of the slices are
	// zero, something *might* have changed, so we have to check

	// 1. did we have emojis before that we don't have now?
	for _, maybeEmoji := range maybeEmojis {
		var stillPresent bool

		for _, gotEmoji := range gotEmojis {
			if maybeEmoji.URI == gotEmoji.URI {
				// the emoji we maybe had is still present now,
				// so we can stop checking gotEmojis
				stillPresent = true
				break
			}
		}

		if !stillPresent {
			// at least one maybeEmoji is no longer present in
			// the got emojis, so we can stop checking now
			changed = true
			targetAccount.Emojis = gotEmojis
			targetAccount.EmojiIDs = gotEmojiIDs
			return changed, nil
		}
	}

	// 2. do we have emojis now that we didn't have before?
	for _, gotEmoji := range gotEmojis {
		var wasPresent bool

		for _, maybeEmoji := range maybeEmojis {
			// check emoji IDs here as well, because unreferenced
			// maybe emojis we didn't already have would not have
			// had IDs set on them yet
			if gotEmoji.URI == maybeEmoji.URI && gotEmoji.ID == maybeEmoji.ID {
				// this got emoji was present already in the maybeEmoji,
				// so we can stop checking through maybeEmojis
				wasPresent = true
				break
			}
		}

		if !wasPresent {
			// at least one gotEmojis was not present in
			// the maybeEmojis, so we can stop checking now
			changed = true
			targetAccount.Emojis = gotEmojis
			targetAccount.EmojiIDs = gotEmojiIDs
			return changed, nil
		}
	}

	return changed, nil
}

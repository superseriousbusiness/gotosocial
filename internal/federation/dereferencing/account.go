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
	"time"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
)

func (d *deref) GetAccountByURI(ctx context.Context, requestUser string, uri *url.URL, block bool) (*gtsmodel.Account, error) {
	var (
		account *gtsmodel.Account
		uriStr  = uri.String()
		err     error
	)

	// Search the database for existing account with ID URI.
	account, err = d.db.GetAccountByURI(ctx, uriStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, fmt.Errorf("GetAccountByURI: error checking database for account %s by uri: %w", uriStr, err)
	}

	if account == nil {
		// Else, search the database for existing by ID URL.
		account, err = d.db.GetAccountByURL(ctx, uriStr)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, fmt.Errorf("GetAccountByURI: error checking database for account %s by url: %w", uriStr, err)
		}
	}

	if account == nil {
		// Ensure that this is isn't a search for a local account.
		if uri.Host == config.GetHost() || uri.Host == config.GetAccountDomain() {
			return nil, NewErrNotRetrievable(err) // this will be db.ErrNoEntries
		}

		// Create and pass-through a new bare-bones model for dereferencing.
		return d.enrichAccount(ctx, requestUser, uri, &gtsmodel.Account{
			ID:     id.NewULID(),
			Domain: uri.Host,
			URI:    uriStr,
		}, false, true)
	}

	// Try to update existing account model
	enriched, err := d.enrichAccount(ctx, requestUser, uri, account, false, block)
	if err != nil {
		log.Errorf("error enriching remote account: %v", err)
		return account, nil // fall back to returning existing
	}

	return enriched, nil
}

func (d *deref) GetAccountByUsernameDomain(ctx context.Context, requestUser string, username string, domain string, block bool) (*gtsmodel.Account, error) {
	if domain == config.GetHost() || domain == config.GetAccountDomain() {
		// We do local lookups using an empty domain,
		// else it will fail the db search below.
		domain = ""
	}

	// Search the database for existing account with USERNAME@DOMAIN
	account, err := d.db.GetAccountByUsernameDomain(ctx, username, domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, fmt.Errorf("GetAccountByUsernameDomain: error checking database for account %s@%s: %w", username, domain, err)
	}

	if account == nil {
		// Check for failed local lookup.
		if domain == "" {
			return nil, NewErrNotRetrievable(err) // will be db.ErrNoEntries
		}

		// Create and pass-through a new bare-bones model for dereferencing.
		return d.enrichAccount(ctx, requestUser, nil, &gtsmodel.Account{
			ID:       id.NewULID(),
			Username: username,
			Domain:   domain,
		}, false, true)
	}

	// Try to update existing account model
	enriched, err := d.enrichAccount(ctx, requestUser, nil, account, false, block)
	if err != nil {
		log.Errorf("GetAccountByUsernameDomain: error enriching account from remote: %v", err)
		return account, nil // fall back to returning unchanged existing account model
	}

	return enriched, nil
}

func (d *deref) UpdateAccount(ctx context.Context, requestUser string, account *gtsmodel.Account, force bool) (*gtsmodel.Account, error) {
	return d.enrichAccount(ctx, requestUser, nil, account, force, false)
}

// enrichAccount will ensure the given account is the most up-to-date model of the account, re-webfingering and re-dereferencing if necessary.
func (d *deref) enrichAccount(ctx context.Context, requestUser string, uri *url.URL, account *gtsmodel.Account, force, block bool) (*gtsmodel.Account, error) {
	if account.IsLocal() {
		// Can't update local accounts.
		return account, nil
	}

	if !account.CreatedAt.IsZero() && account.IsInstance() {
		// Existing instance account. No need for update.
		return account, nil
	}

	if !force {
		const interval = time.Hour * 48

		// If this account was updated recently (last interval), we return as-is.
		if next := account.FetchedAt.Add(interval); time.Now().Before(next) {
			return account, nil
		}
	}

	if account.Username != "" {
		// A username was provided so we can attempt a webfinger, this ensures up-to-date accountdomain info.
		accDomain, accURI, err := d.fingerRemoteAccount(ctx, requestUser, account.Username, account.Domain)

		if err != nil && account.URI == "" {
			// this is a new account (to us) with username@domain but failed
			// webfinger, there is nothing more we can do in this situation.
			return nil, fmt.Errorf("enrichAccount: error webfingering account: %w", err)
		}

		if err == nil {
			// Update account with latest info.
			account.URI = accURI.String()
			account.Domain = accDomain
			uri = accURI
		}
	}

	if uri == nil {
		var err error

		// No URI provided / found, must parse from account.
		uri, err = url.Parse(account.URI)
		if err != nil {
			return nil, fmt.Errorf("enrichAccount: invalid uri %q: %w", account.URI, err)
		}
	}

	// Check whether this account URI is a blocked domain / subdomain
	if blocked, err := d.db.IsDomainBlocked(ctx, uri.Host); err != nil {
		return nil, newErrDB(fmt.Errorf("enrichAccount: error checking blocked domain: %w", err))
	} else if blocked {
		return nil, fmt.Errorf("enrichAccount: %s is blocked", uri.Host)
	}

	// Mark deref+update handshake start
	d.startHandshake(requestUser, uri)
	defer d.stopHandshake(requestUser, uri)

	// Dereference this account to get the latest available.
	apubAcc, err := d.dereferenceAccountable(ctx, requestUser, uri)
	if err != nil {
		return nil, fmt.Errorf("enrichAccount: error dereferencing account %s: %w", uri, err)
	}

	// Convert the dereferenced AP account object to our GTS model.
	latestAcc, err := d.typeConverter.ASRepresentationToAccount(
		ctx, apubAcc, account.Domain,
	)
	if err != nil {
		return nil, fmt.Errorf("enrichAccount: error converting accountable to gts model for account %s: %w", uri, err)
	}

	if account.Username == "" {
		// No username was provided, so no webfinger was attempted earlier.
		//
		// Now we have a username we can attempt it now, this ensures up-to-date accountdomain info.
		accDomain, _, err := d.fingerRemoteAccount(ctx, requestUser, latestAcc.Username, uri.Host)

		if err == nil {
			// Update account with latest info.
			latestAcc.Domain = accDomain
		}
	}

	// Ensure ID is set and update fetch time.
	latestAcc.ID = account.ID
	latestAcc.FetchedAt = time.Now()

	// Fetch latest account media (TODO: check for changed URI to previous).
	if err = d.fetchRemoteAccountMedia(ctx, latestAcc, requestUser, block); err != nil {
		log.Errorf("error fetching remote media for account %s: %v", uri, err)
	}

	// Fetch the latest remote account emoji IDs used in account display name/bio.
	_, err = d.fetchRemoteAccountEmojis(ctx, latestAcc, requestUser)
	if err != nil {
		log.Errorf("error fetching remote emojis for account %s: %v", uri, err)
	}

	if account.CreatedAt.IsZero() {
		// CreatedAt will be zero if no local copy was
		// found in one of the GetAccountBy___() functions.
		//
		// Set time of creation from the last-fetched date.
		latestAcc.CreatedAt = latestAcc.FetchedAt
		latestAcc.UpdatedAt = latestAcc.FetchedAt

		// This is a new account, we need to place it in the database.
		if err := d.db.PutAccount(ctx, latestAcc); err != nil {
			return nil, fmt.Errorf("enrichAccount: error putting in database: %w", err)
		}
	} else {
		// Set time of update from the last-fetched date.
		latestAcc.UpdatedAt = latestAcc.FetchedAt

		// Use existing account values.
		latestAcc.CreatedAt = account.CreatedAt
		latestAcc.Language = account.Language

		// This is an existing account, update the model in the database.
		if err := d.db.UpdateAccount(ctx, latestAcc); err != nil {
			return nil, fmt.Errorf("enrichAccount: error updating database: %w", err)
		}
	}

	return latestAcc, nil
}

// dereferenceAccountable calls remoteAccountID with a GET request, and tries to parse whatever
// it finds as something that an account model can be constructed out of.
//
// Will work for Person, Application, or Service models.
func (d *deref) dereferenceAccountable(ctx context.Context, username string, remoteAccountID *url.URL) (ap.Accountable, error) {
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

	//nolint shutup linter
	switch t.GetTypeName() {
	case ap.ActorApplication:
		return t.(vocab.ActivityStreamsApplication), nil
	case ap.ActorGroup:
		return t.(vocab.ActivityStreamsGroup), nil
	case ap.ActorOrganization:
		return t.(vocab.ActivityStreamsOrganization), nil
	case ap.ActorPerson:
		return t.(vocab.ActivityStreamsPerson), nil
	case ap.ActorService:
		return t.(vocab.ActivityStreamsService), nil
	}

	return nil, newErrWrongType(fmt.Errorf("DereferenceAccountable: type name %s not supported as Accountable", t.GetTypeName()))
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
func (d *deref) fetchRemoteAccountMedia(ctx context.Context, targetAccount *gtsmodel.Account, requestingUsername string, blocking bool) error {
	// Fetch a transport beforehand for either(or both) avatar / header dereferencing.
	tsport, err := d.transportController.NewTransportForUsername(ctx, requestingUsername)
	if err != nil {
		return fmt.Errorf("fetchRemoteAccountMedia: error getting transport for user: %s", err)
	}

	if targetAccount.AvatarRemoteURL != "" {
		var processingMedia *media.ProcessingMedia

		// Parse the target account's avatar URL into URL object.
		avatarIRI, err := url.Parse(targetAccount.AvatarRemoteURL)
		if err != nil {
			return err
		}

		d.dereferencingAvatarsLock.Lock() // LOCK HERE
		// first check if we're already processing this media
		if alreadyProcessing, ok := d.dereferencingAvatars[targetAccount.ID]; ok {
			// we're already on it, no worries
			processingMedia = alreadyProcessing
		} else {
			data := func(innerCtx context.Context) (io.ReadCloser, int64, error) {
				return tsport.DereferenceMedia(innerCtx, avatarIRI)
			}

			avatar := true
			newProcessing, err := d.mediaManager.ProcessMedia(ctx, data, nil, targetAccount.ID, &media.AdditionalMediaInfo{
				RemoteURL: &targetAccount.AvatarRemoteURL,
				Avatar:    &avatar,
			})
			if err != nil {
				d.dereferencingAvatarsLock.Unlock()
				return err
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
				return err
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
	}

	if targetAccount.HeaderRemoteURL != "" {
		var processingMedia *media.ProcessingMedia

		// Parse the target account's header URL into URL object.
		headerIRI, err := url.Parse(targetAccount.HeaderRemoteURL)
		if err != nil {
			return err
		}

		d.dereferencingHeadersLock.Lock() // LOCK HERE
		// first check if we're already processing this media
		if alreadyProcessing, ok := d.dereferencingHeaders[targetAccount.ID]; ok {
			// we're already on it, no worries
			processingMedia = alreadyProcessing
		} else {
			data := func(innerCtx context.Context) (io.ReadCloser, int64, error) {
				return tsport.DereferenceMedia(innerCtx, headerIRI)
			}

			header := true
			newProcessing, err := d.mediaManager.ProcessMedia(ctx, data, nil, targetAccount.ID, &media.AdditionalMediaInfo{
				RemoteURL: &targetAccount.HeaderRemoteURL,
				Header:    &header,
			})
			if err != nil {
				d.dereferencingAvatarsLock.Unlock()
				return err
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
				return err
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
	}

	return nil
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

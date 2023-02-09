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
	"net/url"
	"sync"

	"codeberg.org/gruf/go-mutexes"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// Dereferencer wraps logic and functionality for doing dereferencing of remote accounts, statuses, etc, from federated instances.
type Dereferencer interface {
	// GetAccountByURI will attempt to fetch an account by its URI, first checking the database and in the case of a remote account will either check the
	// last_fetched (and updating if beyond fetch interval) or dereferencing for the first-time if this remote account has never been encountered before.
	GetAccountByURI(ctx context.Context, requestUser string, uri *url.URL, block bool) (*gtsmodel.Account, error)

	// GetAccountByUsernameDomain will attempt to fetch an account by username@domain, first checking the database and in the case of a remote account will either
	// check the last_fetched (and updating if beyond fetch interval) or dereferencing for the first-time if this remote account has never been encountered before.
	GetAccountByUsernameDomain(ctx context.Context, requestUser string, username string, domain string, block bool) (*gtsmodel.Account, error)

	// UpdateAccount updates the given account if last_fetched is beyond fetch interval (or if force is set). An updated account model is returned, any media fetching is done async.
	UpdateAccount(ctx context.Context, requestUser string, account *gtsmodel.Account, force bool) (*gtsmodel.Account, error)

	GetStatus(ctx context.Context, username string, remoteStatusID *url.URL, refetch, includeParent bool) (*gtsmodel.Status, ap.Statusable, error)

	EnrichRemoteStatus(ctx context.Context, username string, status *gtsmodel.Status, includeParent bool) (*gtsmodel.Status, error)
	GetRemoteInstance(ctx context.Context, username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error)
	DereferenceAnnounce(ctx context.Context, announce *gtsmodel.Status, requestingUsername string) error
	DereferenceThread(ctx context.Context, username string, statusIRI *url.URL, status *gtsmodel.Status, statusable ap.Statusable)

	GetRemoteMedia(ctx context.Context, requestingUsername string, accountID string, remoteURL string, ai *media.AdditionalMediaInfo) (*media.ProcessingMedia, error)
	GetRemoteEmoji(ctx context.Context, requestingUsername string, remoteURL string, shortcode string, domain string, id string, emojiURI string, ai *media.AdditionalEmojiInfo, refresh bool) (*media.ProcessingEmoji, error)

	Handshaking(username string, remoteAccountID *url.URL) bool
}

type deref struct {
	db                  db.DB
	typeConverter       typeutils.TypeConverter
	transportController transport.Controller
	mediaManager        media.Manager
	derefAvatars        map[string]*media.ProcessingMedia
	derefAvatarsMu      mutexes.Mutex
	derefHeaders        map[string]*media.ProcessingMedia
	derefHeadersMu      mutexes.Mutex
	derefEmojis         map[string]*media.ProcessingEmoji
	derefEmojisMu       mutexes.Mutex
	handshakes          map[string][]*url.URL
	handshakeSync       sync.Mutex // mutex to lock/unlock when checking or updating the handshakes map
}

// NewDereferencer returns a Dereferencer initialized with the given parameters.
func NewDereferencer(db db.DB, typeConverter typeutils.TypeConverter, transportController transport.Controller, mediaManager media.Manager) Dereferencer {
	return &deref{
		db:                  db,
		typeConverter:       typeConverter,
		transportController: transportController,
		mediaManager:        mediaManager,
		derefAvatars:        make(map[string]*media.ProcessingMedia),
		derefHeaders:        make(map[string]*media.ProcessingMedia),
		derefEmojis:         make(map[string]*media.ProcessingEmoji),
		handshakes:          make(map[string][]*url.URL),

		// use wrapped mutexes to allow safely deferring unlock
		// even when more granular locks are required (only unlocks once).
		derefAvatarsMu: mutexes.WithSafety(mutexes.New()),
		derefHeadersMu: mutexes.WithSafety(mutexes.New()),
		derefEmojisMu:  mutexes.WithSafety(mutexes.New()),
	}
}

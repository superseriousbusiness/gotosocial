// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package dereferencing

import (
	"context"
	"net/url"
	"sync"

	"codeberg.org/gruf/go-mutexes"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// Dereferencer wraps logic and functionality for doing dereferencing of remote accounts, statuses, etc, from federated instances.
type Dereferencer interface {
	// GetAccountByURI will attempt to fetch an account by its URI, first checking the database and in the case of a remote account will either check the
	// last_fetched (and updating if beyond fetch interval) or dereferencing for the first-time if this remote account has never been encountered before.
	GetAccountByURI(ctx context.Context, requestUser string, uri *url.URL) (*gtsmodel.Account, error)

	// GetAccountByUsernameDomain will attempt to fetch an account by username@domain, first checking the database and in the case of a remote account will either
	// check the last_fetched (and updating if beyond fetch interval) or dereferencing for the first-time if this remote account has never been encountered before.
	GetAccountByUsernameDomain(ctx context.Context, requestUser string, username string, domain string) (*gtsmodel.Account, error)

	// RefreshAccount forces a refresh of the given account by fetching the current/latest state of the account from the remote instance.
	// An updated account model is returned, but not yet inserted/updated in the database; this is the caller's responsibility.
	RefreshAccount(ctx context.Context, requestUser string, accountable ap.Accountable, account *gtsmodel.Account) (*gtsmodel.Account, error)

	// UpdateAccountAsync ...
	UpdateAccountAsync(ctx context.Context, requestUser string, account *gtsmodel.Account, force bool)

	// GetStatusByURI ...
	GetStatusByURI(ctx context.Context, requestUser string, uri *url.URL) (*gtsmodel.Status, ap.Statusable, error)

	// UpdateStatus ...
	UpdateStatus(ctx context.Context, requestUser string, status *gtsmodel.Status, force bool) (*gtsmodel.Status, ap.Statusable, error)

	// UpdateStatusAsync ...
	UpdateStatusAsync(ctx context.Context, requestUser string, status *gtsmodel.Status, force bool)

	GetRemoteInstance(ctx context.Context, username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error)

	DereferenceAnnounce(ctx context.Context, announce *gtsmodel.Status, requestingUsername string) error

	GetRemoteMedia(ctx context.Context, requestingUsername string, accountID string, remoteURL string, ai *media.AdditionalMediaInfo) (*media.ProcessingMedia, error)

	GetRemoteEmoji(ctx context.Context, requestingUsername string, remoteURL string, shortcode string, domain string, id string, emojiURI string, ai *media.AdditionalEmojiInfo, refresh bool) (*media.ProcessingEmoji, error)

	Handshaking(username string, remoteAccountID *url.URL) bool
}

type deref struct {
	state               *state.State
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
func NewDereferencer(state *state.State, typeConverter typeutils.TypeConverter, transportController transport.Controller, mediaManager media.Manager) Dereferencer {
	return &deref{
		state:               state,
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

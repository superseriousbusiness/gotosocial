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
	// GetAccountByURI will attempt to fetch an accounts by its URI, first checking the database. In the case of a newly-met remote model, or a remote model
	// whose last_fetched date is beyond a certain interval, the account will be dereferenced. In the case of dereferencing, some low-priority account information
	// may be enqueued for asynchronous fetching, e.g. featured account statuses (pins). An ActivityPub object indicates the account was dereferenced.
	GetAccountByURI(ctx context.Context, requestUser string, uri *url.URL) (*gtsmodel.Account, ap.Accountable, error)

	// GetAccountByUsernameDomain will attempt to fetch an accounts by its username@domain, first checking the database. In the case of a newly-met remote model,
	// or a remote model whose last_fetched date is beyond a certain interval, the account will be dereferenced. In the case of dereferencing, some low-priority
	// account information may be enqueued for asynchronous fetching, e.g. featured account statuses (pins). An ActivityPub object indicates the account was dereferenced.
	GetAccountByUsernameDomain(ctx context.Context, requestUser string, username string, domain string) (*gtsmodel.Account, ap.Accountable, error)

	// RefreshAccount updates the given account if remote and last_fetched is beyond fetch interval, or if force is set. An updated account model is returned,
	// but in the case of dereferencing, some low-priority account information may be enqueued for asynchronous fetching, e.g. featured account statuses (pins).
	// An ActivityPub object indicates the account was dereferenced (i.e. updated).
	RefreshAccount(ctx context.Context, requestUser string, account *gtsmodel.Account, apubAcc ap.Accountable, force bool) (*gtsmodel.Account, ap.Accountable, error)

	// RefreshAccountAsync enqueues the given account for an asychronous update fetching, if last_fetched is beyond fetch interval, or if forcc is set.
	// This is a more optimized form of manually enqueueing .UpdateAccount() to the federation worker, since it only enqueues update if necessary.
	RefreshAccountAsync(ctx context.Context, requestUser string, account *gtsmodel.Account, apubAcc ap.Accountable, force bool)

	// GetStatusByURI will attempt to fetch a status by its URI, first checking the database. In the case of a newly-met remote model, or a remote model
	// whose last_fetched date is beyond a certain interval, the status will be dereferenced. In the case of dereferencing, some low-priority status information
	// may be enqueued for asynchronous fetching, e.g. dereferencing the remainder of the status thread. An ActivityPub object indicates the status was dereferenced.
	GetStatusByURI(ctx context.Context, requestUser string, uri *url.URL) (*gtsmodel.Status, ap.Statusable, error)

	// RefreshStatus updates the given status if remote and last_fetched is beyond fetch interval, or if force is set. An updated status model is returned,
	// but in the case of dereferencing, some low-priority status information may be enqueued for asynchronous fetching, e.g. dereferencing the remainder of the
	// status thread. An ActivityPub object indicates the status was dereferenced (i.e. updated).
	RefreshStatus(ctx context.Context, requestUser string, status *gtsmodel.Status, apubStatus ap.Statusable, force bool) (*gtsmodel.Status, ap.Statusable, error)

	// RefreshStatusAsync enqueues the given status for an asychronous update fetching, if last_fetched is beyond fetch interval, or if force is set.
	// This is a more optimized form of manually enqueueing .UpdateStatus() to the federation worker, since it only enqueues update if necessary.
	RefreshStatusAsync(ctx context.Context, requestUser string, status *gtsmodel.Status, apubStatus ap.Statusable, force bool)

	// DereferenceStatusAncestors iterates upwards from the given status, using InReplyToURI, to ensure that as many parent statuses as possible are dereferenced.
	DereferenceStatusAncestors(ctx context.Context, requestUser string, status *gtsmodel.Status) error

	// DereferenceStatusDescendents iterates downwards from the given status, using its replies, to ensure that as many children statuses as possible are dereferenced.
	DereferenceStatusDescendants(ctx context.Context, requestUser string, statusIRI *url.URL, parent ap.Statusable) error

	GetRemoteInstance(ctx context.Context, username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error)

	DereferenceAnnounce(ctx context.Context, announce *gtsmodel.Status, requestingUsername string) error

	GetRemoteMedia(ctx context.Context, requestingUsername string, accountID string, remoteURL string, ai *media.AdditionalMediaInfo) (*media.ProcessingMedia, error)

	GetRemoteEmoji(ctx context.Context, requestingUsername string, remoteURL string, shortcode string, domain string, id string, emojiURI string, ai *media.AdditionalEmojiInfo, refresh bool) (*media.ProcessingEmoji, error)

	Handshaking(username string, remoteAccountID *url.URL) bool
}

type deref struct {
	state               *state.State
	converter           *typeutils.Converter
	transportController transport.Controller
	mediaManager        *media.Manager
	derefAvatars        map[string]*media.ProcessingMedia
	derefAvatarsMu      mutexes.Mutex
	derefHeaders        map[string]*media.ProcessingMedia
	derefHeadersMu      mutexes.Mutex
	derefEmojis         map[string]*media.ProcessingEmoji
	derefEmojisMu       mutexes.Mutex
	handshakes          map[string][]*url.URL
	handshakesMu        sync.Mutex // mutex to lock/unlock when checking or updating the handshakes map
}

// NewDereferencer returns a Dereferencer initialized with the given parameters.
func NewDereferencer(state *state.State, converter *typeutils.Converter, transportController transport.Controller, mediaManager *media.Manager) Dereferencer {
	return &deref{
		state:               state,
		converter:           converter,
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

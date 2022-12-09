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
	"net/url"
	"sync"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// Dereferencer wraps logic and functionality for doing dereferencing of remote accounts, statuses, etc, from federated instances.
type Dereferencer interface {
	GetAccount(ctx context.Context, params GetAccountParams) (*gtsmodel.Account, error)
	GetStatus(ctx context.Context, username string, remoteStatusID *url.URL, refetch, includeParent bool) (*gtsmodel.Status, ap.Statusable, error)

	EnrichRemoteStatus(ctx context.Context, username string, status *gtsmodel.Status, includeParent bool) (*gtsmodel.Status, error)
	GetRemoteInstance(ctx context.Context, username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error)
	DereferenceAnnounce(ctx context.Context, announce *gtsmodel.Status, requestingUsername string) error
	DereferenceThread(ctx context.Context, username string, statusIRI *url.URL, status *gtsmodel.Status, statusable ap.Statusable)

	GetRemoteMedia(ctx context.Context, requestingUsername string, accountID string, remoteURL string, ai *media.AdditionalMediaInfo) (*media.ProcessingMedia, error)
	GetRemoteEmoji(ctx context.Context, requestingUsername string, remoteURL string, shortcode string, domain string, id string, emojiURI string, ai *media.AdditionalEmojiInfo, refresh bool) (*media.ProcessingEmoji, error)

	Handshaking(ctx context.Context, username string, remoteAccountID *url.URL) bool
}

type deref struct {
	db                       db.DB
	typeConverter            typeutils.TypeConverter
	transportController      transport.Controller
	mediaManager             media.Manager
	dereferencingAvatars     map[string]*media.ProcessingMedia
	dereferencingAvatarsLock *sync.Mutex
	dereferencingHeaders     map[string]*media.ProcessingMedia
	dereferencingHeadersLock *sync.Mutex
	dereferencingEmojis      map[string]*media.ProcessingEmoji
	dereferencingEmojisLock  *sync.Mutex
	handshakes               map[string][]*url.URL
	handshakeSync            *sync.Mutex // mutex to lock/unlock when checking or updating the handshakes map
}

// NewDereferencer returns a Dereferencer initialized with the given parameters.
func NewDereferencer(db db.DB, typeConverter typeutils.TypeConverter, transportController transport.Controller, mediaManager media.Manager) Dereferencer {
	return &deref{
		db:                       db,
		typeConverter:            typeConverter,
		transportController:      transportController,
		mediaManager:             mediaManager,
		dereferencingAvatars:     make(map[string]*media.ProcessingMedia),
		dereferencingAvatarsLock: &sync.Mutex{},
		dereferencingHeaders:     make(map[string]*media.ProcessingMedia),
		dereferencingHeadersLock: &sync.Mutex{},
		dereferencingEmojis:      make(map[string]*media.ProcessingEmoji),
		dereferencingEmojisLock:  &sync.Mutex{},
		handshakeSync:            &sync.Mutex{},
	}
}

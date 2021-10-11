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
	"net/url"
	"sync"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// Dereferencer wraps logic and functionality for doing dereferencing of remote accounts, statuses, etc, from federated instances.
type Dereferencer interface {
	GetRemoteAccount(ctx context.Context, username string, remoteAccountID *url.URL, refresh bool) (*gtsmodel.Account, bool, error)
	EnrichRemoteAccount(ctx context.Context, username string, account *gtsmodel.Account) (*gtsmodel.Account, error)

	GetRemoteStatus(ctx context.Context, username string, remoteStatusID *url.URL, refresh, includeParent bool) (*gtsmodel.Status, ap.Statusable, bool, error)
	EnrichRemoteStatus(ctx context.Context, username string, status *gtsmodel.Status, includeParent bool) (*gtsmodel.Status, error)

	GetRemoteInstance(ctx context.Context, username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error)

	// GetRemoteAttachment takes a minimal attachment struct and converts it into a fully fleshed out attachment, stored in the database and instance storage.
	//
	// The parameter minAttachment must have at least the following fields defined:
	//   * minAttachment.RemoteURL
	//   * minAttachment.AccountID
	//   * minAttachment.File.ContentType
	//
	// The returned attachment will have an ID generated for it, so no need to generate one beforehand.
	// A blurhash will also be generated for the attachment.
	//
	// Most other fields will be preserved on the passed attachment, including:
	//   * minAttachment.StatusID
	//   * minAttachment.CreatedAt
	//   * minAttachment.UpdatedAt
	//   * minAttachment.FileMeta
	//   * minAttachment.AccountID
	//   * minAttachment.Description
	//   * minAttachment.ScheduledStatusID
	//   * minAttachment.Thumbnail.RemoteURL
	//   * minAttachment.Avatar
	//   * minAttachment.Header
	//
	// GetRemoteAttachment will return early if an attachment with the same value as minAttachment.RemoteURL
	// is found in the database -- then that attachment will be returned and nothing else will be changed or stored.
	GetRemoteAttachment(ctx context.Context, requestingUsername string, minAttachment *gtsmodel.MediaAttachment) (*gtsmodel.MediaAttachment, error)
	// RefreshAttachment is like GetRemoteAttachment, but the attachment will always be dereferenced again,
	// whether or not it was already stored in the database.
	RefreshAttachment(ctx context.Context, requestingUsername string, minAttachment *gtsmodel.MediaAttachment) (*gtsmodel.MediaAttachment, error)

	DereferenceAnnounce(ctx context.Context, announce *gtsmodel.Status, requestingUsername string) error
	DereferenceThread(ctx context.Context, username string, statusIRI *url.URL) error

	Handshaking(ctx context.Context, username string, remoteAccountID *url.URL) bool
}

type deref struct {
	db                  db.DB
	typeConverter       typeutils.TypeConverter
	transportController transport.Controller
	mediaHandler        media.Handler
	config              *config.Config
	handshakes          map[string][]*url.URL
	handshakeSync       *sync.Mutex // mutex to lock/unlock when checking or updating the handshakes map
}

// NewDereferencer returns a Dereferencer initialized with the given parameters.
func NewDereferencer(config *config.Config, db db.DB, typeConverter typeutils.TypeConverter, transportController transport.Controller, mediaHandler media.Handler) Dereferencer {
	return &deref{
		db:                  db,
		typeConverter:       typeConverter,
		transportController: transportController,
		mediaHandler:        mediaHandler,
		config:              config,
		handshakeSync:       &sync.Mutex{},
	}
}

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

	"github.com/sirupsen/logrus"
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

	GetRemoteStatus(ctx context.Context, username string, remoteStatusID *url.URL, refresh bool) (*gtsmodel.Status, ap.Statusable, bool, error)
	EnrichRemoteStatus(ctx context.Context, username string, status *gtsmodel.Status) (*gtsmodel.Status, error)

	GetRemoteInstance(ctx context.Context, username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error)

	GetRemoteAttachment(ctx context.Context, username string, remoteAttachmentURI *url.URL, ownerAccountID string, statusID string, expectedContentType string) (*gtsmodel.MediaAttachment, error)
	RefreshAttachment(ctx context.Context, requestingUsername string, remoteAttachmentURI *url.URL, ownerAccountID string, expectedContentType string) (*gtsmodel.MediaAttachment, error)

	DereferenceAnnounce(ctx context.Context, announce *gtsmodel.Status, requestingUsername string) error
	DereferenceThread(ctx context.Context, username string, statusIRI *url.URL) error

	Handshaking(ctx context.Context, username string, remoteAccountID *url.URL) bool
}

type deref struct {
	log                 *logrus.Logger
	db                  db.DB
	typeConverter       typeutils.TypeConverter
	transportController transport.Controller
	mediaHandler        media.Handler
	config              *config.Config
	handshakes          map[string][]*url.URL
	handshakeSync       *sync.Mutex // mutex to lock/unlock when checking or updating the handshakes map
}

// NewDereferencer returns a Dereferencer initialized with the given parameters.
func NewDereferencer(config *config.Config, db db.DB, typeConverter typeutils.TypeConverter, transportController transport.Controller, mediaHandler media.Handler, log *logrus.Logger) Dereferencer {
	return &deref{
		log:                 log,
		db:                  db,
		typeConverter:       typeConverter,
		transportController: transportController,
		mediaHandler:        mediaHandler,
		config:              config,
		handshakeSync:       &sync.Mutex{},
	}
}

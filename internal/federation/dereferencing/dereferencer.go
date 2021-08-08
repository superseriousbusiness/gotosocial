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

type Dereferencer interface {
	GetRemoteAccount(username string, remoteAccountID *url.URL, refresh bool) (*gtsmodel.Account, bool, error)
	EnrichRemoteAccount(username string, account *gtsmodel.Account) (*gtsmodel.Account, error)

	GetRemoteStatus(username string, remoteStatusID *url.URL, refresh bool) (*gtsmodel.Status, ap.Statusable, bool, error)
	EnrichRemoteStatus(username string, status *gtsmodel.Status) (*gtsmodel.Status, error)

	GetRemoteInstance(username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error)

	DereferenceAnnounce(announce *gtsmodel.Status, requestingUsername string) error
	DereferenceThread(username string, statusIRI *url.URL) error

	Handshaking(username string, remoteAccountID *url.URL) bool
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

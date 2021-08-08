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

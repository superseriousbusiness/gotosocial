package dereferencing

import (
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

type Dereferencer interface {
	DereferenceAccountable(username string, remoteAccountID *url.URL) (typeutils.Accountable, error)
	DereferenceStatusable(username string, remoteStatusID *url.URL) (typeutils.Statusable, error)


	PopulateAccountFields(account *gtsmodel.Account, requestingUsername string, refresh bool) error

	DereferenceAnnounce(announce *gtsmodel.Status, requestingUsername string) error

	DereferenceCollectionPage(username string, pageIRI *url.URL) (typeutils.CollectionPageable, error)
	DereferenceRemoteInstance(username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error)
	FullyDereferenceStatusableAndAccount(username string, statusable typeutils.Statusable) error
	DereferenceThread(username string, statusIRI *url.URL) error

	PopulateStatusFields(status *gtsmodel.Status, requestingUsername string) error

}

type deref struct {
	log                 *logrus.Logger
	db                  db.DB
	typeConverter       typeutils.TypeConverter
	transportController transport.Controller
	mediaHandler        media.Handler
	config              *config.Config
}

func NewDereferencer(config *config.Config, db db.DB, typeConverter typeutils.TypeConverter, transportController transport.Controller, mediaHandler media.Handler, log *logrus.Logger) Dereferencer {
	return &deref{
		log:                 log,
		db:                  db,
		typeConverter:       typeConverter,
		transportController: transportController,
		mediaHandler:        mediaHandler,
		config:              config,
	}
}

package fileserver

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
)

// fileServer implements the RESTAPIModule interface.
// The goal here is to serve requested media files if the gotosocial server is configured to use local storage.
type fileServer struct {
	config      *config.Config
	db          db.DB
	storage     storage.Storage
	log         *logrus.Logger
	storageBase string
}

// New returns a new fileServer module
func New(config *config.Config, db db.DB, storage storage.Storage, log *logrus.Logger) apimodule.ClientAPIModule {

	storageBase := fmt.Sprintf("%s", config.StorageConfig.BasePath) // TODO: do this properly

	return &fileServer{
		config:      config,
		db:          db,
		storage:     storage,
		log:         log,
		storageBase: storageBase,
	}
}

// Route satisfies the RESTAPIModule interface
func (m *fileServer) Route(s router.Router) error {
	// s.AttachHandler(http.MethodPost, appsPath, m.appsPOSTHandler)
	return nil
}

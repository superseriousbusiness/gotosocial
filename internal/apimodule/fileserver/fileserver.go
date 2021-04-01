package fileserver

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
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

	storageBase := config.StorageConfig.BasePath // TODO: do this properly

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

func (m *fileServer) CreateTables(db db.DB) error {
	models := []interface{}{
		&model.User{},
		&model.Account{},
		&model.Follow{},
		&model.FollowRequest{},
		&model.Status{},
		&model.Application{},
		&model.EmailDomainBlock{},
		&model.MediaAttachment{},
	}

	for _, m := range models {
		if err := db.CreateTable(m); err != nil {
			return fmt.Errorf("error creating table: %s", err)
		}
	}
	return nil
}

package testrig

import (
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
)

// NewTestStorage returns a new in memory storage with the given config
func NewTestStorage(c *config.Config, log *logrus.Logger) storage.Storage {
	s, err := storage.NewInMem(c, log)
	if err != nil {
		panic(err)
	}
	return s
}

func StandardStorageSetup(s storage.Storage) {

}

func StandardStorageTeardown(s storage.Storage) {

}

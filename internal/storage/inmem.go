package storage

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// NewInMem returns an in-memory implementation of the Storage interface.
// This is good for testing and whatnot but ***SHOULD ABSOLUTELY NOT EVER
// BE USED IN A PRODUCTION SETTING***, because A) everything will be wiped out
// if you restart the server and B) if you store lots of images your RAM use
// will absolutely go through the roof.
func NewInMem(c *config.Config, log *logrus.Logger) (Storage, error) {
	return &inMemStorage{
		stored: make(map[string][]byte),
		log:    log,
	}, nil
}

type inMemStorage struct {
	stored map[string][]byte
	log    *logrus.Logger
}

func (s *inMemStorage) StoreFileAt(path string, data []byte) error {
	l := s.log.WithField("func", "StoreFileAt")
	l.Debugf("storing at path %s", path)
	s.stored[path] = data
	return nil
}

func (s *inMemStorage) RetrieveFileFrom(path string) ([]byte, error) {
	l := s.log.WithField("func", "RetrieveFileFrom")
	l.Debugf("retrieving from path %s", path)
	d, ok := s.stored[path]
	if !ok {
		return nil, fmt.Errorf("no data found at path %s", path)
	}
	return d, nil
}

func (s *inMemStorage) ListKeys() ([]string, error) {
	keys := []string{}
	for k := range s.stored {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *inMemStorage) RemoveFileAt(path string) error {
	delete(s.stored, path)
	return nil
}

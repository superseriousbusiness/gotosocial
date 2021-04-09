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
	}, nil
}

type inMemStorage struct {
	stored map[string][]byte
}

func (s *inMemStorage) StoreFileAt(path string, data []byte) error {
	s.stored[path] = data
	return nil
}

func (s *inMemStorage) RetrieveFileFrom(path string) ([]byte, error) {
	d, ok := s.stored[path]
	if !ok {
		return nil, fmt.Errorf("no data found at path %s", path)
	}
	return d, nil
}

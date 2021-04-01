package storage

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

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

package storage

import (
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

func NewLocal(c *config.Config, log *logrus.Logger) (Storage, error) {
	return &localStorage{}, nil
}

type localStorage struct {
}

func (s *localStorage) StoreFileAt(path string, data []byte) error {
	return nil
}

func (s *localStorage) RetrieveFileFrom(path string) ([]byte, error) {
	return nil, nil
}

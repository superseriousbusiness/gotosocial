package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// NewLocal returns an implementation of the Storage interface that uses
// the local filesystem for storing and retrieving files, attachments, etc.
func NewLocal(c *config.Config, log *logrus.Logger) (Storage, error) {
	return &localStorage{
		config: c,
		log:    log,
	}, nil
}

type localStorage struct {
	config *config.Config
	log    *logrus.Logger
}

func (s *localStorage) StoreFileAt(path string, data []byte) error {
	l := s.log.WithField("func", "StoreFileAt")
	l.Debugf("storing at path %s", path)
	components := strings.Split(path, "/")
	dir := strings.Join(components[0:len(components) - 1], "/")
	if err := os.MkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("error writing file at %s: %s", path, err)
	}
	if err := os.WriteFile(path, data, 0777); err != nil {
		return fmt.Errorf("error writing file at %s: %s", path, err)
	}
	return nil
}

func (s *localStorage) RetrieveFileFrom(path string) ([]byte, error) {
	l := s.log.WithField("func", "RetrieveFileFrom")
	l.Debugf("retrieving from path %s", path)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file at %s: %s", path, err)
	}
	return b, nil
}

func (s *localStorage) ListKeys() ([]string, error) {
	keys := []string{}
	err := filepath.Walk(s.config.StorageConfig.BasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			keys = append(keys, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *localStorage) RemoveFileAt(path string) error {
	return os.Remove(path)
}

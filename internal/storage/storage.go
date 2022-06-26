package storage

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"

	"codeberg.org/gruf/go-store/kv"
	"codeberg.org/gruf/go-store/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

var (
	ErrNotSupported = errors.New("driver does not suppport functionality")
)

// Driver implements the functionality to store and retrieve blobs
// (images,video,audio)
type Driver interface {
	Get(key string) ([]byte, error)
	GetStream(key string) (io.ReadCloser, error)
	PutStream(key string, r io.Reader) error
	Put(key string, value []byte) error
	Delete(key string) error
	URL(key string) *url.URL
}

func AutoConfig() (Driver, error) {
	switch config.GetStorageBackend() {
	case "s3":
		mc, err := minio.New(config.GetStorageS3Endpoint(), &minio.Options{
			Creds:  credentials.NewStaticV4(config.GetStorageS3AccessKey(), config.GetStorageS3SecretKey(), ""),
			Secure: config.GetStorageS3UseSSL(),
		})
		if err != nil {
			return nil, fmt.Errorf("creating minio client: %w", err)
		}
		return NewS3(mc, config.GetStorageS3BucketName()), nil
	case "local":
		storageBasePath := config.GetStorageLocalBasePath()
		storage, err := kv.OpenFile(storageBasePath, &storage.DiskConfig{
			// Put the store lockfile in the storage dir itself.
			// Normally this would not be safe, since we could end up
			// overwriting the lockfile if we store a file called 'store.lock'.
			// However, in this case it's OK because the keys are set by
			// GtS and not the user, so we know we're never going to overwrite it.
			LockFile: path.Join(storageBasePath, "store.lock"),
		})
		if err != nil {
			return nil, fmt.Errorf("error creating storage backend: %s", err)
		}
		return &Local{KVStore: storage}, nil
	}
	return nil, fmt.Errorf("invalid storage backend %s", config.GetStorageBackend())
}

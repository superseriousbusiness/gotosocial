/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"

	"codeberg.org/gruf/go-store/v2/kv"
	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

var (
	ErrNotSupported  = errors.New("driver does not suppport functionality")
	ErrAlreadyExists = storage.ErrAlreadyExists
)

// Driver implements the functionality to store and retrieve blobs
// (images,video,audio)
type Driver interface {
	Get(ctx context.Context, key string) ([]byte, error)
	GetStream(ctx context.Context, key string) (io.ReadCloser, error)
	PutStream(ctx context.Context, key string, r io.Reader) error
	Put(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error
	URL(ctx context.Context, key string) *url.URL
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
		return NewS3(
			mc,
			config.GetStorageS3BucketName(),
			config.GetStorageS3Proxy(),
		), nil
	case "local":
		basePath := config.GetStorageLocalBasePath()
		disk, err := storage.OpenDisk(basePath, &storage.DiskConfig{
			// Put the store lockfile in the storage dir itself.
			// Normally this would not be safe, since we could end up
			// overwriting the lockfile if we store a file called 'store.lock'.
			// However, in this case it's OK because the keys are set by
			// GtS and not the user, so we know we're never going to overwrite it.
			LockFile: path.Join(basePath, "store.lock"),
		})
		if err != nil {
			return nil, fmt.Errorf("error openingdisk storage: %v", err)
		}
		return &Local{kv.New(disk)}, nil
	}
	return nil, fmt.Errorf("invalid storage backend %s", config.GetStorageBackend())
}

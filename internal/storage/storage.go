/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"fmt"
	"mime"
	"net/url"
	"path"
	"time"

	"codeberg.org/gruf/go-cache/v3/ttl"
	"codeberg.org/gruf/go-store/v2/kv"
	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

const (
	urlCacheTTL             = time.Hour * 24
	urlCacheExpiryFrequency = time.Minute * 5
)

// ErrAlreadyExists is a ptr to underlying storage.ErrAlreadyExists,
// to put the related errors in the same package as our storage wrapper.
var ErrAlreadyExists = storage.ErrAlreadyExists

// Driver wraps a kv.KVStore to also provide S3 presigned GET URLs.
type Driver struct {
	// Underlying storage
	*kv.KVStore
	Storage storage.Storage

	// S3-only parameters
	Proxy          bool
	Bucket         string
	PresignedCache *ttl.Cache[string, *url.URL]
}

// URL will return a presigned GET object URL, but only if running on S3 storage with proxying disabled.
func (d *Driver) URL(ctx context.Context, key string) *url.URL {
	// Check whether S3 *without* proxying is enabled
	s3, ok := d.Storage.(*storage.S3Storage)
	if !ok || d.Proxy {
		return nil
	}

	// access the cache member directly to avoid extending the TTL
	if u, ok := d.PresignedCache.Cache.Get(key); ok {
		return u.Value
	}

	u, err := s3.Client().PresignedGetObject(ctx, d.Bucket, key, urlCacheTTL, url.Values{
		"response-content-type": []string{mime.TypeByExtension(path.Ext(key))},
	})
	if err != nil {
		// If URL request fails, fallback is to fetch the file. So ignore the error here
		return nil
	}
	d.PresignedCache.Set(key, u)
	return u
}

func AutoConfig() (*Driver, error) {
	switch backend := config.GetStorageBackend(); backend {
	case "s3":
		return NewS3Storage()
	case "local":
		return NewFileStorage()
	default:
		return nil, fmt.Errorf("invalid storage backend: %s", backend)
	}

}

func NewFileStorage() (*Driver, error) {
	// Load runtime configuration
	basePath := config.GetStorageLocalBasePath()

	// Open the disk storage implementation
	disk, err := storage.OpenDisk(basePath, &storage.DiskConfig{
		// Put the store lockfile in the storage dir itself.
		// Normally this would not be safe, since we could end up
		// overwriting the lockfile if we store a file called 'store.lock'.
		// However, in this case it's OK because the keys are set by
		// GtS and not the user, so we know we're never going to overwrite it.
		LockFile: path.Join(basePath, "store.lock"),
	})
	if err != nil {
		return nil, fmt.Errorf("error opening disk storage: %w", err)
	}

	return &Driver{
		KVStore: kv.New(disk),
		Storage: disk,
	}, nil
}

func NewS3Storage() (*Driver, error) {
	// Load runtime configuration
	endpoint := config.GetStorageS3Endpoint()
	access := config.GetStorageS3AccessKey()
	secret := config.GetStorageS3SecretKey()
	secure := config.GetStorageS3UseSSL()
	bucket := config.GetStorageS3BucketName()

	// Open the s3 storage implementation
	s3, err := storage.OpenS3(endpoint, bucket, &storage.S3Config{
		CoreOpts: minio.Options{
			Creds:  credentials.NewStaticV4(access, secret, ""),
			Secure: secure,
		},
		GetOpts:      minio.GetObjectOptions{},
		PutOpts:      minio.PutObjectOptions{},
		PutChunkSize: 5 * 1024 * 1024, // 5MiB
		StatOpts:     minio.StatObjectOptions{},
		RemoveOpts:   minio.RemoveObjectOptions{},
		ListSize:     200,
	})
	if err != nil {
		return nil, fmt.Errorf("error opening s3 storage: %w", err)
	}

	// ttl should be lower than the expiry used by S3 to avoid serving invalid URLs
	presignedCache := ttl.New[string, *url.URL](0, 1000, urlCacheTTL-urlCacheExpiryFrequency)
	presignedCache.Start(urlCacheExpiryFrequency)

	return &Driver{
		KVStore:        kv.New(s3),
		Proxy:          config.GetStorageS3Proxy(),
		Bucket:         config.GetStorageS3BucketName(),
		Storage:        s3,
		PresignedCache: presignedCache,
	}, nil
}

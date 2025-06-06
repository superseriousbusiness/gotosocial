// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"path"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-cache/v3/ttl"
	"codeberg.org/gruf/go-storage"
	"codeberg.org/gruf/go-storage/disk"
	"codeberg.org/gruf/go-storage/s3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	urlCacheTTL             = time.Hour * 24
	urlCacheExpiryFrequency = time.Minute * 5
)

// PresignedURL represents a pre signed S3 URL with
// an expiry time.
type PresignedURL struct {
	*url.URL
	Expiry time.Time // link expires at this time
}

// IsInvalidKey returns whether error is an invalid-key
// type error returned by the underlying storage library.
func IsInvalidKey(err error) bool {
	return errors.Is(err, storage.ErrInvalidKey)
}

// IsAlreadyExist returns whether error is an already-exists
// type error returned by the underlying storage library.
func IsAlreadyExist(err error) bool {
	return errors.Is(err, storage.ErrAlreadyExists)
}

// IsNotFound returns whether error is a not-found error
// type returned by the underlying storage library.
func IsNotFound(err error) bool {
	return errors.Is(err, storage.ErrNotFound)
}

// Driver wraps a kv.KVStore to also provide S3 presigned GET URLs.
type Driver struct {
	// Underlying storage
	Storage storage.Storage

	// S3-only parameters
	Proxy          bool
	Bucket         string
	PresignedCache *ttl.Cache[string, PresignedURL]
	RedirectURL    string
}

// Get returns the byte value for key in storage.
func (d *Driver) Get(ctx context.Context, key string) ([]byte, error) {
	return d.Storage.ReadBytes(ctx, key)
}

// GetStream returns an io.ReadCloser for the value bytes at key in the storage.
func (d *Driver) GetStream(ctx context.Context, key string) (io.ReadCloser, error) {
	return d.Storage.ReadStream(ctx, key)
}

// Put writes the supplied value bytes at key in the storage
func (d *Driver) Put(ctx context.Context, key string, value []byte) (int, error) {
	return d.Storage.WriteBytes(ctx, key, value)
}

// PutFile moves the contents of file at path, to storage.Driver{} under given key (with content-type if supported).
func (d *Driver) PutFile(ctx context.Context, key, filepath, contentType string) (int64, error) {

	// Open file at path for reading.
	file, err := os.Open(filepath)
	if err != nil {
		return 0, gtserror.Newf("error opening file %s: %w", filepath, err)
	}

	var sz int64

	switch d := d.Storage.(type) {
	case *s3.S3Storage:
		var info minio.UploadInfo

		// For S3 storage, write the file but specifically pass in the
		// content-type as an extra option. This handles the case of media
		// being served via CDN redirect (where we don't handle content-type).
		info, err = d.PutObject(ctx, key, file, minio.PutObjectOptions{
			ContentType: contentType,
		})

		// Get size from
		// uploaded info.
		sz = info.Size

	default:
		// Write the file data to storage under key. Note
		// that for disk.DiskStorage{} this should end up
		// being a highly optimized Linux sendfile syscall.
		sz, err = d.WriteStream(ctx, key, file)
	}

	// Wrap write error.
	if err != nil {
		err = gtserror.Newf("error writing file %s: %w", key, err)
	}

	// Close the file: done with it.
	if e := file.Close(); e != nil {
		log.Errorf(ctx, "error closing file %s: %v", filepath, e)
	}

	return sz, err
}

// Delete attempts to remove the supplied key (and corresponding value) from storage.
func (d *Driver) Delete(ctx context.Context, key string) error {
	return d.Storage.Remove(ctx, key)
}

// Has checks if the supplied key is in the storage.
func (d *Driver) Has(ctx context.Context, key string) (bool, error) {
	stat, err := d.Storage.Stat(ctx, key)
	return (stat != nil), err
}

// WalkKeys walks the keys in the storage.
func (d *Driver) WalkKeys(ctx context.Context, walk func(string) error) error {
	return d.Storage.WalkKeys(ctx, storage.WalkKeysOpts{
		Step: func(entry storage.Entry) error {
			return walk(entry.Key)
		},
	})
}

// URL will return a presigned GET object URL, but only if running on S3 storage with proxying disabled.
func (d *Driver) URL(ctx context.Context, key string) *PresignedURL {

	// Check whether S3 *without* proxying is enabled
	s3, ok := d.Storage.(*s3.S3Storage)
	if !ok || d.Proxy {
		return nil
	}

	// Check cache underlying cache map directly to
	// avoid extending the TTL (which cache.Get() does).
	d.PresignedCache.Lock()
	e, ok := d.PresignedCache.Cache.Get(key)
	d.PresignedCache.Unlock()

	if ok {
		return &e.Value
	}

	var (
		u   *url.URL
		err error
	)

	if d.RedirectURL != "" {
		u, err = url.Parse(d.RedirectURL + "/" + key)
		if err != nil {
			// If URL parsing fails, fallback is to
			// fetch the file. So ignore the error here
			return nil
		}
	} else {
		u, err = s3.Client().PresignedGetObject(ctx, d.Bucket, key, urlCacheTTL, url.Values{
			"response-content-type": []string{mime.TypeByExtension(path.Ext(key))},
		})
		if err != nil {
			// If URL request fails, fallback is to
			// fetch the file. So ignore the error here
			return nil
		}
	}

	psu := PresignedURL{
		URL:    u,
		Expiry: time.Now().Add(urlCacheTTL), // link expires in 24h time
	}

	d.PresignedCache.Set(key, psu)
	return &psu
}

// ProbeCSPUri returns a URI string that can be added
// to a content-security-policy to allow requests to
// endpoints served by this driver.
//
// If the driver is not backed by non-proxying S3,
// this will return an empty string and no error.
//
// Otherwise, this function probes for a CSP URI by
// doing the following:
//
//  1. Create a temporary file in the S3 bucket.
//  2. Generate a pre-signed URL for that file.
//  3. Extract '[scheme]://[host]' from the URL.
//  4. Remove the temporary file.
//  5. Return the '[scheme]://[host]' string.
func (d *Driver) ProbeCSPUri(ctx context.Context) (string, error) {
	// Check whether S3 without proxying
	// is enabled. If it's not, there's
	// no need to add anything to the CSP.
	s3, ok := d.Storage.(*s3.S3Storage)
	if !ok || d.Proxy {
		return "", nil
	}

	// If an S3 redirect URL is set, just
	// return this URL without probing; we
	// likely don't have write access on it
	// anyway since it's probs a CDN bucket.
	if d.RedirectURL != "" {
		return d.RedirectURL + "/", nil
	}

	const cspKey = "gotosocial-csp-probe"

	// Create an empty file in S3 storage.
	if _, err := d.Put(ctx, cspKey, make([]byte, 0)); err != nil {
		return "", gtserror.Newf("error putting file in bucket at key %s: %w", cspKey, err)
	}

	// Try to clean up file whatever happens.
	defer func() {
		if err := d.Delete(ctx, cspKey); err != nil {
			log.Warnf(ctx, "error deleting file from bucket at key %s (%v); "+
				"you may want to remove this file manually from your S3 bucket", cspKey, err)
		}
	}()

	// Get a presigned URL for that empty file.
	u, err := s3.Client().PresignedGetObject(ctx, d.Bucket, cspKey, 1*time.Second, nil)
	if err != nil {
		return "", err
	}

	// Create a stripped version of the presigned
	// URL that includes only the host and scheme.
	uStripped := &url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
	}

	return uStripped.String(), nil
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

	// Use default disk config with
	// increased write buffer size.
	diskCfg := disk.DefaultConfig()
	diskCfg.WriteBufSize = int(16 * bytesize.KiB)

	// Open the disk storage implementation
	disk, err := disk.Open(basePath, &diskCfg)
	if err != nil {
		return nil, fmt.Errorf("error opening disk storage: %w", err)
	}

	return &Driver{
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
	redirectURL := config.GetStorageS3RedirectURL()

	var bucketLookup minio.BucketLookupType
	switch s := config.GetStorageS3BucketLookup(); s {
	case "auto":
		bucketLookup = minio.BucketLookupAuto
	case "dns":
		bucketLookup = minio.BucketLookupDNS
	case "path":
		bucketLookup = minio.BucketLookupPath
	default:
		log.Warnf(nil, "%s set to %s which is not recognized, defaulting to 'auto'", config.StorageS3BucketLookupFlag(), s)
		bucketLookup = minio.BucketLookupAuto
	}

	// Open the s3 storage implementation
	s3, err := s3.Open(endpoint, bucket, &s3.Config{
		KeyPrefix: config.GetStorageS3KeyPrefix(),
		CoreOpts: minio.Options{
			Creds:        credentials.NewStaticV4(access, secret, ""),
			Secure:       secure,
			BucketLookup: bucketLookup,
		},
		PutChunkSize: 5 * 1024 * 1024, // 5MiB
		ListSize:     200,
	})
	if err != nil {
		return nil, fmt.Errorf("error opening s3 storage: %w", err)
	}

	// ttl should be lower than the expiry used by S3 to avoid serving invalid URLs
	presignedCache := ttl.New[string, PresignedURL](0, 1000, urlCacheTTL-urlCacheExpiryFrequency)
	presignedCache.Start(urlCacheExpiryFrequency)

	return &Driver{
		Proxy:          config.GetStorageS3Proxy(),
		Bucket:         config.GetStorageS3BucketName(),
		Storage:        s3,
		PresignedCache: presignedCache,
		RedirectURL:    redirectURL,
	}, nil
}

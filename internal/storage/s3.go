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
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/url"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
)

type S3 struct {
	mc     *minio.Client
	bucket string
}

func NewS3(mc *minio.Client, bucket string) *S3 {
	return &S3{
		mc:     mc,
		bucket: bucket,
	}
}

func (s *S3) Get(key string) ([]byte, error) {
	r, err := s.GetStream(key)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading data from s3: %w", err)
	}
	return b, nil
}
func (s *S3) GetStream(key string) (io.ReadCloser, error) {
	o, err := s.mc.GetObject(context.TODO(), s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		err = fmt.Errorf("retrieving object from s3: %w", err)
	}
	return o, err
}
func (s *S3) PutStream(key string, r io.Reader) error {
	if _, err := s.mc.PutObject(context.TODO(), s.bucket, key, r, -1, minio.PutObjectOptions{}); err != nil {
		return fmt.Errorf("uploading data stream: %w", err)
	}
	return nil
}
func (s *S3) Put(key string, value []byte) error {
	if _, err := s.mc.PutObject(context.TODO(), s.bucket, key, bytes.NewBuffer(value), -1, minio.PutObjectOptions{}); err != nil {
		return fmt.Errorf("uploading data slice: %w", err)
	}
	return nil
}
func (s *S3) Delete(key string) error {
	return s.mc.RemoveObject(context.TODO(), s.bucket, key, minio.RemoveObjectOptions{})
}
func (s *S3) URL(key string) *url.URL {
	// it's safe to ignore the error here, as we just fall back to fetching the
	// file if the url request fails
	url, _ := s.mc.PresignedGetObject(context.TODO(), s.bucket, key, time.Hour, url.Values{
		"response-content-type": []string{mime.TypeByExtension(path.Ext(key))},
	})
	return url
}

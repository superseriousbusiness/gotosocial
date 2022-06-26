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

type s3 struct {
	mc     *minio.Client
	bucket string
}

func NewS3(mc *minio.Client, bucket string) *s3 {
	return &s3{
		mc:     mc,
		bucket: bucket,
	}
}

func (s *s3) Get(key string) ([]byte, error) {
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
func (s *s3) GetStream(key string) (io.ReadCloser, error) {
	o, err := s.mc.GetObject(context.TODO(), s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		err = fmt.Errorf("retrieving object from s3: %w", err)
	}
	return o, err
}
func (s *s3) PutStream(key string, r io.Reader) error {
	if _, err := s.mc.PutObject(context.TODO(), s.bucket, key, r, -1, minio.PutObjectOptions{}); err != nil {
		return fmt.Errorf("uploading data stream: %w", err)
	}
	return nil
}
func (s *s3) Put(key string, value []byte) error {
	if _, err := s.mc.PutObject(context.TODO(), s.bucket, key, bytes.NewBuffer(value), -1, minio.PutObjectOptions{}); err != nil {
		return fmt.Errorf("uploading data slice: %w", err)
	}
	return nil
}
func (s *s3) Delete(key string) error {
	return s.mc.RemoveObject(context.TODO(), s.bucket, key, minio.RemoveObjectOptions{})
}
func (s *s3) URL(key string) *url.URL {
	// it's safe to ignore the error here, as we just fall back to fetching the
	// file if the url request fails
	url, _ := s.mc.PresignedGetObject(context.TODO(), s.bucket, key, time.Hour, url.Values{
		"response-content-type": []string{mime.TypeByExtension(path.Ext(key))},
	})
	return url
}

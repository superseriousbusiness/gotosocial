package storage

import (
	"bytes"
	"context"
	"io"
	"sync/atomic"

	"codeberg.org/gruf/go-store/v2/util"
	"github.com/minio/minio-go/v7"
)

// DefaultS3Config is the default S3Storage configuration.
var DefaultS3Config = &S3Config{
	CoreOpts:     minio.Options{},
	GetOpts:      minio.GetObjectOptions{},
	PutOpts:      minio.PutObjectOptions{},
	PutChunkSize: 4 * 1024 * 1024, // 4MiB
	StatOpts:     minio.StatObjectOptions{},
	RemoveOpts:   minio.RemoveObjectOptions{},
	ListSize:     200,
}

// S3Config defines options to be used when opening an S3Storage,
// mostly options for underlying S3 client library.
type S3Config struct {
	// CoreOpts are S3 client options passed during initialization.
	CoreOpts minio.Options

	// GetOpts are S3 client options passed during .Read___() calls.
	GetOpts minio.GetObjectOptions

	// PutOpts are S3 client options passed during .Write___() calls.
	PutOpts minio.PutObjectOptions

	// PutChunkSize is the chunk size (in bytes) to use when sending
	// a byte stream reader of unknown size as a multi-part object.
	PutChunkSize int64

	// StatOpts are S3 client options passed during .Stat() calls.
	StatOpts minio.StatObjectOptions

	// RemoveOpts are S3 client options passed during .Remove() calls.
	RemoveOpts minio.RemoveObjectOptions

	// ListSize determines how many items to include in each
	// list request, made during calls to .WalkKeys().
	ListSize int
}

// getS3Config returns a valid S3Config for supplied ptr.
func getS3Config(cfg *S3Config) S3Config {
	const minChunkSz = 5 * 1024 * 1024

	// If nil, use default
	if cfg == nil {
		cfg = DefaultS3Config
	}

	// Ensure a minimum compatible chunk size
	if cfg.PutChunkSize <= minChunkSz {
		// See: https://docs.aws.amazon.com/AmazonS3/latest/userguide/qfacts.html
		cfg.PutChunkSize = minChunkSz
	}

	// Assume 0 list size == use default
	if cfg.ListSize <= 0 {
		cfg.ListSize = 200
	}

	// Return owned config copy
	return S3Config{
		CoreOpts:     cfg.CoreOpts,
		GetOpts:      cfg.GetOpts,
		PutOpts:      cfg.PutOpts,
		PutChunkSize: cfg.PutChunkSize,
		ListSize:     cfg.ListSize,
		StatOpts:     cfg.StatOpts,
		RemoveOpts:   cfg.RemoveOpts,
	}
}

// S3Storage is a storage implementation that stores key-value
// pairs in an S3 instance at given endpoint with bucket name.
type S3Storage struct {
	client *minio.Core
	bucket string
	config S3Config
	state  uint32
}

// OpenS3 opens a new S3Storage instance with given S3 endpoint URL, bucket name and configuration.
func OpenS3(endpoint string, bucket string, cfg *S3Config) (*S3Storage, error) {
	// Get checked config
	config := getS3Config(cfg)

	// Create new S3 client connection
	client, err := minio.NewCore(endpoint, &config.CoreOpts)
	if err != nil {
		return nil, err
	}

	// Check that provided bucket actually exists
	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, new_error("bucket does not exist")
	}

	return &S3Storage{
		client: client,
		bucket: bucket,
		config: config,
	}, nil
}

// Client returns access to the underlying S3 client.
func (st *S3Storage) Client() *minio.Core {
	return st.client
}

// Clean implements Storage.Clean().
func (st *S3Storage) Clean(ctx context.Context) error {
	return nil // nothing to do for S3
}

// ReadBytes implements Storage.ReadBytes().
func (st *S3Storage) ReadBytes(ctx context.Context, key string) ([]byte, error) {
	// Fetch object reader from S3 bucket
	rc, err := st.ReadStream(ctx, key)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	// Read all bytes and return
	return io.ReadAll(rc)
}

// ReadStream implements Storage.ReadStream().
func (st *S3Storage) ReadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	// Check storage open
	if st.closed() {
		return nil, ErrClosed
	}

	// Fetch object reader from S3 bucket
	rc, _, _, err := st.client.GetObject(
		ctx,
		st.bucket,
		key,
		st.config.GetOpts,
	)
	if err != nil {
		return nil, transformS3Error(err)
	}

	return rc, nil
}

// WriteBytes implements Storage.WriteBytes().
func (st *S3Storage) WriteBytes(ctx context.Context, key string, value []byte) error {
	return st.WriteStream(ctx, key, util.NewByteReaderSize(value))
}

// WriteStream implements Storage.WriteStream().
func (st *S3Storage) WriteStream(ctx context.Context, key string, r io.Reader) error {
	// Check storage open
	if st.closed() {
		return ErrClosed
	}

	if rs, ok := r.(util.ReaderSize); ok {
		// This reader supports providing us the size of
		// the encompassed data, allowing us to perform
		// a singular .PutObject() call with length.
		_, err := st.client.PutObject(
			ctx,
			st.bucket,
			key,
			r,
			rs.Size(),
			"",
			"",
			st.config.PutOpts,
		)
		if err != nil {
			return transformS3Error(err)
		}
		return nil
	}

	// Start a new multipart upload to get ID
	uploadID, err := st.client.NewMultipartUpload(
		ctx,
		st.bucket,
		key,
		st.config.PutOpts,
	)
	if err != nil {
		return transformS3Error(err)
	}

	var (
		count = 1
		parts []minio.CompletePart
		chunk = make([]byte, st.config.PutChunkSize)
		rdr   = bytes.NewReader(nil)
	)

	// Note that we do not perform any kind of
	// memory pooling of the chunk buffers here.
	// Optimal chunking sizes for S3 writes are in
	// the orders of megabytes, so letting the GC
	// collect these ASAP is much preferred.

loop:
	for done := false; !done; {
		// Read next chunk into byte buffer
		n, err := io.ReadFull(r, chunk)

		switch err {
		// Successful read
		case nil:

		// Reached end, buffer empty
		case io.EOF:
			break loop

		// Reached end, but buffer not empty
		case io.ErrUnexpectedEOF:
			done = true

		// All other errors
		default:
			return err
		}

		// Reset byte reader
		rdr.Reset(chunk[:n])

		// Put this object chunk in S3 store
		pt, err := st.client.PutObjectPart(
			ctx,
			st.bucket,
			key,
			uploadID,
			count,
			rdr,
			int64(n),
			"",
			"",
			nil,
		)
		if err != nil {
			return err
		}

		// Append completed part to slice
		parts = append(parts, minio.CompletePart{
			PartNumber:     pt.PartNumber,
			ETag:           pt.ETag,
			ChecksumCRC32:  pt.ChecksumCRC32,
			ChecksumCRC32C: pt.ChecksumCRC32C,
			ChecksumSHA1:   pt.ChecksumSHA1,
			ChecksumSHA256: pt.ChecksumSHA256,
		})

		// Iterate part count
		count++
	}

	// Complete this multi-part upload operation
	_, err = st.client.CompleteMultipartUpload(
		ctx,
		st.bucket,
		key,
		uploadID,
		parts,
		st.config.PutOpts,
	)
	if err != nil {
		return err
	}

	return nil
}

// Stat implements Storage.Stat().
func (st *S3Storage) Stat(ctx context.Context, key string) (bool, error) {
	// Check storage open
	if st.closed() {
		return false, ErrClosed
	}

	// Query object in S3 bucket
	_, err := st.client.StatObject(
		ctx,
		st.bucket,
		key,
		st.config.StatOpts,
	)
	if err != nil {
		return false, transformS3Error(err)
	}

	return true, nil
}

// Remove implements Storage.Remove().
func (st *S3Storage) Remove(ctx context.Context, key string) error {
	// Check storage open
	if st.closed() {
		return ErrClosed
	}

	// S3 returns no error on remove for non-existent keys
	if ok, err := st.Stat(ctx, key); err != nil {
		return err
	} else if !ok {
		return ErrNotFound
	}

	// Remove object from S3 bucket
	err := st.client.RemoveObject(
		ctx,
		st.bucket,
		key,
		st.config.RemoveOpts,
	)
	if err != nil {
		return transformS3Error(err)
	}

	return nil
}

// WalkKeys implements Storage.WalkKeys().
func (st *S3Storage) WalkKeys(ctx context.Context, opts WalkKeysOptions) error {
	var (
		prev  string
		token string
	)

	for {
		// List the objects in bucket starting at marker
		result, err := st.client.ListObjectsV2(
			st.bucket,
			"",
			prev,
			token,
			"",
			st.config.ListSize,
		)
		if err != nil {
			return err
		}

		// Pass each object through walk func
		for _, obj := range result.Contents {
			if err := opts.WalkFn(ctx, Entry{
				Key:  obj.Key,
				Size: obj.Size,
			}); err != nil {
				return err
			}
		}

		// No token means we reached end of bucket
		if result.NextContinuationToken == "" {
			return nil
		}

		// Set continue token and prev mark
		token = result.NextContinuationToken
		prev = result.StartAfter
	}
}

// Close implements Storage.Close().
func (st *S3Storage) Close() error {
	atomic.StoreUint32(&st.state, 1)
	return nil
}

// closed returns whether S3Storage is closed.
func (st *S3Storage) closed() bool {
	return (atomic.LoadUint32(&st.state) == 1)
}

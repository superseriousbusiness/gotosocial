package storage

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	_path "path"
	"strings"
	"syscall"

	"codeberg.org/gruf/go-bytes"
	"codeberg.org/gruf/go-fastcopy"
	"codeberg.org/gruf/go-store/v2/util"
)

// DefaultDiskConfig is the default DiskStorage configuration.
var DefaultDiskConfig = &DiskConfig{
	Overwrite:    true,
	WriteBufSize: 4096,
	Transform:    NopTransform(),
	Compression:  NoCompression(),
}

// DiskConfig defines options to be used when opening a DiskStorage.
type DiskConfig struct {
	// Transform is the supplied key <--> path KeyTransform.
	Transform KeyTransform

	// WriteBufSize is the buffer size to use when writing file streams.
	WriteBufSize int

	// Overwrite allows overwriting values of stored keys in the storage.
	Overwrite bool

	// LockFile allows specifying the filesystem path to use for the lockfile,
	// providing only a filename it will store the lockfile within provided store
	// path and nest the store under `path/store` to prevent access to lockfile.
	LockFile string

	// Compression is the Compressor to use when reading / writing files,
	// default is no compression.
	Compression Compressor
}

// getDiskConfig returns a valid DiskConfig for supplied ptr.
func getDiskConfig(cfg *DiskConfig) DiskConfig {
	// If nil, use default
	if cfg == nil {
		cfg = DefaultDiskConfig
	}

	// Assume nil transform == none
	if cfg.Transform == nil {
		cfg.Transform = NopTransform()
	}

	// Assume nil compress == none
	if cfg.Compression == nil {
		cfg.Compression = NoCompression()
	}

	// Assume 0 buf size == use default
	if cfg.WriteBufSize <= 0 {
		cfg.WriteBufSize = DefaultDiskConfig.WriteBufSize
	}

	// Assume empty lockfile path == use default
	if len(cfg.LockFile) == 0 {
		cfg.LockFile = LockFile
	}

	// Return owned config copy
	return DiskConfig{
		Transform:    cfg.Transform,
		WriteBufSize: cfg.WriteBufSize,
		Overwrite:    cfg.Overwrite,
		LockFile:     cfg.LockFile,
		Compression:  cfg.Compression,
	}
}

// DiskStorage is a Storage implementation that stores directly to a filesystem.
type DiskStorage struct {
	path   string            // path is the root path of this store
	cppool fastcopy.CopyPool // cppool is the prepared io copier with buffer pool
	config DiskConfig        // cfg is the supplied configuration for this store
	lock   *Lock             // lock is the opened lockfile for this storage instance
}

// OpenDisk opens a DiskStorage instance for given folder path and configuration.
func OpenDisk(path string, cfg *DiskConfig) (*DiskStorage, error) {
	// Get checked config
	config := getDiskConfig(cfg)

	// Acquire path builder
	pb := util.GetPathBuilder()
	defer util.PutPathBuilder(pb)

	// Clean provided store path, ensure
	// ends in '/' to help later path trimming
	storePath := pb.Clean(path) + "/"

	// Clean provided lockfile path
	lockfile := pb.Clean(config.LockFile)

	// Check if lockfile is an *actual* path or just filename
	if lockDir, _ := _path.Split(lockfile); lockDir == "" {
		// Lockfile is a filename, store must be nested under
		// $storePath/store to prevent access to the lockfile
		storePath += "store/"
		lockfile = pb.Join(path, lockfile)
	}

	// Attempt to open dir path
	file, err := os.OpenFile(storePath, defaultFileROFlags, defaultDirPerms)
	if err != nil {
		// If not a not-exist error, return
		if !os.IsNotExist(err) {
			return nil, err
		}

		// Attempt to make store path dirs
		err = os.MkdirAll(storePath, defaultDirPerms)
		if err != nil {
			return nil, err
		}

		// Reopen dir now it's been created
		file, err = os.OpenFile(storePath, defaultFileROFlags, defaultDirPerms)
		if err != nil {
			return nil, err
		}
	}
	defer file.Close()

	// Double check this is a dir (NOT a file!)
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, errors.New("store/storage: path is file")
	}

	// Open and acquire storage lock for path
	lock, err := OpenLock(lockfile)
	if err != nil {
		return nil, err
	}

	// Prepare DiskStorage
	st := &DiskStorage{
		path:   storePath,
		config: config,
		lock:   lock,
	}

	// Set copypool buffer size
	st.cppool.Buffer(config.WriteBufSize)

	return st, nil
}

// Clean implements Storage.Clean().
func (st *DiskStorage) Clean(ctx context.Context) error {
	// Check if open
	if st.lock.Closed() {
		return ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return err
	}

	// Clean-out unused directories
	return cleanDirs(st.path)
}

// ReadBytes implements Storage.ReadBytes().
func (st *DiskStorage) ReadBytes(ctx context.Context, key string) ([]byte, error) {
	// Get stream reader for key
	rc, err := st.ReadStream(ctx, key)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	// Read all bytes and return
	return io.ReadAll(rc)
}

// ReadStream implements Storage.ReadStream().
func (st *DiskStorage) ReadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	// Get file path for key
	kpath, err := st.filepath(key)
	if err != nil {
		return nil, err
	}

	// Check if open
	if st.lock.Closed() {
		return nil, ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Attempt to open file (replace ENOENT with our own)
	file, err := open(kpath, defaultFileROFlags)
	if err != nil {
		return nil, errSwapNotFound(err)
	}

	// Wrap the file in a compressor
	cFile, err := st.config.Compression.Reader(file)
	if err != nil {
		_ = file.Close()
		return nil, err
	}

	return cFile, nil
}

// WriteBytes implements Storage.WriteBytes().
func (st *DiskStorage) WriteBytes(ctx context.Context, key string, value []byte) (int, error) {
	n, err := st.WriteStream(ctx, key, bytes.NewReader(value))
	return int(n), err
}

// WriteStream implements Storage.WriteStream().
func (st *DiskStorage) WriteStream(ctx context.Context, key string, r io.Reader) (int64, error) {
	// Get file path for key
	kpath, err := st.filepath(key)
	if err != nil {
		return 0, err
	}

	// Check if open
	if st.lock.Closed() {
		return 0, ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	// Ensure dirs leading up to file exist
	err = os.MkdirAll(path.Dir(kpath), defaultDirPerms)
	if err != nil {
		return 0, err
	}

	// Prepare to swap error if need-be
	errSwap := errSwapNoop

	// Build file RW flags
	flags := defaultFileRWFlags
	if !st.config.Overwrite {
		flags |= syscall.O_EXCL

		// Catch + replace err exist
		errSwap = errSwapExist
	}

	// Attempt to open file
	file, err := open(kpath, flags)
	if err != nil {
		return 0, errSwap(err)
	}

	// Wrap the file in a compressor
	cFile, err := st.config.Compression.Writer(file)
	if err != nil {
		_ = file.Close()
		return 0, err
	}

	// Wraps file.Close().
	defer cFile.Close()

	// Copy provided reader to file
	return st.cppool.Copy(cFile, r)
}

// Stat implements Storage.Stat().
func (st *DiskStorage) Stat(ctx context.Context, key string) (bool, error) {
	// Get file path for key
	kpath, err := st.filepath(key)
	if err != nil {
		return false, err
	}

	// Check if open
	if st.lock.Closed() {
		return false, ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return false, err
	}

	// Check for file on disk
	return stat(kpath)
}

// Remove implements Storage.Remove().
func (st *DiskStorage) Remove(ctx context.Context, key string) error {
	// Get file path for key
	kpath, err := st.filepath(key)
	if err != nil {
		return err
	}

	// Check if open
	if st.lock.Closed() {
		return ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return err
	}

	// Remove at path (we know this is file)
	if err := unlink(kpath); err != nil {
		return errSwapNotFound(err)
	}

	return nil
}

// Close implements Storage.Close().
func (st *DiskStorage) Close() error {
	return st.lock.Close()
}

// WalkKeys implements Storage.WalkKeys().
func (st *DiskStorage) WalkKeys(ctx context.Context, opts WalkKeysOptions) error {
	// Check if open
	if st.lock.Closed() {
		return ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return err
	}

	// Acquire path builder
	pb := util.GetPathBuilder()
	defer util.PutPathBuilder(pb)

	// Walk dir for entries
	return walkDir(pb, st.path, func(kpath string, fsentry fs.DirEntry) error {
		if !fsentry.Type().IsRegular() {
			// Only deal with regular files
			return nil
		}

		// Get full item path (without root)
		kpath = pb.Join(kpath, fsentry.Name())
		kpath = kpath[len(st.path):]

		// Load file info. This should already
		// be loaded due to the underlying call
		// to os.File{}.ReadDir() populating them
		info, err := fsentry.Info()
		if err != nil {
			return err
		}

		// Perform provided walk function
		return opts.WalkFn(ctx, Entry{
			Key:  st.config.Transform.PathToKey(kpath),
			Size: info.Size(),
		})
	})
}

// filepath checks and returns a formatted filepath for given key.
func (st *DiskStorage) filepath(key string) (string, error) {
	// Calculate transformed key path
	key = st.config.Transform.KeyToPath(key)

	// Acquire path builder
	pb := util.GetPathBuilder()
	defer util.PutPathBuilder(pb)

	// Generate key path
	pb.Append(st.path)
	pb.Append(key)

	// Check for dir traversal outside of root
	if isDirTraversal(st.path, pb.String()) {
		return "", ErrInvalidKey
	}

	return string(pb.B), nil
}

// isDirTraversal will check if rootPlusPath is a dir traversal outside of root,
// assuming that both are cleaned and that rootPlusPath is path.Join(root, somePath).
func isDirTraversal(root, rootPlusPath string) bool {
	switch {
	// Root is $PWD, check for traversal out of
	case root == ".":
		return strings.HasPrefix(rootPlusPath, "../")

	// The path MUST be prefixed by root
	case !strings.HasPrefix(rootPlusPath, root):
		return true

	// In all other cases, check not equal
	default:
		return len(root) == len(rootPlusPath)
	}
}

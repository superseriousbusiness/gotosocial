package storage

import (
	"io"
	"io/fs"
	"os"
	"path"
	_path "path"
	"strings"
	"syscall"

	"codeberg.org/gruf/go-bytes"
	"codeberg.org/gruf/go-fastcopy"
	"codeberg.org/gruf/go-store/util"
)

// DefaultDiskConfig is the default DiskStorage configuration
var DefaultDiskConfig = &DiskConfig{
	Overwrite:    true,
	WriteBufSize: 4096,
	Transform:    NopTransform(),
	Compression:  NoCompression(),
}

// DiskConfig defines options to be used when opening a DiskStorage
type DiskConfig struct {
	// Transform is the supplied key<-->path KeyTransform
	Transform KeyTransform

	// WriteBufSize is the buffer size to use when writing file streams (PutStream)
	WriteBufSize int

	// Overwrite allows overwriting values of stored keys in the storage
	Overwrite bool

	// LockFile allows specifying the filesystem path to use for the lockfile,
	// providing only a filename it will store the lockfile within provided store
	// path and nest the store under `path/store` to prevent access to lockfile
	LockFile string

	// Compression is the Compressor to use when reading / writing files, default is no compression
	Compression Compressor
}

// getDiskConfig returns a valid DiskConfig for supplied ptr
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
	if cfg.WriteBufSize < 1 {
		cfg.WriteBufSize = DefaultDiskConfig.WriteBufSize
	}

	// Assume empty lockfile path == use default
	if len(cfg.LockFile) < 1 {
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

// DiskStorage is a Storage implementation that stores directly to a filesystem
type DiskStorage struct {
	path   string            // path is the root path of this store
	cppool fastcopy.CopyPool // cppool is the prepared io copier with buffer pool
	config DiskConfig        // cfg is the supplied configuration for this store
	lock   *Lock             // lock is the opened lockfile for this storage instance
}

// OpenFile opens a DiskStorage instance for given folder path and configuration
func OpenFile(path string, cfg *DiskConfig) (*DiskStorage, error) {
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
	if lockDir, _ := _path.Split(lockfile); len(lockDir) < 1 {
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
		return nil, errPathIsFile
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

// Clean implements Storage.Clean()
func (st *DiskStorage) Clean() error {
	st.lock.Add()
	defer st.lock.Done()
	if st.lock.Closed() {
		return ErrClosed
	}
	return util.CleanDirs(st.path)
}

// ReadBytes implements Storage.ReadBytes()
func (st *DiskStorage) ReadBytes(key string) ([]byte, error) {
	// Get stream reader for key
	rc, err := st.ReadStream(key)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	// Read all bytes and return
	return io.ReadAll(rc)
}

// ReadStream implements Storage.ReadStream()
func (st *DiskStorage) ReadStream(key string) (io.ReadCloser, error) {
	// Get file path for key
	kpath, err := st.filepath(key)
	if err != nil {
		return nil, err
	}

	// Track open
	st.lock.Add()

	// Check if open
	if st.lock.Closed() {
		return nil, ErrClosed
	}

	// Attempt to open file (replace ENOENT with our own)
	file, err := open(kpath, defaultFileROFlags)
	if err != nil {
		st.lock.Done()
		return nil, errSwapNotFound(err)
	}

	// Wrap the file in a compressor
	cFile, err := st.config.Compression.Reader(file)
	if err != nil {
		file.Close() // close this here, ignore error
		st.lock.Done()
		return nil, err
	}

	// Wrap compressor to ensure file close
	return util.ReadCloserWithCallback(cFile, func() {
		file.Close()
		st.lock.Done()
	}), nil
}

// WriteBytes implements Storage.WriteBytes()
func (st *DiskStorage) WriteBytes(key string, value []byte) error {
	return st.WriteStream(key, bytes.NewReader(value))
}

// WriteStream implements Storage.WriteStream()
func (st *DiskStorage) WriteStream(key string, r io.Reader) error {
	// Get file path for key
	kpath, err := st.filepath(key)
	if err != nil {
		return err
	}

	// Track open
	st.lock.Add()
	defer st.lock.Done()

	// Check if open
	if st.lock.Closed() {
		return ErrClosed
	}

	// Ensure dirs leading up to file exist
	err = os.MkdirAll(path.Dir(kpath), defaultDirPerms)
	if err != nil {
		return err
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
		return errSwap(err)
	}
	defer file.Close()

	// Wrap the file in a compressor
	cFile, err := st.config.Compression.Writer(file)
	if err != nil {
		return err
	}
	defer cFile.Close()

	// Copy provided reader to file
	_, err = st.cppool.Copy(cFile, r)
	return err
}

// Stat implements Storage.Stat()
func (st *DiskStorage) Stat(key string) (bool, error) {
	// Get file path for key
	kpath, err := st.filepath(key)
	if err != nil {
		return false, err
	}

	// Track open
	st.lock.Add()
	defer st.lock.Done()

	// Check if open
	if st.lock.Closed() {
		return false, ErrClosed
	}

	// Check for file on disk
	return stat(kpath)
}

// Remove implements Storage.Remove()
func (st *DiskStorage) Remove(key string) error {
	// Get file path for key
	kpath, err := st.filepath(key)
	if err != nil {
		return err
	}

	// Track open
	st.lock.Add()
	defer st.lock.Done()

	// Check if open
	if st.lock.Closed() {
		return ErrClosed
	}

	// Remove at path (we know this is file)
	if err := unlink(kpath); err != nil {
		return errSwapNotFound(err)
	}

	return nil
}

// Close implements Storage.Close()
func (st *DiskStorage) Close() error {
	return st.lock.Close()
}

// WalkKeys implements Storage.WalkKeys()
func (st *DiskStorage) WalkKeys(opts WalkKeysOptions) error {
	// Track open
	st.lock.Add()
	defer st.lock.Done()

	// Check if open
	if st.lock.Closed() {
		return ErrClosed
	}

	// Acquire path builder
	pb := util.GetPathBuilder()
	defer util.PutPathBuilder(pb)

	// Walk dir for entries
	return util.WalkDir(pb, st.path, func(kpath string, fsentry fs.DirEntry) {
		if fsentry.Type().IsRegular() {
			// Only deal with regular files

			// Get full item path (without root)
			kpath = pb.Join(kpath, fsentry.Name())[len(st.path):]

			// Perform provided walk function
			opts.WalkFn(entry(st.config.Transform.PathToKey(kpath)))
		}
	})
}

// filepath checks and returns a formatted filepath for given key
func (st *DiskStorage) filepath(key string) (string, error) {
	// Calculate transformed key path
	key = st.config.Transform.KeyToPath(key)

	// Acquire path builder
	pb := util.GetPathBuilder()
	defer util.PutPathBuilder(pb)

	// Generated joined root path
	pb.AppendString(st.path)
	pb.AppendString(key)

	// Check for dir traversal outside of root
	if isDirTraversal(st.path, pb.StringPtr()) {
		return "", ErrInvalidKey
	}

	return pb.String(), nil
}

// isDirTraversal will check if rootPlusPath is a dir traversal outside of root,
// assuming that both are cleaned and that rootPlusPath is path.Join(root, somePath)
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

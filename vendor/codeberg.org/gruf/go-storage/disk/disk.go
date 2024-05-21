package disk

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"syscall"

	"codeberg.org/gruf/go-fastcopy"
	"codeberg.org/gruf/go-fastpath/v2"
	"codeberg.org/gruf/go-storage"
	"codeberg.org/gruf/go-storage/internal"
)

// ensure DiskStorage conforms to storage.Storage.
var _ storage.Storage = (*DiskStorage)(nil)

// DefaultConfig returns the default DiskStorage configuration.
func DefaultConfig() Config {
	return defaultConfig
}

// immutable default configuration.
var defaultConfig = Config{
	OpenRead:     OpenArgs{syscall.O_RDONLY, 0o644},
	OpenWrite:    OpenArgs{syscall.O_CREAT | syscall.O_WRONLY, 0o644},
	MkdirPerms:   0o755,
	WriteBufSize: 4096,
}

// OpenArgs defines args passed
// in a syscall.Open() operation.
type OpenArgs struct {
	Flags int
	Perms uint32
}

// Config defines options to be
// used when opening a DiskStorage.
type Config struct {

	// OpenRead are the arguments passed
	// to syscall.Open() when opening a
	// file for read operations.
	OpenRead OpenArgs

	// OpenWrite are the arguments passed
	// to syscall.Open() when opening a
	// file for write operations.
	OpenWrite OpenArgs

	// MkdirPerms are the permissions used
	// when creating necessary sub-dirs in
	// a storage key with slashes.
	MkdirPerms uint32

	// WriteBufSize is the buffer size
	// to use when writing file streams.
	WriteBufSize int
}

// getDiskConfig returns valid (and owned!) Config for given ptr.
func getDiskConfig(cfg *Config) Config {
	if cfg == nil {
		// use defaults.
		return defaultConfig
	}

	// Ensure non-zero syscall args.
	if cfg.OpenRead.Flags == 0 {
		cfg.OpenRead.Flags = defaultConfig.OpenRead.Flags
	}
	if cfg.OpenRead.Perms == 0 {
		cfg.OpenRead.Perms = defaultConfig.OpenRead.Perms
	}
	if cfg.OpenWrite.Flags == 0 {
		cfg.OpenWrite.Flags = defaultConfig.OpenWrite.Flags
	}
	if cfg.OpenWrite.Perms == 0 {
		cfg.OpenWrite.Perms = defaultConfig.OpenWrite.Perms
	}
	if cfg.MkdirPerms == 0 {
		cfg.MkdirPerms = defaultConfig.MkdirPerms
	}

	// Ensure valid write buf.
	if cfg.WriteBufSize <= 0 {
		cfg.WriteBufSize = defaultConfig.WriteBufSize
	}

	return Config{
		OpenRead:     cfg.OpenRead,
		OpenWrite:    cfg.OpenWrite,
		MkdirPerms:   cfg.MkdirPerms,
		WriteBufSize: cfg.WriteBufSize,
	}
}

// DiskStorage is a Storage implementation
// that stores directly to a filesystem.
type DiskStorage struct {
	path string            // path is the root path of this store
	pool fastcopy.CopyPool // pool is the prepared io copier with buffer pool
	cfg  Config            // cfg is the supplied configuration for this store
}

// Open opens a DiskStorage instance for given folder path and configuration.
func Open(path string, cfg *Config) (*DiskStorage, error) {
	// Check + set config defaults.
	config := getDiskConfig(cfg)

	// Clean provided storage path, ensure
	// final '/' to help with path trimming.
	pb := internal.GetPathBuilder()
	path = pb.Clean(path) + "/"
	internal.PutPathBuilder(pb)

	// Ensure directories up-to path exist.
	perms := fs.FileMode(config.MkdirPerms)
	err := os.MkdirAll(path, perms)
	if err != nil {
		return nil, err
	}

	// Prepare DiskStorage.
	st := &DiskStorage{
		path: path,
		cfg:  config,
	}

	// Set fastcopy pool buffer size.
	st.pool.Buffer(config.WriteBufSize)

	return st, nil
}

// Clean: implements Storage.Clean().
func (st *DiskStorage) Clean(ctx context.Context) error {
	// Check context still valid.
	if err := ctx.Err(); err != nil {
		return err
	}

	// Clean unused directories.
	return cleanDirs(st.path, OpenArgs{
		Flags: syscall.O_RDONLY,
	})
}

// ReadBytes: implements Storage.ReadBytes().
func (st *DiskStorage) ReadBytes(ctx context.Context, key string) ([]byte, error) {
	// Get stream reader for key
	rc, err := st.ReadStream(ctx, key)
	if err != nil {
		return nil, err
	}

	// Read all data to memory.
	data, err := io.ReadAll(rc)
	if err != nil {
		_ = rc.Close()
		return nil, err
	}

	// Close storage stream reader.
	if err := rc.Close(); err != nil {
		return nil, err
	}

	return data, nil
}

// ReadStream: implements Storage.ReadStream().
func (st *DiskStorage) ReadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	// Generate file path for key.
	kpath, err := st.Filepath(key)
	if err != nil {
		return nil, err
	}

	// Check context still valid.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Attempt to open file with read args.
	file, err := open(kpath, st.cfg.OpenRead)
	if err != nil {

		if err == syscall.ENOENT {
			// Translate not-found errors and wrap with key.
			err = internal.ErrWithKey(storage.ErrNotFound, key)
		}

		return nil, err
	}

	return file, nil
}

// WriteBytes: implements Storage.WriteBytes().
func (st *DiskStorage) WriteBytes(ctx context.Context, key string, value []byte) (int, error) {
	n, err := st.WriteStream(ctx, key, bytes.NewReader(value))
	return int(n), err
}

// WriteStream: implements Storage.WriteStream().
func (st *DiskStorage) WriteStream(ctx context.Context, key string, stream io.Reader) (int64, error) {
	// Acquire path builder buffer.
	pb := internal.GetPathBuilder()

	// Generate the file path for given key.
	kpath, subdir, err := st.filepath(pb, key)
	if err != nil {
		return 0, err
	}

	// Done with path buffer.
	internal.PutPathBuilder(pb)

	// Check context still valid.
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	if subdir {
		// Get dir of key path.
		dir := path.Dir(kpath)

		// Note that subdir will only be set if
		// the transformed key (without base path)
		// contains any slashes. This is not a
		// definitive check, but it allows us to
		// skip a syscall if mkdirall not needed!
		perms := fs.FileMode(st.cfg.MkdirPerms)
		err = os.MkdirAll(dir, perms)
		if err != nil {
			return 0, err
		}
	}

	// Attempt to open file with write args.
	file, err := open(kpath, st.cfg.OpenWrite)
	if err != nil {

		if st.cfg.OpenWrite.Flags&syscall.O_EXCL != 0 &&
			err == syscall.EEXIST {
			// Translate already exists errors and wrap with key.
			err = internal.ErrWithKey(storage.ErrAlreadyExists, key)
		}

		return 0, err
	}

	// Copy provided stream to file interface.
	n, err := st.pool.Copy(file, stream)
	if err != nil {
		_ = file.Close()
		return n, err
	}

	// Finally, close file.
	return n, file.Close()
}

// Stat implements Storage.Stat().
func (st *DiskStorage) Stat(ctx context.Context, key string) (*storage.Entry, error) {
	// Generate file path for key.
	kpath, err := st.Filepath(key)
	if err != nil {
		return nil, err
	}

	// Check context still valid.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Stat file on disk.
	stat, err := stat(kpath)
	if stat == nil {
		return nil, err
	}

	return &storage.Entry{
		Key:  key,
		Size: stat.Size,
	}, nil
}

// Remove implements Storage.Remove().
func (st *DiskStorage) Remove(ctx context.Context, key string) error {
	// Generate file path for key.
	kpath, err := st.Filepath(key)
	if err != nil {
		return err
	}

	// Check context still valid.
	if err := ctx.Err(); err != nil {
		return err
	}

	// Stat file on disk.
	stat, err := stat(kpath)
	if err != nil {
		return err
	}

	// Not-found (or handled
	// as) error situations.
	if stat == nil {
		return internal.ErrWithKey(storage.ErrNotFound, key)
	} else if stat.Mode&syscall.S_IFREG == 0 {
		err := errors.New("storage/disk: not a regular file")
		return internal.ErrWithKey(err, key)
	}

	// Remove at path (we know this is file).
	if err := unlink(kpath); err != nil {

		if err == syscall.ENOENT {
			// Translate not-found errors and wrap with key.
			err = internal.ErrWithKey(storage.ErrNotFound, key)
		}

		return err
	}

	return nil
}

// WalkKeys implements Storage.WalkKeys().
func (st *DiskStorage) WalkKeys(ctx context.Context, opts storage.WalkKeysOpts) error {
	if opts.Step == nil {
		panic("nil step fn")
	}

	// Check context still valid.
	if err := ctx.Err(); err != nil {
		return err
	}

	// Acquire path builder for walk.
	pb := internal.GetPathBuilder()
	defer internal.PutPathBuilder(pb)

	// Dir to walk.
	dir := st.path

	if opts.Prefix != "" {
		// Convert key prefix to one of our storage filepaths.
		pathprefix, subdir, err := st.filepath(pb, opts.Prefix)
		if err != nil {
			return internal.ErrWithMsg(err, "prefix error")
		}

		if subdir {
			// Note that subdir will only be set if
			// the transformed key (without base path)
			// contains any slashes. This is not a
			// definitive check, but it allows us to
			// update the directory we walk in case
			// it might narrow search parameters!
			dir = path.Dir(pathprefix)
		}

		// Set updated storage
		// path prefix in opts.
		opts.Prefix = pathprefix
	}

	// Only need to open dirs as read-only.
	args := OpenArgs{Flags: syscall.O_RDONLY}

	return walkDir(pb, dir, args, func(kpath string, fsentry fs.DirEntry) error {
		if !fsentry.Type().IsRegular() {
			// Ignore anything but
			// regular file types.
			return nil
		}

		// Get full item path (without root).
		kpath = pb.Join(kpath, fsentry.Name())

		// Perform a fast filter check against storage path prefix (if set).
		if opts.Prefix != "" && !strings.HasPrefix(kpath, opts.Prefix) {
			return nil // ignore
		}

		// Storage key without base.
		key := kpath[len(st.path):]

		// Ignore filtered keys.
		if opts.Filter != nil &&
			!opts.Filter(key) {
			return nil // ignore
		}

		// Load file info. This should already
		// be loaded due to the underlying call
		// to os.File{}.ReadDir() populating them.
		info, err := fsentry.Info()
		if err != nil {
			return err
		}

		// Perform provided walk function
		return opts.Step(storage.Entry{
			Key:  key,
			Size: info.Size(),
		})
	})
}

// Filepath checks and returns a formatted Filepath for given key.
func (st *DiskStorage) Filepath(key string) (path string, err error) {
	pb := internal.GetPathBuilder()
	path, _, err = st.filepath(pb, key)
	internal.PutPathBuilder(pb)
	return
}

// filepath performs the "meat" of Filepath(), returning also if path *may* be a subdir of base.
func (st *DiskStorage) filepath(pb *fastpath.Builder, key string) (path string, subdir bool, err error) {
	// Fast check for whether this may be a
	// sub-directory. This is not a definitive
	// check, it's only for a fastpath check.
	subdir = strings.ContainsRune(key, '/')

	// Build from base.
	pb.Append(st.path)
	pb.Append(key)

	// Take COPY of bytes.
	path = string(pb.B)

	// Check for dir traversal outside base.
	if isDirTraversal(st.path, path) {
		err = internal.ErrWithKey(storage.ErrInvalidKey, key)
	}

	return
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

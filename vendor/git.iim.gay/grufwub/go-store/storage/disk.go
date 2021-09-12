package storage

import (
	"io"
	"io/fs"
	"os"
	"path"
	"syscall"

	"git.iim.gay/grufwub/fastpath"
	"git.iim.gay/grufwub/go-bytes"
	"git.iim.gay/grufwub/go-store/util"
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

	// Return owned config copy
	return DiskConfig{
		Transform:    cfg.Transform,
		WriteBufSize: cfg.WriteBufSize,
		Overwrite:    cfg.Overwrite,
		Compression:  cfg.Compression,
	}
}

// DiskStorage is a Storage implementation that stores directly to a filesystem
type DiskStorage struct {
	path   string     // path is the root path of this store
	dots   int        // dots is the "dotdot" count for the root store path
	config DiskConfig // cfg is the supplied configuration for this store
}

// OpenFile opens a DiskStorage instance for given folder path and configuration
func OpenFile(path string, cfg *DiskConfig) (*DiskStorage, error) {
	// Acquire path builder
	pb := util.AcquirePathBuilder()
	defer util.ReleasePathBuilder(pb)

	// Clean provided path, ensure ends in '/' (should
	// be dir, this helps with file path trimming later)
	path = pb.Clean(path) + "/"

	// Get checked config
	config := getDiskConfig(cfg)

	// Attempt to open dir path
	file, err := os.OpenFile(path, defaultFileROFlags, defaultDirPerms)
	if err != nil {
		// If not a not-exist error, return
		if !os.IsNotExist(err) {
			return nil, err
		}

		// Attempt to make store path dirs
		err = os.MkdirAll(path, defaultDirPerms)
		if err != nil {
			return nil, err
		}

		// Reopen dir now it's been created
		file, err = os.OpenFile(path, defaultFileROFlags, defaultDirPerms)
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

	// Return new DiskStorage
	return &DiskStorage{
		path:   path,
		dots:   util.CountDotdots(path),
		config: config,
	}, nil
}

// Clean implements Storage.Clean()
func (st *DiskStorage) Clean() error {
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

	// Attempt to open file (replace ENOENT with our own)
	file, err := open(kpath, defaultFileROFlags)
	if err != nil {
		return nil, errSwapNotFound(err)
	}

	// Wrap the file in a compressor
	cFile, err := st.config.Compression.Reader(file)
	if err != nil {
		file.Close() // close this here, ignore error
		return nil, err
	}

	// Wrap compressor to ensure file close
	return util.ReadCloserWithCallback(cFile, func() {
		file.Close()
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

	// Acquire write buffer
	buf := util.AcquireBuffer(st.config.WriteBufSize)
	defer util.ReleaseBuffer(buf)
	buf.Grow(st.config.WriteBufSize)

	// Copy reader to file
	_, err = io.CopyBuffer(cFile, r, buf.B)
	return err
}

// Stat implements Storage.Stat()
func (st *DiskStorage) Stat(key string) (bool, error) {
	// Get file path for key
	kpath, err := st.filepath(key)
	if err != nil {
		return false, err
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

	// Attempt to remove file
	return os.Remove(kpath)
}

// WalkKeys implements Storage.WalkKeys()
func (st *DiskStorage) WalkKeys(opts *WalkKeysOptions) error {
	// Acquire path builder
	pb := fastpath.AcquireBuilder()
	defer fastpath.ReleaseBuilder(pb)

	// Walk dir for entries
	return util.WalkDir(pb, st.path, func(kpath string, fsentry fs.DirEntry) {
		// Only deal with regular files
		if fsentry.Type().IsRegular() {
			// Get full item path (without root)
			kpath = pb.Join(kpath, fsentry.Name())[len(st.path):]

			// Perform provided walk function
			opts.WalkFn(entry(st.config.Transform.PathToKey(kpath)))
		}
	})
}

// filepath checks and returns a formatted filepath for given key
func (st *DiskStorage) filepath(key string) (string, error) {
	// Acquire path builder
	pb := util.AcquirePathBuilder()
	defer util.ReleasePathBuilder(pb)

	// Calculate transformed key path
	key = st.config.Transform.KeyToPath(key)

	// Generated joined root path
	pb.AppendString(st.path)
	pb.AppendString(key)

	// If path is dir traversal, and traverses FURTHER
	// than store root, this is an error
	if util.CountDotdots(pb.StringPtr()) > st.dots {
		return "", ErrInvalidKey
	}
	return pb.String(), nil
}

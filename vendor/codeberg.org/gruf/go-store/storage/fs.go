package storage

import (
	"os"
	"syscall"

	"codeberg.org/gruf/go-store/util"
)

const (
	// default file permission bits
	defaultDirPerms  = 0o755
	defaultFilePerms = 0o644

	// default file open flags
	defaultFileROFlags   = syscall.O_RDONLY
	defaultFileRWFlags   = syscall.O_CREAT | syscall.O_RDWR
	defaultFileLockFlags = syscall.O_RDONLY | syscall.O_CREAT
)

// NOTE:
// These functions are for opening storage files,
// not necessarily for e.g. initial setup (OpenFile)

// open should not be called directly.
func open(path string, flags int) (*os.File, error) {
	var fd int
	err := util.RetryOnEINTR(func() (err error) {
		fd, err = syscall.Open(path, flags, defaultFilePerms)
		return
	})
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}

// stat checks for a file on disk.
func stat(path string) (bool, error) {
	var stat syscall.Stat_t
	err := util.RetryOnEINTR(func() error {
		return syscall.Stat(path, &stat)
	})
	if err != nil {
		if err == syscall.ENOENT { //nolint
			err = nil
		}
		return false, err
	}
	return true, nil
}

// unlink removes a file (not dir!) on disk.
func unlink(path string) error {
	return util.RetryOnEINTR(func() error {
		return syscall.Unlink(path)
	})
}

// rmdir removes a dir (not file!) on disk.
func rmdir(path string) error {
	return util.RetryOnEINTR(func() error {
		return syscall.Rmdir(path)
	})
}

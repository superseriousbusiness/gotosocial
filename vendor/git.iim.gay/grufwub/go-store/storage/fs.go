package storage

import (
	"os"
	"syscall"

	"git.iim.gay/grufwub/go-store/util"
)

const (
	defaultDirPerms      = 0755
	defaultFilePerms     = 0644
	defaultFileROFlags   = syscall.O_RDONLY
	defaultFileRWFlags   = syscall.O_CREAT | syscall.O_RDWR
	defaultFileLockFlags = syscall.O_RDONLY | syscall.O_EXCL | syscall.O_CREAT
)

// NOTE:
// These functions are for opening storage files,
// not necessarily for e.g. initial setup (OpenFile)

// open should not be called directly
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

// stat checks for a file on disk
func stat(path string) (bool, error) {
	var stat syscall.Stat_t
	err := util.RetryOnEINTR(func() error {
		return syscall.Stat(path, &stat)
	})
	if err != nil {
		if err == syscall.ENOENT {
			err = nil
		}
		return false, err
	}
	return true, nil
}

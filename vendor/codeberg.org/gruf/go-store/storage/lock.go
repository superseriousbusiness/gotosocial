package storage

import (
	"os"
	"syscall"

	"codeberg.org/gruf/go-store/util"
)

// LockFile is our standard lockfile name.
const LockFile = "store.lock"

type LockableFile struct {
	*os.File
}

// OpenLock opens a lockfile at path.
func OpenLock(path string) (*LockableFile, error) {
	file, err := open(path, defaultFileLockFlags)
	if err != nil {
		return nil, err
	}
	return &LockableFile{file}, nil
}

func (f *LockableFile) Lock() error {
	return f.flock(syscall.LOCK_EX | syscall.LOCK_NB)
}

func (f *LockableFile) Unlock() error {
	return f.flock(syscall.LOCK_UN | syscall.LOCK_NB)
}

func (f *LockableFile) flock(how int) error {
	return util.RetryOnEINTR(func() error {
		return syscall.Flock(int(f.Fd()), how)
	})
}

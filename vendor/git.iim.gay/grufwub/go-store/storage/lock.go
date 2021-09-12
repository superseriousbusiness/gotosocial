package storage

import (
	"os"
	"syscall"

	"git.iim.gay/grufwub/go-store/util"
)

type lockableFile struct {
	*os.File
}

func openLock(path string) (*lockableFile, error) {
	file, err := open(path, defaultFileLockFlags)
	if err != nil {
		return nil, err
	}
	return &lockableFile{file}, nil
}

func (f *lockableFile) lock() error {
	return f.flock(syscall.LOCK_EX | syscall.LOCK_NB)
}

func (f *lockableFile) unlock() error {
	return f.flock(syscall.LOCK_UN | syscall.LOCK_NB)
}

func (f *lockableFile) flock(how int) error {
	return util.RetryOnEINTR(func() error {
		return syscall.Flock(int(f.Fd()), how)
	})
}

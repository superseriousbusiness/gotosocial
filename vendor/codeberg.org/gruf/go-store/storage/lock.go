package storage

import (
	"sync"
	"sync/atomic"
	"syscall"

	"codeberg.org/gruf/go-store/util"
)

// LockFile is our standard lockfile name.
const LockFile = "store.lock"

// Lock represents a filesystem lock to ensure only one storage instance open per path.
type Lock struct {
	fd int
	wg sync.WaitGroup
	st uint32
}

// OpenLock opens a lockfile at path.
func OpenLock(path string) (*Lock, error) {
	var fd int

	// Open the file descriptor at path
	err := util.RetryOnEINTR(func() (err error) {
		fd, err = syscall.Open(path, defaultFileLockFlags, defaultFilePerms)
		return
	})
	if err != nil {
		return nil, err
	}

	// Get a flock on the file descriptor
	err = util.RetryOnEINTR(func() error {
		return syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
	})
	if err != nil {
		return nil, errSwapUnavailable(err)
	}

	return &Lock{fd: fd}, nil
}

// Add will add '1' to the underlying sync.WaitGroup.
func (f *Lock) Add() {
	f.wg.Add(1)
}

// Done will decrememnt '1' from the underlying sync.WaitGroup.
func (f *Lock) Done() {
	f.wg.Done()
}

// Close will attempt to close the lockfile and file descriptor.
func (f *Lock) Close() error {
	var err error
	if atomic.CompareAndSwapUint32(&f.st, 0, 1) {
		// Wait until done
		f.wg.Wait()

		// Ensure gets closed
		defer syscall.Close(f.fd)

		// Call funlock on the file descriptor
		err = util.RetryOnEINTR(func() error {
			return syscall.Flock(f.fd, syscall.LOCK_UN|syscall.LOCK_NB)
		})
	}
	return err
}

// Closed will return whether this lockfile has been closed (and unlocked).
func (f *Lock) Closed() bool {
	return (atomic.LoadUint32(&f.st) == 1)
}

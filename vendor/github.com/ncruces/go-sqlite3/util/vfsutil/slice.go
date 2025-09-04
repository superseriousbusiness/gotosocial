package vfsutil

import (
	"io"

	"github.com/ncruces/go-sqlite3"
	"github.com/ncruces/go-sqlite3/vfs"
)

// SliceFile implements [vfs.File] with a byte slice.
// It is suitable for temporary files (such as [vfs.OPEN_TEMP_JOURNAL]),
// but not concurrency safe.
type SliceFile []byte

var (
	// Ensure these interfaces are implemented:
	_ vfs.FileSizeHint = &SliceFile{}
)

// ReadAt implements [io.ReaderAt].
func (f *SliceFile) ReadAt(b []byte, off int64) (n int, err error) {
	if d := *f; off < int64(len(d)) {
		n = copy(b, d[off:])
	}
	if n < len(b) {
		err = io.EOF
	}
	return
}

// WriteAt implements [io.WriterAt].
func (f *SliceFile) WriteAt(b []byte, off int64) (n int, err error) {
	d := *f
	if off > int64(len(d)) {
		d = append(d, make([]byte, off-int64(len(d)))...)
	}
	d = append(d[:off], b...)
	if len(d) > len(*f) {
		*f = d
	}
	return len(b), nil
}

// Size implements [vfs.File].
func (f *SliceFile) Size() (int64, error) {
	return int64(len(*f)), nil
}

// Truncate implements [vfs.File].
func (f *SliceFile) Truncate(size int64) error {
	if d := *f; size < int64(len(d)) {
		*f = d[:size]
	}
	return nil
}

// SizeHint implements [vfs.FileSizeHint].
func (f *SliceFile) SizeHint(size int64) error {
	if d := *f; size > int64(len(d)) {
		*f = append(d, make([]byte, size-int64(len(d)))...)
	}
	return nil
}

// Close implements [io.Closer].
func (*SliceFile) Close() error { return nil }

// Sync implements [vfs.File].
func (*SliceFile) Sync(flags vfs.SyncFlag) error { return nil }

// Lock implements [vfs.File].
func (*SliceFile) Lock(lock vfs.LockLevel) error {
	// notest // not concurrency safe
	return sqlite3.IOERR_LOCK
}

// Unlock implements [vfs.File].
func (*SliceFile) Unlock(lock vfs.LockLevel) error {
	// notest // not concurrency safe
	return sqlite3.IOERR_UNLOCK
}

// CheckReservedLock implements [vfs.File].
func (*SliceFile) CheckReservedLock() (bool, error) {
	// notest // not concurrency safe
	return false, sqlite3.IOERR_CHECKRESERVEDLOCK
}

// SectorSize implements [vfs.File].
func (*SliceFile) SectorSize() int {
	// notest // safe default
	return 0
}

// DeviceCharacteristics implements [vfs.File].
func (*SliceFile) DeviceCharacteristics() vfs.DeviceCharacteristic {
	return vfs.IOCAP_ATOMIC |
		vfs.IOCAP_SEQUENTIAL |
		vfs.IOCAP_SAFE_APPEND |
		vfs.IOCAP_POWERSAFE_OVERWRITE |
		vfs.IOCAP_SUBPAGE_READ
}

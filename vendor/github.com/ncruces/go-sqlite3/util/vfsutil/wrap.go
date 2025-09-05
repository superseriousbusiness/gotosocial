// Package vfsutil implements virtual filesystem utilities.
package vfsutil

import (
	"github.com/ncruces/go-sqlite3"
	"github.com/ncruces/go-sqlite3/vfs"
)

// UnwrapFile unwraps a [vfs.File],
// possibly implementing [vfs.FileUnwrap],
// to a concrete type.
func UnwrapFile[T vfs.File](f vfs.File) (_ T, _ bool) {
	for {
		switch t := f.(type) {
		default:
			return
		case T:
			return t, true
		case vfs.FileUnwrap:
			f = t.Unwrap()
		}
	}
}

// WrapOpen helps wrap [vfs.VFS].
func WrapOpen(f vfs.VFS, name string, flags vfs.OpenFlag) (file vfs.File, _ vfs.OpenFlag, err error) {
	if f, ok := f.(vfs.VFSFilename); name == "" && ok {
		return f.OpenFilename(nil, flags)
	}
	return f.Open(name, flags)
}

// WrapOpenFilename helps wrap [vfs.VFSFilename].
func WrapOpenFilename(f vfs.VFS, name *vfs.Filename, flags vfs.OpenFlag) (file vfs.File, _ vfs.OpenFlag, err error) {
	if f, ok := f.(vfs.VFSFilename); ok {
		return f.OpenFilename(name, flags)
	}
	return f.Open(name.String(), flags)
}

// WrapLockState helps wrap [vfs.FileLockState].
func WrapLockState(f vfs.File) vfs.LockLevel {
	if f, ok := f.(vfs.FileLockState); ok {
		return f.LockState()
	}
	return vfs.LOCK_EXCLUSIVE + 1 // UNKNOWN_LOCK
}

// WrapPersistWAL helps wrap [vfs.FilePersistWAL].
func WrapPersistWAL(f vfs.File) bool {
	if f, ok := f.(vfs.FilePersistWAL); ok {
		return f.PersistWAL()
	}
	return false
}

// WrapSetPersistWAL helps wrap [vfs.FilePersistWAL].
func WrapSetPersistWAL(f vfs.File, keepWAL bool) {
	if f, ok := f.(vfs.FilePersistWAL); ok {
		f.SetPersistWAL(keepWAL)
	}
}

// WrapPowersafeOverwrite helps wrap [vfs.FilePowersafeOverwrite].
func WrapPowersafeOverwrite(f vfs.File) bool {
	if f, ok := f.(vfs.FilePowersafeOverwrite); ok {
		return f.PowersafeOverwrite()
	}
	return false
}

// WrapSetPowersafeOverwrite helps wrap [vfs.FilePowersafeOverwrite].
func WrapSetPowersafeOverwrite(f vfs.File, psow bool) {
	if f, ok := f.(vfs.FilePowersafeOverwrite); ok {
		f.SetPowersafeOverwrite(psow)
	}
}

// WrapChunkSize helps wrap [vfs.FileChunkSize].
func WrapChunkSize(f vfs.File, size int) {
	if f, ok := f.(vfs.FileChunkSize); ok {
		f.ChunkSize(size)
	}
}

// WrapSizeHint helps wrap [vfs.FileSizeHint].
func WrapSizeHint(f vfs.File, size int64) error {
	if f, ok := f.(vfs.FileSizeHint); ok {
		return f.SizeHint(size)
	}
	return sqlite3.NOTFOUND
}

// WrapHasMoved helps wrap [vfs.FileHasMoved].
func WrapHasMoved(f vfs.File) (bool, error) {
	if f, ok := f.(vfs.FileHasMoved); ok {
		return f.HasMoved()
	}
	return false, sqlite3.NOTFOUND
}

// WrapOverwrite helps wrap [vfs.FileOverwrite].
func WrapOverwrite(f vfs.File) error {
	if f, ok := f.(vfs.FileOverwrite); ok {
		return f.Overwrite()
	}
	return sqlite3.NOTFOUND
}

// WrapSyncSuper helps wrap [vfs.FileSync].
func WrapSyncSuper(f vfs.File, super string) error {
	if f, ok := f.(vfs.FileSync); ok {
		return f.SyncSuper(super)
	}
	return sqlite3.NOTFOUND
}

// WrapCommitPhaseTwo helps wrap [vfs.FileCommitPhaseTwo].
func WrapCommitPhaseTwo(f vfs.File) error {
	if f, ok := f.(vfs.FileCommitPhaseTwo); ok {
		return f.CommitPhaseTwo()
	}
	return sqlite3.NOTFOUND
}

// WrapBeginAtomicWrite helps wrap [vfs.FileBatchAtomicWrite].
func WrapBeginAtomicWrite(f vfs.File) error {
	if f, ok := f.(vfs.FileBatchAtomicWrite); ok {
		return f.BeginAtomicWrite()
	}
	return sqlite3.NOTFOUND
}

// WrapCommitAtomicWrite helps wrap [vfs.FileBatchAtomicWrite].
func WrapCommitAtomicWrite(f vfs.File) error {
	if f, ok := f.(vfs.FileBatchAtomicWrite); ok {
		return f.CommitAtomicWrite()
	}
	return sqlite3.NOTFOUND
}

// WrapRollbackAtomicWrite helps wrap [vfs.FileBatchAtomicWrite].
func WrapRollbackAtomicWrite(f vfs.File) error {
	if f, ok := f.(vfs.FileBatchAtomicWrite); ok {
		return f.RollbackAtomicWrite()
	}
	return sqlite3.NOTFOUND
}

// WrapCheckpointStart helps wrap [vfs.FileCheckpoint].
func WrapCheckpointStart(f vfs.File) {
	if f, ok := f.(vfs.FileCheckpoint); ok {
		f.CheckpointStart()
	}
}

// WrapCheckpointDone helps wrap [vfs.FileCheckpoint].
func WrapCheckpointDone(f vfs.File) {
	if f, ok := f.(vfs.FileCheckpoint); ok {
		f.CheckpointDone()
	}
}

// WrapPragma helps wrap [vfs.FilePragma].
func WrapPragma(f vfs.File, name, value string) (string, error) {
	if f, ok := f.(vfs.FilePragma); ok {
		return f.Pragma(name, value)
	}
	return "", sqlite3.NOTFOUND
}

// WrapBusyHandler helps wrap [vfs.FilePragma].
func WrapBusyHandler(f vfs.File, handler func() bool) {
	if f, ok := f.(vfs.FileBusyHandler); ok {
		f.BusyHandler(handler)
	}
}

// WrapSharedMemory helps wrap [vfs.FileSharedMemory].
func WrapSharedMemory(f vfs.File) vfs.SharedMemory {
	if f, ok := f.(vfs.FileSharedMemory); ok {
		return f.SharedMemory()
	}
	return nil
}

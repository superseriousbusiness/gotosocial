package storage

import (
	"fmt"
	"io/fs"
	"os"
	"syscall"

	"codeberg.org/gruf/go-fastpath/v2"
	"codeberg.org/gruf/go-store/v2/util"
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

// walkDir traverses the dir tree of the supplied path, performing the supplied walkFn on each entry
func walkDir(pb *fastpath.Builder, path string, walkFn func(string, fs.DirEntry) error) error {
	// Read directory entries
	entries, err := readDir(path)
	if err != nil {
		return err
	}

	// frame represents a directory entry
	// walk-loop snapshot, taken when a sub
	// directory requiring iteration is found
	type frame struct {
		path    string
		entries []fs.DirEntry
	}

	// stack contains a list of held snapshot
	// frames, representing unfinished upper
	// layers of a directory structure yet to
	// be traversed.
	var stack []frame

outer:
	for {
		if len(entries) == 0 {
			if len(stack) == 0 {
				// Reached end
				break outer
			}

			// Pop frame from stack
			frame := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			// Update loop vars
			entries = frame.entries
			path = frame.path
		}

		for len(entries) > 0 {
			// Pop next entry from queue
			entry := entries[0]
			entries = entries[1:]

			// Pass to provided walk function
			if err := walkFn(path, entry); err != nil {
				return err
			}

			if entry.IsDir() {
				// Push current frame to stack
				stack = append(stack, frame{
					path:    path,
					entries: entries,
				})

				// Update current directory path
				path = pb.Join(path, entry.Name())

				// Read next directory entries
				next, err := readDir(path)
				if err != nil {
					return err
				}

				// Set next entries
				entries = next

				continue outer
			}
		}
	}

	return nil
}

// cleanDirs traverses the dir tree of the supplied path, removing any folders with zero children
func cleanDirs(path string) error {
	pb := util.GetPathBuilder()
	defer util.PutPathBuilder(pb)
	return cleanDir(pb, path, true)
}

// cleanDir performs the actual dir cleaning logic for the above top-level version.
func cleanDir(pb *fastpath.Builder, path string, top bool) error {
	// Get dir entries at path.
	entries, err := readDir(path)
	if err != nil {
		return err
	}

	// If no entries, delete dir.
	if !top && len(entries) == 0 {
		return rmdir(path)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Calculate directory path.
			dirPath := pb.Join(path, entry.Name())

			// Recursively clean sub-directory entries.
			if err := cleanDir(pb, dirPath, false); err != nil {
				fmt.Fprintf(os.Stderr, "[go-store/storage] error cleaning %s: %v", dirPath, err)
			}
		}
	}

	return nil
}

// readDir will open file at path, read the unsorted list of entries, then close.
func readDir(path string) ([]fs.DirEntry, error) {
	// Open file at path
	file, err := open(path, defaultFileROFlags)
	if err != nil {
		return nil, err
	}

	// Read directory entries
	entries, err := file.ReadDir(-1)

	// Done with file
	_ = file.Close()

	return entries, err
}

// open will open a file at the given path with flags and default file perms.
func open(path string, flags int) (*os.File, error) {
	var fd int
	err := retryOnEINTR(func() (err error) {
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
	err := retryOnEINTR(func() error {
		return syscall.Stat(path, &stat)
	})
	if err != nil {
		if err == syscall.ENOENT {
			// not-found is no error
			err = nil
		}
		return false, err
	}
	return true, nil
}

// unlink removes a file (not dir!) on disk.
func unlink(path string) error {
	return retryOnEINTR(func() error {
		return syscall.Unlink(path)
	})
}

// rmdir removes a dir (not file!) on disk.
func rmdir(path string) error {
	return retryOnEINTR(func() error {
		return syscall.Rmdir(path)
	})
}

// retryOnEINTR is a low-level filesystem function for retrying syscalls on O_EINTR received.
func retryOnEINTR(do func() error) error {
	for {
		err := do()
		if err == syscall.EINTR {
			continue
		}
		return err
	}
}

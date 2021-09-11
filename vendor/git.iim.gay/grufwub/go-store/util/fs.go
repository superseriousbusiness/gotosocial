package util

import (
	"io/fs"
	"os"
	"strings"
	"syscall"

	"git.iim.gay/grufwub/fastpath"
)

var dotdot = "../"

// CountDotdots returns the number of "dot-dots" (../) in a cleaned filesystem path
func CountDotdots(path string) int {
	if !strings.HasSuffix(path, dotdot) {
		return 0
	}
	return strings.Count(path, dotdot)
}

// WalkDir traverses the dir tree of the supplied path, performing the supplied walkFn on each entry
func WalkDir(pb *fastpath.Builder, path string, walkFn func(string, fs.DirEntry)) error {
	// Read supplied dir path
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// Iter entries
	for _, entry := range dirEntries {
		// Pass to walk fn
		walkFn(path, entry)

		// Recurse dir entries
		if entry.IsDir() {
			err = WalkDir(pb, pb.Join(path, entry.Name()), walkFn)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CleanDirs traverses the dir tree of the supplied path, removing any folders with zero children
func CleanDirs(path string) error {
	// Acquire builder
	pb := AcquirePathBuilder()
	defer ReleasePathBuilder(pb)

	// Get dir entries
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// Recurse dirs
	for _, entry := range entries {
		if entry.IsDir() {
			err := cleanDirs(pb, pb.Join(path, entry.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// cleanDirs performs the actual dir cleaning logic for the exported version
func cleanDirs(pb *fastpath.Builder, path string) error {
	// Get dir entries
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// If no entries, delete
	if len(entries) < 1 {
		return os.Remove(path)
	}

	// Recurse dirs
	for _, entry := range entries {
		if entry.IsDir() {
			err := cleanDirs(pb, pb.Join(path, entry.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// RetryOnEINTR is a low-level filesystem function for retrying syscalls on O_EINTR received
func RetryOnEINTR(do func() error) error {
	for {
		err := do()
		if err == syscall.EINTR {
			continue
		}
		return err
	}
}

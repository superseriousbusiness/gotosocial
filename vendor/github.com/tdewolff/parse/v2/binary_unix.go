//go:build unix

package parse

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"syscall"
)

type binaryReaderMmap struct {
	data []byte
}

func newBinaryReaderMmap(filename string) (*binaryReaderMmap, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size := fi.Size()
	if size == 0 {
		// Treat (size == 0) as a special case, avoiding the syscall, since
		// "man 2 mmap" says "the length... must be greater than 0".
		//
		// As we do not call syscall.Mmap, there is no need to call
		// runtime.SetFinalizer to enforce a balancing syscall.Munmap.
		return &binaryReaderMmap{
			data: make([]byte, 0),
		}, nil
	} else if size < 0 {
		return nil, fmt.Errorf("mmap: file %q has negative size", filename)
	} else if size != int64(int(size)) {
		return nil, fmt.Errorf("mmap: file %q is too large", filename)
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	r := &binaryReaderMmap{data}
	runtime.SetFinalizer(r, (*binaryReaderMmap).Close)
	return r, nil
}

// Close closes the reader.
func (r *binaryReaderMmap) Close() error {
	if r.data == nil {
		return nil
	} else if len(r.data) == 0 {
		r.data = nil
		return nil
	}
	data := r.data
	r.data = nil
	runtime.SetFinalizer(r, nil)
	return syscall.Munmap(data)
}

// Len returns the length of the underlying memory-mapped file.
func (r *binaryReaderMmap) Len() int {
	return len(r.data)
}

func (r *binaryReaderMmap) Bytes(n int, off int64) ([]byte, error) {
	if r.data == nil {
		return nil, errors.New("mmap: closed")
	} else if off < 0 || int64(len(r.data)) < off {
		return nil, fmt.Errorf("mmap: invalid offset %d", off)
	} else if int64(len(r.data)-n) < off {
		return r.data[off:len(r.data):len(r.data)], io.EOF
	}
	return r.data[off : off+int64(n) : off+int64(n)], nil
}

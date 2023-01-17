/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package iotools

import (
	"io"
	"os"
)

// ReadFnCloser takes an io.Reader and wraps it to use the provided function to implement io.Closer.
func ReadFnCloser(r io.Reader, close func() error) io.ReadCloser {
	return &readFnCloser{
		Reader: r,
		close:  close,
	}
}

type readFnCloser struct {
	io.Reader
	close func() error
}

func (r *readFnCloser) Close() error {
	return r.close()
}

// WriteFnCloser takes an io.Writer and wraps it to use the provided function to implement io.Closer.
func WriteFnCloser(w io.Writer, close func() error) io.WriteCloser {
	return &writeFnCloser{
		Writer: w,
		close:  close,
	}
}

type writeFnCloser struct {
	io.Writer
	close func() error
}

func (r *writeFnCloser) Close() error {
	return r.close()
}

// SilentReader wraps an io.Reader to silence any
// error output during reads. Instead they are stored
// and accessible (not concurrency safe!) via .Error().
type SilentReader struct {
	io.Reader
	err error
}

// SilenceReader wraps an io.Reader within SilentReader{}.
func SilenceReader(r io.Reader) *SilentReader {
	return &SilentReader{Reader: r}
}

func (r *SilentReader) Read(b []byte) (int, error) {
	n, err := r.Reader.Read(b)
	if err != nil {
		// Store error for now
		if r.err == nil {
			r.err = err
		}

		// Pretend we're happy
		// to continue reading.
		n = len(b)
	}
	return n, nil
}

func (r *SilentReader) Error() error {
	return r.err
}

// SilentWriter wraps an io.Writer to silence any
// error output during writes. Instead they are stored
// and accessible (not concurrency safe!) via .Error().
type SilentWriter struct {
	io.Writer
	err error
}

// SilenceWriter wraps an io.Writer within SilentWriter{}.
func SilenceWriter(w io.Writer) *SilentWriter {
	return &SilentWriter{Writer: w}
}

func (w *SilentWriter) Write(b []byte) (int, error) {
	n, err := w.Writer.Write(b)
	if err != nil {
		// Store error for now
		if w.err == nil {
			w.err = err
		}

		// Pretend we're happy
		// to continue writing.
		n = len(b)
	}
	return n, nil
}

func (w *SilentWriter) Error() error {
	return w.err
}

func StreamReadFunc(read func(io.Reader) error) io.Writer {
	// In-memory stream.
	pr, pw := io.Pipe()

	go func() {
		var err error

		defer func() {
			// Always pass along error.
			pr.CloseWithError(err)
		}()

		// Start reading.
		err = read(pr)
	}()

	return pw
}

func StreamWriteFunc(write func(io.Writer) error) io.Reader {
	// In-memory stream.
	pr, pw := io.Pipe()

	go func() {
		var err error

		defer func() {
			// Always pass along error.
			pw.CloseWithError(err)
		}()

		// Start writing.
		err = write(pw)
	}()

	return pr
}

type tempFileSeeker struct {
	io.Reader
	io.Seeker
	tmp *os.File
}

func (tfs *tempFileSeeker) Close() error {
	tfs.tmp.Close()
	return os.Remove(tfs.tmp.Name())
}

// TempFileSeeker converts the provided Reader into a ReadSeekCloser
// by using an underlying temporary file. Callers should call the Close
// function when they're done with the TempFileSeeker, to release +
// clean up the temporary file.
func TempFileSeeker(r io.Reader) (io.ReadSeekCloser, error) {
	tmp, err := os.CreateTemp(os.TempDir(), "gotosocial-")
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(tmp, r); err != nil {
		return nil, err
	}

	return &tempFileSeeker{
		Reader: tmp,
		Seeker: tmp,
		tmp:    tmp,
	}, nil
}

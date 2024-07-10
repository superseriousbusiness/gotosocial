package iotools

import "io"

// AtEOF returns true when reader at EOF,
// this is checked with a 0 length read.
func AtEOF(r io.Reader) bool {
	_, err := r.Read(nil)
	return (err == io.EOF)
}

// GetReadCloserLimit attempts to cast io.Reader to access its io.LimitedReader with limit.
func GetReaderLimit(r io.Reader) (*io.LimitedReader, int64) {
	lr, ok := r.(*io.LimitedReader)
	if !ok {
		return nil, -1
	}
	return lr, lr.N
}

// UpdateReaderLimit attempts to  update the limit of a reader for existing, newly wrapping if necessary.
func UpdateReaderLimit(r io.Reader, limit int64) (*io.LimitedReader, int64) {
	lr, ok := r.(*io.LimitedReader)
	if !ok {
		lr = &io.LimitedReader{r, limit}
		return lr, limit
	}

	if limit < lr.N {
		// Update existing.
		lr.N = limit
	}

	return lr, lr.N
}

// GetReadCloserLimit attempts to unwrap io.ReadCloser to access its io.LimitedReader with limit.
func GetReadCloserLimit(rc io.ReadCloser) (*io.LimitedReader, int64) {
	rct, ok := rc.(*ReadCloserType)
	if !ok {
		return nil, -1
	}
	lr, ok := rct.Reader.(*io.LimitedReader)
	if !ok {
		return nil, -1
	}
	return lr, lr.N
}

// UpdateReadCloserLimit attempts to update the limit of a readcloser for existing, newly wrapping if necessary.
func UpdateReadCloserLimit(rc io.ReadCloser, limit int64) (io.ReadCloser, *io.LimitedReader, int64) {

	// Check for our wrapped ReadCloserType.
	if rct, ok := rc.(*ReadCloserType); ok {

		// Attempt to update existing wrapped limit reader.
		if lr, ok := rct.Reader.(*io.LimitedReader); ok {

			if limit < lr.N {
				// Update existing.
				lr.N = limit
			}

			return rct, lr, lr.N
		}

		// Wrap the reader type with new limit.
		lr := &io.LimitedReader{rct.Reader, limit}
		rct.Reader = lr

		return rct, lr, lr.N
	}

	// Wrap separated types.
	rct := &ReadCloserType{
		Reader: rc,
		Closer: rc,
	}

	// Wrap separated reader part with limit.
	lr := &io.LimitedReader{rct.Reader, limit}
	rct.Reader = lr

	return rct, lr, lr.N
}

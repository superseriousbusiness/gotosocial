package fastpath

import (
	"unsafe"
)

// allocate these just once
var (
	dot    = []byte(dotStr)
	dotStr = "."
)

type Builder struct {
	B   []byte // B is the underlying byte buffer
	dd  int    // pos of last '..' appended to builder
	abs bool   // abs stores whether path passed to first .Append() is absolute
	set bool   // set stores whether b.abs has been set i.e. not first call to .Append()
}

// NewBuilder returns a new Builder object using the
// supplied byte slice as the underlying buffer
func NewBuilder(b []byte) Builder {
	if b != nil {
		b = b[:0]
	}
	return Builder{
		B:  b,
		dd: 0,

		abs: false,
		set: false,
	}
}

// Reset resets the Builder object
func (b *Builder) Reset() {
	b.B = b.B[:0]
	b.dd = 0
	b.abs = false
	b.set = false
}

// Len returns the number of accumulated bytes in the Builder
func (b *Builder) Len() int {
	return len(b.B)
}

// Cap returns the capacity of the underlying Builder buffer
func (b *Builder) Cap() int {
	return cap(b.B)
}

// Bytes returns the accumulated path bytes.
func (b *Builder) Bytes() []byte {
	if len(b.B) < 1 {
		return dot
	}
	return b.B
}

// String returns the accumulated path string.
func (b *Builder) String() string {
	if len(b.B) < 1 {
		return dotStr
	}
	return string(b.B)
}

// StringPtr returns a ptr to the accumulated path string.
//
// Please note the underlying byte slice for this string is
// tied to the builder, so any changes will result in the
// returned string changing. Consider using .String() if
// this is undesired behaviour.
func (b *Builder) StringPtr() string {
	if len(b.B) < 1 {
		return dotStr
	}
	return *(*string)(unsafe.Pointer(&b.B))
}

// Absolute returns whether current path is absolute (not relative)
func (b *Builder) Absolute() bool {
	return b.abs
}

// SetAbsolute converts the current path to / from absolute
func (b *Builder) SetAbsolute(val bool) {
	if !b.set {
		if val {
			// .Append() has not been
			// called, add a '/' and set abs
			b.Guarantee(1)
			b.appendByte('/')
			b.abs = true
		}

		// Set as having been set
		b.set = true
		return
	}

	if !val && b.abs {
		// Already set and absolute. Update
		b.abs = false

		// If not empty (i.e. not just '/'),
		// then shift bytes 1 left
		if len(b.B) > 1 {
			copy(b.B, b.B[1:])
		}

		// Truncate 1 byte. In the case of empty,
		// i.e. just '/' then it will drop this
		b.truncate(1)
	} else if val && !b.abs {
		// Already set but NOT abs. Update
		b.abs = true

		// Guarantee 1 byte available
		b.Guarantee(1)

		// If empty, just append '/'
		if len(b.B) < 1 {
			b.appendByte('/')
			return
		}

		// Increase length
		l := len(b.B)
		b.B = b.B[:l+1]

		// Shift bytes 1 right
		copy(b.B[1:], b.B[:l])

		// Set first byte '/'
		b.B[0] = '/'
	}
}

// Append adds and cleans the supplied path bytes to the
// builder's internal buffer, growing the buffer if necessary
// to accomodate the extra path length
func (b *Builder) Append(p []byte) {
	b.AppendString(*(*string)(unsafe.Pointer(&p)))
}

// AppendString adds and cleans the supplied path string to the
// builder's internal buffer, growing the buffer if necessary
// to accomodate the extra path length
func (b *Builder) AppendString(path string) {
	defer func() {
		// If buffer is empty, and an absolute
		// path, ensure it starts with a '/'
		if len(b.B) < 1 && b.abs {
			b.appendByte('/')
		}
	}()

	// Empty path, nothing to do
	if len(path) == 0 {
		return
	}

	// Guarantee at least the total length
	// of supplied path available in the buffer
	b.Guarantee(len(path))

	// Try store if absolute
	if !b.set {
		b.abs = len(path) > 0 && path[0] == '/'
		b.set = true
	}

	i := 0
	for i < len(path) {
		switch {
		// Empty path segment
		case path[i] == '/':
			i++

		// Singular '.' path segment, treat as empty
		case path[i] == '.' && (i+1 == len(path) || path[i+1] == '/'):
			i++

		// Backtrack segment
		case path[i] == '.' && path[i+1] == '.' && (i+2 == len(path) || path[i+2] == '/'):
			i += 2

			switch {
			// Check if it's possible to backtrack with
			// our current state of the buffer. i.e. is
			// our buffer length longer than the last
			// '..' we placed?
			case len(b.B) > b.dd:
				b.backtrack()
				// b.cp = b.lp
				// b.lp = 0

			// If we reached here, need to check if
			// we can append '..' to the path buffer,
			// which is ONLY when path is NOT absolute
			case !b.abs:
				if len(b.B) > 0 {
					b.appendByte('/')
				}
				b.appendByte('.')
				b.appendByte('.')
				b.dd = len(b.B)
				// b.lp = lp - 2
				// b.cp = b.dd
			}

		default:
			if (b.abs && len(b.B) != 1) || (!b.abs && len(b.B) > 0) {
				b.appendByte('/')
			}
			// b.lp = b.cp
			// b.cp = len(b.B)
			i += b.appendSlice(path[i:])
		}
	}
}

// Clean creates the shortest possible functional equivalent
// to the supplied path, resetting the builder before performing
// this operation. The builder object is NOT reset after return
func (b *Builder) Clean(path string) string {
	b.Reset()
	b.AppendString(path)
	return b.String()
}

// Join connects and cleans multiple paths, resetting the builder before
// performing this operation and returning the shortest possible combination
// of all the supplied paths. The builder object is NOT reset after return
func (b *Builder) Join(base string, paths ...string) string {
	b.Reset()
	b.AppendString(base)
	size := len(base)
	for i := 0; i < len(paths); i++ {
		b.AppendString(paths[i])
		size += len(paths[i])
	}
	if size < 1 {
		return ""
	} else if len(b.B) < 1 {
		return dotStr
	}
	return string(b.B)
}

// Guarantee ensures there is at least the requested size
// free bytes available in the buffer, reallocating if necessary
func (b *Builder) Guarantee(size int) {
	if size > cap(b.B)-len(b.B) {
		nb := make([]byte, 2*cap(b.B)+size)
		copy(nb, b.B)
		b.B = nb[:len(b.B)]
	}
}

// Truncate reduces the length of the buffer by the requested
// number of bytes. If the builder is set to absolute, the first
// byte (i.e. '/') will never be truncated
func (b *Builder) Truncate(size int) {
	// If absolute and just '/', do nothing
	if b.abs && len(b.B) == 1 {
		return
	}

	// Truncate requested bytes
	b.truncate(size)
}

// truncate reduces the length of the buffer by the requested
// size, no sanity checks are performed
func (b *Builder) truncate(size int) {
	b.B = b.B[:len(b.B)-size]
}

// appendByte appends the supplied byte to the end of
// the buffer. appending is achieved by continually reslicing the
// buffer and setting the next byte-at-index, this is safe as guarantee()
// will have been called beforehand
func (b *Builder) appendByte(c byte) {
	b.B = b.B[:len(b.B)+1]
	b.B[len(b.B)-1] = c
}

// appendSlice appends the supplied string slice to
// the end of the buffer and returns the number of indices
// we were able to iterate before hitting a path separator '/'.
// appending is achieved by continually reslicing the buffer
// and setting the next byte-at-index, this is safe as guarantee()
// will have been called beforehand
func (b *Builder) appendSlice(slice string) int {
	i := 0
	for i < len(slice) && slice[i] != '/' {
		b.B = b.B[:len(b.B)+1]
		b.B[len(b.B)-1] = slice[i]
		i++
	}
	return i
}

// backtrack reduces the end of the buffer back to the last
// separating '/', or end of buffer
func (b *Builder) backtrack() {
	b.B = b.B[:len(b.B)-1]

	for len(b.B)-1 > b.dd && b.B[len(b.B)-1] != '/' {
		b.B = b.B[:len(b.B)-1]
	}

	if len(b.B) > 0 {
		b.B = b.B[:len(b.B)-1]
	}
}

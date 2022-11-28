package fastpath

import (
	"unsafe"
)

// Clean: see Builder.Clean(). Analogous to path.Clean().
func Clean(path string) string {
	return (&Builder{}).Clean(path)
}

// Join: see Builder.Join(). Analogous to path.Join().
func Join(elems ...string) string {
	return (&Builder{}).Join(elems...)
}

// Builder provides a means of cleaning and joining system paths,
// while retaining a singular underlying byte buffer for performance.
type Builder struct {
	// B is the underlying byte buffer
	B []byte

	dd  int  // pos of last '..' appended to builder
	abs bool // abs stores whether path passed to first .Append() is absolute
	set bool // set stores whether b.abs has been set i.e. not first call to .Append()
}

// Reset resets the Builder object
func (b *Builder) Reset() {
	b.B = b.B[:0]
	b.dd = 0
	b.abs = false
	b.set = false
}

// Len returns the number of accumulated bytes in the Builder
func (b Builder) Len() int {
	return len(b.B)
}

// Cap returns the capacity of the underlying Builder buffer
func (b Builder) Cap() int {
	return cap(b.B)
}

// Bytes returns the accumulated path bytes.
func (b Builder) Bytes() []byte {
	return b.B
}

// String returns the accumulated path string.
func (b Builder) String() string {
	return *(*string)(unsafe.Pointer(&b.B))
}

// Absolute returns whether current path is absolute (not relative).
func (b Builder) Absolute() bool {
	return b.abs
}

// SetAbsolute converts the current path to-or-from absolute.
func (b *Builder) SetAbsolute(enabled bool) {
	if !b.set {
		// Ensure 1B avail
		b.Guarantee(1)

		if enabled {
			// Set empty 'abs'
			b.appendByte('/')
			b.abs = true
		} else {
			// Set empty 'rel'
			b.appendByte('.')
			b.abs = false
		}

		b.set = true
		return
	}

	if !enabled && b.abs {
		// set && absolute
		// -> update
		b.abs = false

		// If empty, set to '.' (empty rel path)
		if len(b.B) == 0 || (len(b.B) == 1 && b.B[0] == '/') {
			b.Guarantee(1)
			b.B = b.B[:1]
			b.B[0] = '.'
			return
		}

		if b.B[0] != '/' {
			// No need to change
			return
		}

		if len(b.B) > 1 {
			// Shift bytes 1 left
			copy(b.B, b.B[1:])
		}

		// and drop the '/' prefix'
		b.B = b.B[:len(b.B)-1]
	} else if enabled && !b.abs {
		// set && !absolute
		// -> update
		b.abs = true

		// Ensure 1B avail
		b.Guarantee(1)

		// If empty, set to '/' (empty abs path)
		if len(b.B) == 0 || (len(b.B) == 1 && b.B[0] == '.') {
			b.Guarantee(1)
			b.B = b.B[:1]
			b.B[0] = '/'
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

// AppendBytes adds and cleans the supplied path bytes to the
// builder's internal buffer, growing the buffer if necessary
// to accomodate the extra path length.
func (b *Builder) AppendBytes(path []byte) {
	if len(path) == 0 {
		return
	}
	b.Guarantee(len(path) + 1)
	b.append(*(*string)(unsafe.Pointer(&b)))
}

// Append adds and cleans the supplied path string to the
// builder's internal buffer, growing the buffer if necessary
// to accomodate the extra path length.
func (b *Builder) Append(path string) {
	if len(path) == 0 {
		return
	}
	b.Guarantee(len(path) + 1)
	b.append(path)
}

// Clean creates the shortest possible functional equivalent
// to the supplied path, resetting the builder before performing
// this operation. The builder object is NOT reset after return.
func (b *Builder) Clean(path string) string {
	if path == "" {
		return "."
	}
	b.Reset()
	b.Guarantee(len(path) + 1)
	b.append(path)
	return string(b.B)
}

// Join connects and cleans multiple paths, resetting the builder before
// performing this operation and returning the shortest possible combination
// of all the supplied paths. The builder object is NOT reset after return.
func (b *Builder) Join(elems ...string) string {
	var size int
	for _, elem := range elems {
		size += len(elem)
	}
	if size == 0 {
		return ""
	}
	b.Reset()
	b.Guarantee(size + 1)
	for _, elem := range elems {
		if elem == "" {
			continue
		}
		b.append(elem)
	}
	return string(b.B)
}

// append performs the main logic of 'Append()' but without an empty path check or preallocation.
func (b *Builder) append(path string) {
	if !b.set {
		// Set if absolute or not
		b.abs = path[0] == '/'
		b.set = true
	} else if !b.abs && len(b.B) == 1 && b.B[0] == '.' {
		// Empty non-abs path segment, drop
		// the period so not prefixed './'
		b.B = b.B[:0]
	}

	for i := 0; i < len(path); {
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
			}

		default:
			if (b.abs && len(b.B) != 1) || (!b.abs && len(b.B) > 0) {
				// Append path separator
				b.appendByte('/')
			}

			// Append slice up to next '/'
			i += b.appendSlice(path[i:])
		}
	}

	if len(b.B) > 0 {
		return
	}

	if b.abs {
		// Empty absolute path => /
		b.appendByte('/')
	} else {
		// Empty relative path => .
		b.appendByte('.')
	}
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
// number of bytes. If the byte slice is *effectively* empty,
// i.e. absolute and "/" or relative and ".", it won't be truncated.
func (b *Builder) Truncate(size int) {
	if len(b.B) == 0 {
		return
	}

	if len(b.B) == 1 && ((b.abs && b.B[0] == '/') ||
		(!b.abs && b.B[0] == '.')) {
		// *effectively* empty
		return
	}

	// Truncate requested bytes
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

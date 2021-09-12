package fastpath

import (
	"unsafe"
)

// allocate this just once
var dot = []byte(".")

type Builder struct {
	noCopy noCopy

	b  []byte // b is the underlying byte buffer
	dd int    // pos of last '..' appended to builder

	abs bool // abs stores whether path passed to first .Append() is absolute
	set bool // set stores whether b.abs has been set i.e. not first call to .Append()

	// lp int // pos of beginning of previous path segment
	// cp int // pos of beginning of current path segment
}

// NewBuilder returns a new Builder object using the supplied byte
// slice as the underlying buffer
func NewBuilder(b []byte) Builder {
	if b != nil {
		b = b[:0]
	}
	return Builder{
		noCopy: noCopy{},

		b:  b,
		dd: 0,

		abs: false,
		set: false,
	}
}

// Reset resets the Builder object
func (b *Builder) Reset() {
	b.b = b.b[:0]
	b.dd = 0
	b.abs = false
	b.set = false
	// b.lp = 0
	// b.cp = 0
}

// Len returns the number of accumulated bytes in the Builder
func (b *Builder) Len() int {
	return len(b.b)
}

// Cap returns the capacity of the underlying Builder buffer
func (b *Builder) Cap() int {
	return cap(b.b)
}

// Bytes returns the accumulated path bytes.
func (b *Builder) Bytes() []byte {
	if b.Len() < 1 {
		return dot
	}
	return b.b
}

// String returns the accumulated path string.
func (b *Builder) String() string {
	if b.Len() < 1 {
		return string(dot)
	}
	return string(b.b)
}

// StringPtr returns a ptr to the accumulated path string.
//
// Please note the underlying byte slice for this string is
// tied to the builder, so any changes will result in the
// returned string changing. Consider using .String() if
// this is undesired behaviour.
func (b *Builder) StringPtr() string {
	if b.Len() < 1 {
		return *(*string)(unsafe.Pointer(&dot))
	}
	return *(*string)(unsafe.Pointer(&b.b))
}

// Basename returns the base name of the accumulated path string
// func (b *Builder) Basename() string {
// 	if b.cp >= b.Len() {
// 		return dot
// 	}
// 	return deepcopy(b.string()[b.cp:])
// }

// BasenamePtr returns a ptr to the base name of the accumulated
// path string.
//
// Please note the underlying byte slice for this string is
// tied to the builder, so any changes will result in the
// returned string changing. Consider using .NewString() if
// this is undesired behaviour.
// func (b *Builder) BasenamePtr() string {
// 	if b.cp >= b.Len() {
// 		return dot
// 	}
// 	return b.string()[b.cp:]
// }

// Dirname returns the dir path of the accumulated path string
// func (b *Builder) Dirname() string {
// 	if b.cp < 1 || b.cp-1 >= b.Len() {
// 		return dot
// 	}
// 	return deepcopy(b.string()[:b.cp-1])
// }

// DirnamePtr returns a ptr to the dir path of the accumulated
// path string.
//
// Please note the underlying byte slice for this string is
// tied to the builder, so any changes will result in the
// returned string changing. Consider using .NewString() if
// this is undesired behaviour.
// func (b *Builder) DirnamePtr() string {
// 	if b.cp < 1 || b.cp-1 >= b.Len() {
// 		return dot
// 	}
// 	return b.String()[:b.cp-1]
// }

func (b *Builder) Absolute() bool {
	return b.abs
}

func (b *Builder) SetAbsolute(val bool) {
	if !b.set {
		if val {
			// .Append() has not be called,
			// add a '/' and set abs
			b.guarantee(1)
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
		if b.Len() > 1 {
			copy(b.b, b.b[1:])
		}

		// Truncate 1 byte. In the case of empty,
		// i.e. just '/' then it will drop this
		b.truncate(1)
	} else if val && !b.abs {
		// Already set but NOT abs. Update
		b.abs = true

		// Guarantee 1 byte available
		b.guarantee(1)

		// If empty, just append '/'
		if b.Len() < 1 {
			b.appendByte('/')
			return
		}

		// Increase length
		l := b.Len()
		b.b = b.b[:l+1]

		// Shift bytes 1 right
		copy(b.b[1:], b.b[:l])

		// Set first byte '/'
		b.b[0] = '/'
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
		// If buffer is empty, and an absolute path,
		// ensure it starts with a '/'
		if b.Len() < 1 && b.abs {
			b.appendByte('/')
		}
	}()

	// Empty path, nothing to do
	if len(path) == 0 {
		return
	}

	// Guarantee at least the total length
	// of supplied path available in the buffer
	b.guarantee(len(path))

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
			case b.Len() > b.dd:
				b.backtrack()
				// b.cp = b.lp
				// b.lp = 0

			// If we reached here, need to check if
			// we can append '..' to the path buffer,
			// which is ONLY when path is NOT absolute
			case !b.abs:
				if b.Len() > 0 {
					b.appendByte('/')
				}
				b.appendByte('.')
				b.appendByte('.')
				b.dd = b.Len()
				// b.lp = lp - 2
				// b.cp = b.dd
			}

		default:
			if (b.abs && b.Len() != 1) || (!b.abs && b.Len() > 0) {
				b.appendByte('/')
			}
			// b.lp = b.cp
			// b.cp = b.Len()
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
	empty := (len(base) < 1)
	b.Reset()
	b.AppendString(base)
	for _, path := range paths {
		b.AppendString(path)
		empty = empty && (len(path) < 1)
	}
	if empty {
		return ""
	}
	return b.String()
}

// Guarantee ensures there is at least the requested size
// free bytes available in the buffer, reallocating if
// necessary
func (b *Builder) Guarantee(size int) {
	b.guarantee(size)
}

// Truncate reduces the length of the buffer by the requested
// number of bytes. If the builder is set to absolute, the first
// byte (i.e. '/') will never be truncated
func (b *Builder) Truncate(size int) {
	// If absolute and just '/', do nothing
	if b.abs && b.Len() == 1 {
		return
	}

	// Truncate requested bytes
	b.truncate(size)
}

// truncate reduces the length of the buffer by the requested size,
// no sanity checks are performed
func (b *Builder) truncate(size int) {
	b.b = b.b[:b.Len()-size]
}

// guarantee ensures there is at least the requested size
// free bytes available in the buffer, reallocating if necessary.
// no sanity checks are performed
func (b *Builder) guarantee(size int) {
	if size > b.Cap()-b.Len() {
		nb := make([]byte, 2*b.Cap()+size)
		copy(nb, b.b)
		b.b = nb[:b.Len()]
	}
}

// appendByte appends the supplied byte to the end of
// the buffer. appending is achieved by continually reslicing the
// buffer and setting the next byte-at-index, this is safe as guarantee()
// will have been called beforehand
func (b *Builder) appendByte(c byte) {
	b.b = b.b[:b.Len()+1]
	b.b[b.Len()-1] = c
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
		b.b = b.b[:b.Len()+1]
		b.b[b.Len()-1] = slice[i]
		i++
	}
	return i
}

// backtrack reduces the end of the buffer back to the last
// separating '/', or end of buffer
func (b *Builder) backtrack() {
	b.b = b.b[:b.Len()-1]

	for b.Len()-1 > b.dd && b.b[b.Len()-1] != '/' {
		b.b = b.b[:b.Len()-1]
	}

	if b.Len() > 0 {
		b.b = b.b[:b.Len()-1]
	}
}

type noCopy struct{}

func (n *noCopy) Lock()   {}
func (n *noCopy) Unlock() {}

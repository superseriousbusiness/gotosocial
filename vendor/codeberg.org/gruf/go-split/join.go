package split

import (
	"strconv"
	"time"
	"unsafe"

	"codeberg.org/gruf/go-bytesize"
)

// joinFunc will join given slice of elements, using the passed function to append each element at index
// from the slice, forming a combined comma-space separated string. Passed size is for buffer preallocation.
func JoinFunc[T any](slice []T, each func(buf []byte, value T) []byte, size int) string {

	// Move nil check
	// outside main loop.
	if each == nil {
		panic("nil func")
	}

	// Catch easiest case
	if len(slice) == 0 {
		return ""
	}

	// Preallocate string buffer (size + commas)
	buf := make([]byte, 0, size+len(slice)-1)

	for _, value := range slice {
		// Append each item
		buf = each(buf, value)
		buf = append(buf, ',', ' ')
	}

	// Drop final comma-space
	buf = buf[:len(buf)-2]

	// Directly cast buf to string
	data := unsafe.SliceData(buf)
	return unsafe.String(data, len(buf))
}

// JoinStrings will pass string slice to JoinFunc(), quoting where
// necessary and combining into a single comma-space separated string.
func JoinStrings[String ~string](slice []String) string {
	var size int
	for _, str := range slice {
		size += len(str)
	}
	return JoinFunc(slice, func(buf []byte, value String) []byte {
		return appendQuote(buf, string(value))
	}, size)
}

// JoinBools will pass bool slice to JoinFunc(), formatting
// and combining into a single comma-space separated string.
func JoinBools[Bool ~bool](slice []Bool) string {
	return JoinFunc(slice, func(buf []byte, value Bool) []byte {
		return strconv.AppendBool(buf, bool(value))
	}, len(slice)*5 /* len("false") */)
}

// JoinInts will pass signed integer slice to JoinFunc(), formatting
// and combining into a single comma-space separated string.
func JoinInts[Int Signed](slice []Int) string {
	return JoinFunc(slice, func(buf []byte, value Int) []byte {
		return strconv.AppendInt(buf, int64(value), 10)
	}, len(slice)*20) // max signed int str len
}

// JoinUints will pass unsigned integer slice to JoinFunc(),
// formatting and combining into a single comma-space separated string.
func JoinUints[Uint Unsigned](slice []Uint) string {
	return JoinFunc(slice, func(buf []byte, value Uint) []byte {
		return strconv.AppendUint(buf, uint64(value), 10)
	}, len(slice)*20) // max unsigned int str len
}

// JoinFloats will pass float slice to JoinFunc(), formatting
// and combining into a single comma-space separated string.
func JoinFloats[Float_ Float](slice []Float_) string {
	bits := int(unsafe.Sizeof(Float_(0)) * 8) // param type bits
	return JoinFunc(slice, func(buf []byte, value Float_) []byte {
		return strconv.AppendFloat(buf, float64(value), 'g', -1, bits)
	}, len(slice)*20) // max signed int str len (it's a good guesstimate)
}

// JoinSizes will pass byte size slice to JoinFunc(), formatting
// and combining into a single comma-space separated string.
func JoinSizes(slice []bytesize.Size) string {
	const iecLen = 7 // max IEC string length
	return JoinFunc(slice, func(buf []byte, value bytesize.Size) []byte {
		return value.AppendFormatIEC(buf)
	}, len(slice)*iecLen)
}

// JoinDurations will pass duration slice to JoinFunc(), formatting
// and combining into a single comma-space separated string.
func JoinDurations(slice []time.Duration) string {
	const durLen = 10 // max duration string length
	return JoinFunc(slice, func(buf []byte, value time.Duration) []byte {
		return append(buf, value.String()...)
	}, len(slice)*durLen)
}

// JoinTimes will pass time slice to JoinFunc(), formatting
// and combining into a single comma-space separated string.
func JoinTimes(slice []time.Time, format string) string {
	return JoinFunc(slice, func(buf []byte, value time.Time) []byte {
		return value.AppendFormat(buf, format)
	}, len(slice)*len(format))
}

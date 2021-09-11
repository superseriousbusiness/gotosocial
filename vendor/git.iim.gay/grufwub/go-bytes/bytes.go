package bytes

import (
	"bytes"
	"reflect"
	"unsafe"
)

var (
	_ Bytes = &Buffer{}
	_ Bytes = bytesType{}
)

// Bytes defines a standard way of retrieving the content of a
// byte buffer of some-kind.
type Bytes interface {
	// Bytes returns the byte slice content
	Bytes() []byte

	// String returns byte slice cast directly to string, this
	// will cause an allocation but comes with the safety of
	// being an immutable Go string
	String() string

	// StringPtr returns byte slice cast to string via the unsafe
	// package. This comes with the same caveats of accessing via
	// .Bytes() in that the content is liable change and is NOT
	// immutable, despite being a string type
	StringPtr() string
}

type bytesType []byte

func (b bytesType) Bytes() []byte {
	return b
}

func (b bytesType) String() string {
	return string(b)
}

func (b bytesType) StringPtr() string {
	return BytesToString(b)
}

// ToBytes casts the provided byte slice as the simplest possible
// Bytes interface implementation
func ToBytes(b []byte) Bytes {
	return bytesType(b)
}

// Copy returns a new copy of slice b, does NOT maintain nil values
func Copy(b []byte) []byte {
	p := make([]byte, len(b))
	copy(p, b)
	return p
}

// BytesToString returns byte slice cast to string via the "unsafe" package
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToBytes returns the string cast to string via the "unsafe" and "reflect" packages
func StringToBytes(s string) []byte {
	// thank you to https://github.com/valyala/fasthttp/blob/master/bytesconv.go
	var b []byte

	// Get byte + string headers
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))

	// Manually set bytes to string
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len

	return b
}

// // InsertByte inserts the supplied byte into the slice at provided position
// func InsertByte(b []byte, at int, c byte) []byte {
// 	return append(append(b[:at], c), b[at:]...)
// }

// // Insert inserts the supplied byte slice into the slice at provided position
// func Insert(b []byte, at int, s []byte) []byte {
// 	return append(append(b[:at], s...), b[at:]...)
// }

// ToUpper offers a faster ToUpper implementation using a lookup table
func ToUpper(b []byte) {
	for i := 0; i < len(b); i++ {
		c := &b[i]
		*c = toUpperTable[*c]
	}
}

// ToLower offers a faster ToLower implementation using a lookup table
func ToLower(b []byte) {
	for i := 0; i < len(b); i++ {
		c := &b[i]
		*c = toLowerTable[*c]
	}
}

// HasBytePrefix returns whether b has the provided byte prefix
func HasBytePrefix(b []byte, c byte) bool {
	return (len(b) > 0) && (b[0] == c)
}

// HasByteSuffix returns whether b has the provided byte suffix
func HasByteSuffix(b []byte, c byte) bool {
	return (len(b) > 0) && (b[len(b)-1] == c)
}

// HasBytePrefix returns b without the provided leading byte
func TrimBytePrefix(b []byte, c byte) []byte {
	if HasBytePrefix(b, c) {
		return b[1:]
	}
	return b
}

// TrimByteSuffix returns b without the provided trailing byte
func TrimByteSuffix(b []byte, c byte) []byte {
	if HasByteSuffix(b, c) {
		return b[:len(b)-1]
	}
	return b
}

// Compare is a direct call-through to standard library bytes.Compare()
func Compare(b, s []byte) int {
	return bytes.Compare(b, s)
}

// Contains is a direct call-through to standard library bytes.Contains()
func Contains(b, s []byte) bool {
	return bytes.Contains(b, s)
}

// TrimPrefix is a direct call-through to standard library bytes.TrimPrefix()
func TrimPrefix(b, s []byte) []byte {
	return bytes.TrimPrefix(b, s)
}

// TrimSuffix is a direct call-through to standard library bytes.TrimSuffix()
func TrimSuffix(b, s []byte) []byte {
	return bytes.TrimSuffix(b, s)
}

// Equal is a direct call-through to standard library bytes.Equal()
func Equal(b, s []byte) bool {
	return bytes.Equal(b, s)
}

// EqualFold is a direct call-through to standard library bytes.EqualFold()
func EqualFold(b, s []byte) bool {
	return bytes.EqualFold(b, s)
}

// Fields is a direct call-through to standard library bytes.Fields()
func Fields(b []byte) [][]byte {
	return bytes.Fields(b)
}

// FieldsFunc is a direct call-through to standard library bytes.FieldsFunc()
func FieldsFunc(b []byte, fn func(rune) bool) [][]byte {
	return bytes.FieldsFunc(b, fn)
}

// HasPrefix is a direct call-through to standard library bytes.HasPrefix()
func HasPrefix(b, s []byte) bool {
	return bytes.HasPrefix(b, s)
}

// HasSuffix is a direct call-through to standard library bytes.HasSuffix()
func HasSuffix(b, s []byte) bool {
	return bytes.HasSuffix(b, s)
}

// Index is a direct call-through to standard library bytes.Index()
func Index(b, s []byte) int {
	return bytes.Index(b, s)
}

// IndexByte is a direct call-through to standard library bytes.IndexByte()
func IndexByte(b []byte, c byte) int {
	return bytes.IndexByte(b, c)
}

// IndexAny is a direct call-through to standard library bytes.IndexAny()
func IndexAny(b []byte, s string) int {
	return bytes.IndexAny(b, s)
}

// IndexRune is a direct call-through to standard library bytes.IndexRune()
func IndexRune(b []byte, r rune) int {
	return bytes.IndexRune(b, r)
}

// IndexFunc is a direct call-through to standard library bytes.IndexFunc()
func IndexFunc(b []byte, fn func(rune) bool) int {
	return bytes.IndexFunc(b, fn)
}

// LastIndex is a direct call-through to standard library bytes.LastIndex()
func LastIndex(b, s []byte) int {
	return bytes.LastIndex(b, s)
}

// LastIndexByte is a direct call-through to standard library bytes.LastIndexByte()
func LastIndexByte(b []byte, c byte) int {
	return bytes.LastIndexByte(b, c)
}

// LastIndexAny is a direct call-through to standard library bytes.LastIndexAny()
func LastIndexAny(b []byte, s string) int {
	return bytes.LastIndexAny(b, s)
}

// LastIndexFunc is a direct call-through to standard library bytes.LastIndexFunc()
func LastIndexFunc(b []byte, fn func(rune) bool) int {
	return bytes.LastIndexFunc(b, fn)
}

// Replace is a direct call-through to standard library bytes.Replace()
func Replace(b, s, r []byte, c int) []byte {
	return bytes.Replace(b, s, r, c)
}

// ReplaceAll is a direct call-through to standard library bytes.ReplaceAll()
func ReplaceAll(b, s, r []byte) []byte {
	return bytes.ReplaceAll(b, s, r)
}

// Split is a direct call-through to standard library bytes.Split()
func Split(b, s []byte) [][]byte {
	return bytes.Split(b, s)
}

// SplitAfter is a direct call-through to standard library bytes.SplitAfter()
func SplitAfter(b, s []byte) [][]byte {
	return bytes.SplitAfter(b, s)
}

// SplitN is a direct call-through to standard library bytes.SplitN()
func SplitN(b, s []byte, c int) [][]byte {
	return bytes.SplitN(b, s, c)
}

// SplitAfterN is a direct call-through to standard library bytes.SplitAfterN()
func SplitAfterN(b, s []byte, c int) [][]byte {
	return bytes.SplitAfterN(b, s, c)
}

// NewReader is a direct call-through to standard library bytes.NewReader()
func NewReader(b []byte) *bytes.Reader {
	return bytes.NewReader(b)
}

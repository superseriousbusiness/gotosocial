package mangler

import (
	"encoding/binary"
	"reflect"
	"unsafe"

	"github.com/cespare/xxhash"
	"github.com/cornelk/hashmap"
)

var (
	// manglers is a map of runtime type ptrs => Mangler functions.
	manglers = hashmap.New[uintptr, Mangler]()

	// bin is a short-hand for our chosen byteorder encoding.
	bin = binary.LittleEndian
)

// Mangled is an interface that allows any type to implement a custom
// Mangler function to improve performance when mangling this type.
type Mangled interface {
	Mangle(buf []byte) []byte
}

// Mangler is a function that will take an input interface value of known
// type, and append it in mangled serialized form to the given byte buffer.
// While the value type is an interface, the Mangler functions are accessed
// by the value's runtime type pointer, allowing the input value type to be known.
type Mangler func(buf []byte, value any) []byte

// rMangler is functionally the same as a Mangler function, but it
// takes the value input in reflected form. By specifying these differences
// in mangler function types, it allows us to cut back on new calls to
// `reflect.ValueOf()` and instead pass by existing reflected values.
type rMangler func(buf []byte, value reflect.Value) []byte

// Get will fetch the Mangler function for given runtime type.
func Get(t reflect.Type) (Mangler, bool) {
	if t == nil {
		return nil, false
	}
	uptr := uintptr(iface_value(t))
	return manglers.Get(uptr)
}

// Register will register the given Mangler function for use with vars of given runtime type. This allows
// registering performant manglers for existing types not implementing Mangled (e.g. std library types).
// NOTE: panics if there already exists a Mangler function for given type. Register on init().
func Register(t reflect.Type, m Mangler) {
	if t == nil {
		// Nil interface{} types cannot be searched by, do not accept
		panic("cannot register mangler for nil interface{} type")
	}

	// Get raw runtime type ptr
	uptr := uintptr(iface_value(t))

	// Ensure this is a unique encoder
	if _, ok := manglers.Get(uptr); ok {
		panic("already registered mangler for type: " + t.String())
	}

	// Cache this encoder func
	manglers.Set(uptr, m)
}

// Append will append the mangled form of input value 'a' to buffer 'b'.
// See mangler.String() for more information on mangled output.
func Append(b []byte, a any) []byte {
	// Get reflect type of 'a'
	t := reflect.TypeOf(a)

	// Get raw runtime type ptr
	uptr := uintptr(iface_value(t))

	// Look for a cached mangler
	mng, ok := manglers.Get(uptr)

	if !ok {
		// Load mangler into cache
		mng = loadMangler(a, t)
		manglers.Set(uptr, mng)
	}

	// First write the type ptr (this adds
	// a unique prefix for each runtime type).
	b = mangle_platform_int(b, uptr)

	// Finally, mangle value
	return mng(b, a)
}

// String will return the mangled format of input value 'a'. This
// mangled output will be unique for all default supported input types
// during a single runtime instance. Uniqueness cannot be guaranteed
// between separate runtime instances (whether running concurrently, or
// the same application running at different times).
//
// The exact formatting of the output data should not be relied upon,
// only that it is unique given the above constraints. Generally though,
// the mangled output is the binary formatted text of given input data.
//
// Uniqueness is guaranteed for similar input data of differing types
// (e.g. string("hello world") vs. []byte("hello world")) by prefixing
// mangled output with the input data's runtime type pointer.
//
// Default supported types include:
// - string
// - bool
// - int,int8,int16,int32,int64
// - uint,uint8,uint16,uint32,uint64,uintptr
// - float32,float64
// - complex64,complex128
// - all type aliases of above
// - time.Time{}, *url.URL{}
// - mangler.Mangled{}
// - encoding.BinaryMarshaler{}
// - all pointers to the above
// - all slices / arrays of the above
// - all map keys / values of the above
func String(a any) string {
	b := Append(make([]byte, 0, 32), a)
	return *(*string)(unsafe.Pointer(&b))
}

// Hash returns the xxHash digest of the result of mangler.Append(nil, 'a').
func Hash(a any) uint64 {
	b := make([]byte, 0, 32)
	b = Append(b, a)
	return xxhash.Sum64(b)
}

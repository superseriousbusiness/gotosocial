package mangler

import (
	"fmt"
	"unsafe"

	"codeberg.org/gruf/go-xunsafe"
)

// Mangler is a function that will take an input value of known type,
// and append it in mangled serialized form to the given byte buffer.
type Mangler func(buf []byte, ptr unsafe.Pointer) []byte

// Get will fetch the Mangler function for given runtime type information.
// The required argument is of type xunsafe.TypeIter{} as unsafe pointer
// access requires further contextual information like type nesting.
func Get(t xunsafe.TypeIter) Mangler {
	t.Parent = nil // enforce type prefix
	fn := loadOrStore(t)
	if fn == nil {
		panic(fmt.Sprintf("cannot mangle type: %s", t.Type))
	}
	return fn
}

// GetNoLoad is functionally similar to Get(),
// without caching the resulting Mangler.
func GetNoLoad(t xunsafe.TypeIter) Mangler {
	t.Parent = nil // enforce type prefix
	fn := loadOrGet(t)
	if fn == nil {
		panic(fmt.Sprintf("cannot mangle type: %s", t.Type))
	}
	return fn
}

// Append will append the mangled form of input value 'a' to buffer 'b'.
//
// See mangler.String() for more information on mangled output.
func Append(b []byte, a any) []byte {
	t := xunsafe.TypeIterFrom(a)
	p := xunsafe.UnpackEface(a)
	return Get(t)(b, p)
}

// AppendMulti appends all mangled forms of input value(s) 'a' to buffer 'b'
// separated by colon characters. When all type manglers are currently cached
// for all types in 'a', this will be faster than multiple calls to Append().
//
// See mangler.String() for more information on mangled output.
func AppendMulti(b []byte, a ...any) []byte {
	if p := manglers.load(); p != nil {
		b4 := len(b)
		for _, a := range a {
			t := xunsafe.TypeIterFrom(a)
			m := (*p)[t.TypeInfo]
			if m == nil {
				b = b[:b4]
				goto slow
			}
			b = m(b, xunsafe.UnpackEface(a))
			b = append(b, '.')
		}
		return b
	}
slow:
	for _, a := range a {
		b = Append(b, a)
		b = append(b, '.')
	}
	return b
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
// Default supported types include all concrete (i.e. non-interface{})
// data types, and interfaces implementing Mangleable{}.
func String(a any) string {
	b := Append(make([]byte, 0, 32), a)
	return *(*string)(unsafe.Pointer(&b))
}

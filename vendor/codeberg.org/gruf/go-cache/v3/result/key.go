package result

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-mangler"
)

// structKeys provides convience methods for a list
// of structKey field combinations used for cache keys.
type structKeys []structKey

// get fetches the structKey info for given lookup name (else, panics).
func (sk structKeys) get(name string) *structKey {
	for i := range sk {
		if sk[i].name == name {
			return &sk[i]
		}
	}
	panic("unknown lookup: \"" + name + "\"")
}

// generate will calculate and produce a slice of cache keys the given value
// can be stored under in the, as determined by receiving struct keys.
func (sk structKeys) generate(a any) []cacheKey {
	var keys []cacheKey

	// Get reflected value in order
	// to access the struct fields
	v := reflect.ValueOf(a)

	// Iteratively deref pointer value
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			panic("nil ptr")
		}
		v = v.Elem()
	}

	// Acquire byte buffer
	buf := getBuf()
	defer putBuf(buf)

outer:
	for i := range sk {
		// Reset buffer
		buf.Reset()

		// Append each field value to buffer.
		for _, field := range sk[i].fields {
			fv := v.Field(field.index)
			fi := fv.Interface()

			// Mangle this key part into buffer.
			ok := field.manglePart(buf, fi)

			if !ok {
				// don't generate keys
				// for zero value parts.
				continue outer
			}

			// Append part separator.
			buf.B = append(buf.B, '.')
		}

		// Drop last '.'
		buf.Truncate(1)

		// Append new cached key to slice
		keys = append(keys, cacheKey{
			info: &sk[i],
			key:  string(buf.B), // copy
		})
	}

	return keys
}

type cacheKeys []cacheKey

// drop will drop the cachedKey with lookup name from receiving cacheKeys slice.
func (ck *cacheKeys) drop(name string) {
	_ = *ck // move out of loop
	for i := range *ck {
		if (*ck)[i].info.name == name {
			(*ck) = append((*ck)[:i], (*ck)[i+1:]...)
			break
		}
	}
}

// cacheKey represents an actual cached key.
type cacheKey struct {
	// info is a reference to the structKey this
	// cacheKey is representing. This is a shared
	// reference and as such only the structKey.pkeys
	// lookup map is expecting to be modified.
	info *structKey

	// value is the actual string representing
	// this cache key for hashmap lookups.
	key string
}

// structKey represents a list of struct fields
// encompassing a single cache key, the string name
// of the lookup, the lookup map to primary cache
// keys, and the key's possible zero value string.
type structKey struct {
	// name is the provided cache lookup name for
	// this particular struct key, consisting of
	// period ('.') separated struct field names.
	name string

	// unique determines whether this structKey supports
	// multiple or just the singular unique result.
	unique bool

	// fields is a slice of runtime struct field
	// indices, of the fields encompassed by this key.
	fields []structField

	// pkeys is a lookup of stored struct key values
	// to the primary cache lookup key (int64).
	pkeys map[string][]int64
}

// newStructKey will generate a structKey{} information object for user-given lookup
// key information, and the receiving generic paramter's type information. Panics on error.
func newStructKey(lk Lookup, t reflect.Type) structKey {
	var sk structKey

	// Set the lookup name
	sk.name = lk.Name

	// Split dot-separated lookup to get
	// the individual struct field names
	names := strings.Split(lk.Name, ".")

	// Allocate the mangler and field indices slice.
	sk.fields = make([]structField, len(names))

	for i, name := range names {
		// Get field info for given name
		ft, ok := t.FieldByName(name)
		if !ok {
			panic("no field found for name: \"" + name + "\"")
		}

		// Check field is usable
		if !isExported(name) {
			panic("field must be exported")
		}

		// Set the runtime field index
		sk.fields[i].index = ft.Index[0]

		// Allocate new instance of field
		v := reflect.New(ft.Type)
		v = v.Elem()

		// Fetch mangler for field type.
		sk.fields[i].mangle = mangler.Get(ft.Type)

		if !lk.AllowZero {
			// Append the mangled zero value interface
			zero := sk.fields[i].mangle(nil, v.Interface())
			sk.fields[i].zero = string(zero)
		}
	}

	// Set unique lookup flag.
	sk.unique = !lk.Multi

	// Allocate primary lookup map
	sk.pkeys = make(map[string][]int64)

	return sk
}

// genKey generates a cache key string for given key parts (i.e. serializes them using "go-mangler").
func (sk *structKey) genKey(parts []any) string {
	// Check this expected no. key parts.
	if len(parts) != len(sk.fields) {
		panic(fmt.Sprintf("incorrect no. key parts provided: want=%d received=%d", len(parts), len(sk.fields)))
	}

	// Acquire byte buffer
	buf := getBuf()
	defer putBuf(buf)
	buf.Reset()

	for i, part := range parts {
		// Mangle this key part into buffer.
		// specifically ignoring whether this
		// is returning a zero value key part.
		_ = sk.fields[i].manglePart(buf, part)

		// Append part separator.
		buf.B = append(buf.B, '.')
	}

	// Drop last '.'
	buf.Truncate(1)

	// Return string copy
	return string(buf.B)
}

type structField struct {
	// index is the reflect index of this struct field.
	index int

	// zero is the possible zero value for this
	// key part. if set, this will _always_ be
	// non-empty due to how the mangler works.
	//
	// i.e. zero = ""  --> allow zero value keys
	//      zero != "" --> don't allow zero value keys
	zero string

	// mangle is the mangler function for
	// serializing values of this struct field.
	mangle mangler.Mangler
}

// manglePart ...
func (field *structField) manglePart(buf *byteutil.Buffer, part any) bool {
	// Start of part bytes.
	start := len(buf.B)

	// Mangle this key part into buffer.
	buf.B = field.mangle(buf.B, part)

	// End of part bytes.
	end := len(buf.B)

	// Return whether this is zero value.
	return (field.zero == "" ||
		string(buf.B[start:end]) != field.zero)
}

// isExported checks whether function name is exported.
func isExported(fnName string) bool {
	r, _ := utf8.DecodeRuneInString(fnName)
	return unicode.IsUpper(r)
}

// bufpool provides a memory pool of byte
// buffers use when encoding key types.
var bufPool = sync.Pool{
	New: func() any {
		return &byteutil.Buffer{B: make([]byte, 0, 512)}
	},
}

// getBuf acquires a byte buffer from memory pool.
func getBuf() *byteutil.Buffer {
	return bufPool.Get().(*byteutil.Buffer)
}

// putBuf replaces a byte buffer back in memory pool.
func putBuf(buf *byteutil.Buffer) {
	if buf.Cap() > int(^uint16(0)) {
		return // drop large bufs
	}
	bufPool.Put(buf)
}

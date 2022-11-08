package result

import (
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-mangler"
)

// structKeys provides convience methods for a list
// of struct field combinations used for cache keys.
type structKeys []keyFields

// get fetches the key-fields for given prefix (else, panics).
func (sk structKeys) get(prefix string) *keyFields {
	for i := range sk {
		if sk[i].prefix == prefix {
			return &sk[i]
		}
	}
	panic("unknown lookup (key prefix): \"" + prefix + "\"")
}

// generate will calculate the value string for each required
// cache key as laid-out by the receiving structKeys{}.
func (sk structKeys) generate(a any) []cacheKey {
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

	// Preallocate expected slice of keys
	keys := make([]cacheKey, len(sk))

	// Acquire byte buffer
	buf := bufpool.Get().(*byteutil.Buffer)
	defer bufpool.Put(buf)

	for i := range sk {
		// Reset buffer
		buf.B = buf.B[:0]

		// Set the key-fields reference
		keys[i].fields = &sk[i]

		// Calculate cache-key value
		keys[i].populate(buf, v)
	}

	return keys
}

// cacheKey represents an actual cache key.
type cacheKey struct {
	// value is the actual string representing
	// this cache key for hashmap lookups.
	value string

	// fieldsRO is a read-only slice (i.e. we should
	// NOT be modifying them, only using for reference)
	// of struct fields encapsulated by this cache key.
	fields *keyFields
}

// populate will calculate the cache key's value string for given
// value's reflected information. Passed encoder is for string building.
func (k *cacheKey) populate(buf *byteutil.Buffer, v reflect.Value) {
	// Append precalculated prefix
	buf.B = append(buf.B, k.fields.prefix...)
	buf.B = append(buf.B, '.')

	// Append each field value to buffer.
	for _, idx := range k.fields.fields {
		fv := v.Field(idx)
		fi := fv.Interface()
		buf.B = mangler.Append(buf.B, fi)
		buf.B = append(buf.B, '.')
	}

	// Drop last '.'
	buf.Truncate(1)

	// Create string copy from buf
	k.value = string(buf.B)
}

// keyFields represents a list of struct fields
// encompassed in a single cache key, including
// the string used as they key's prefix.
type keyFields struct {
	// prefix is the calculated (well, provided)
	// cache key prefix, consisting of dot sep'd
	// struct field names.
	prefix string

	// fields is a slice of runtime struct field
	// indices, of the fields encompassed by this key.
	fields []int
}

// populate will populate this keyFields{} object's .fields member by determining
// the field names from current prefix, and querying given reflected type to get
// the runtime field indices for each of the fields. this speeds-up future value lookups.
func (kf *keyFields) populate(t reflect.Type) {
	// Split dot-separated prefix to get
	// the individual struct field names
	names := strings.Split(kf.prefix, ".")
	if len(names) < 1 {
		panic("no key fields specified")
	}

	// Pre-allocate slice of expected length
	kf.fields = make([]int, len(names))

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
		kf.fields[i] = ft.Index[0]
	}
}

// genkey generates a cache key for given lookup and key value.
func genkey(lookup string, parts ...any) string {
	if len(parts) < 1 {
		// Panic to prevent annoying usecase
		// where user forgets to pass lookup
		// and instead only passes a key part,
		// e.g. cache.Get("key")
		// which then always returns false.
		panic("no key parts provided")
	}

	// Acquire buffer and reset
	buf := bufpool.Get().(*byteutil.Buffer)
	defer bufpool.Put(buf)
	buf.Reset()

	// Append the lookup prefix
	buf.B = append(buf.B, lookup...)
	buf.B = append(buf.B, '.')

	// Encode each key part
	for _, part := range parts {
		buf.B = mangler.Append(buf.B, part)
		buf.B = append(buf.B, '.')
	}

	// Drop last '.'
	buf.Truncate(1)

	// Return string copy
	return string(buf.B)
}

// isExported checks whether function name is exported.
func isExported(fnName string) bool {
	r, _ := utf8.DecodeRuneInString(fnName)
	return unicode.IsUpper(r)
}

// bufpool provides a memory pool of byte
// buffers use when encoding key types.
var bufpool = sync.Pool{
	New: func() any {
		return &byteutil.Buffer{B: make([]byte, 0, 512)}
	},
}

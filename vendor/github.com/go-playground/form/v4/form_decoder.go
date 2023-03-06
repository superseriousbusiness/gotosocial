package form

import (
	"bytes"
	"net/url"
	"reflect"
	"strings"
	"sync"
)

// DecodeCustomTypeFunc allows for registering/overriding types to be parsed.
type DecodeCustomTypeFunc func([]string) (interface{}, error)

// DecodeErrors is a map of errors encountered during form decoding
type DecodeErrors map[string]error

func (d DecodeErrors) Error() string {
	buff := bytes.NewBufferString(blank)

	for k, err := range d {
		buff.WriteString(fieldNS)
		buff.WriteString(k)
		buff.WriteString(errorText)
		buff.WriteString(err.Error())
		buff.WriteString("\n")
	}

	return strings.TrimSpace(buff.String())
}

// An InvalidDecoderError describes an invalid argument passed to Decode.
// (The argument passed to Decode must be a non-nil pointer.)
type InvalidDecoderError struct {
	Type reflect.Type
}

func (e *InvalidDecoderError) Error() string {

	if e.Type == nil {
		return "form: Decode(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "form: Decode(non-pointer " + e.Type.String() + ")"
	}

	return "form: Decode(nil " + e.Type.String() + ")"
}

type key struct {
	ivalue      int
	value       string
	searchValue string
}

type recursiveData struct {
	alias    string
	sliceLen int
	keys     []key
}

type dataMap []*recursiveData

// Decoder is the main decode instance
type Decoder struct {
	tagName         string
	mode            Mode
	structCache     *structCacheMap
	customTypeFuncs map[reflect.Type]DecodeCustomTypeFunc
	maxArraySize    int
	dataPool        *sync.Pool
	namespacePrefix string
	namespaceSuffix string
}

// NewDecoder creates a new decoder instance with sane defaults
func NewDecoder() *Decoder {

	d := &Decoder{
		tagName:         "form",
		mode:            ModeImplicit,
		structCache:     newStructCacheMap(),
		maxArraySize:    10000,
		namespacePrefix: ".",
	}

	d.dataPool = &sync.Pool{New: func() interface{} {
		return &decoder{
			d:         d,
			namespace: make([]byte, 0, 64),
		}
	}}

	return d
}

// SetTagName sets the given tag name to be used by the decoder.
// Default is "form"
func (d *Decoder) SetTagName(tagName string) {
	d.tagName = tagName
}

// SetMode sets the mode the decoder should run
// Default is ModeImplicit
func (d *Decoder) SetMode(mode Mode) {
	d.mode = mode
}

// SetNamespacePrefix sets a struct namespace prefix.
func (d *Decoder) SetNamespacePrefix(namespacePrefix string) {
	d.namespacePrefix = namespacePrefix
}

// SetNamespaceSuffix sets a struct namespace suffix.
func (d *Decoder) SetNamespaceSuffix(namespaceSuffix string) {
	d.namespaceSuffix = namespaceSuffix
}

// SetMaxArraySize sets maximum array size that can be created.
// This limit is for the array indexing this library supports to
// avoid potential DOS or man-in-the-middle attacks using an unusually
// high number.
// DEFAULT: 10000
func (d *Decoder) SetMaxArraySize(size uint) {
	d.maxArraySize = int(size)
}

// RegisterTagNameFunc registers a custom tag name parser function
// NOTE: This method is not thread-safe it is intended that these all be registered prior to any parsing
//
// ADDITIONAL: once a custom function has been registered the default, or custom set, tag name is ignored
// and relies 100% on the function for the name data. The return value WILL BE CACHED and so return value
// must be consistent.
func (d *Decoder) RegisterTagNameFunc(fn TagNameFunc) {
	d.structCache.tagFn = fn
}

// RegisterCustomTypeFunc registers a CustomTypeFunc against a number of types.
// NOTE: This method is not thread-safe it is intended that these all be registered prior to any parsing
//
// ADDITIONAL: if a struct type is registered, the function will only be called if a url.Value exists for
// the struct and not just the struct fields eg. url.Values{"User":"Name%3Djoeybloggs"} will call the
// custom type function with `User` as the type, however url.Values{"User.Name":"joeybloggs"} will not.
func (d *Decoder) RegisterCustomTypeFunc(fn DecodeCustomTypeFunc, types ...interface{}) {

	if d.customTypeFuncs == nil {
		d.customTypeFuncs = map[reflect.Type]DecodeCustomTypeFunc{}
	}

	for _, t := range types {
		d.customTypeFuncs[reflect.TypeOf(t)] = fn
	}
}

// Decode parses the given values and sets the corresponding struct and/or type values
//
// Decode returns an InvalidDecoderError if interface passed is invalid.
func (d *Decoder) Decode(v interface{}, values url.Values) (err error) {

	val := reflect.ValueOf(v)

	if val.Kind() != reflect.Ptr || val.IsNil() {
		return &InvalidDecoderError{reflect.TypeOf(v)}
	}

	dec := d.dataPool.Get().(*decoder)
	dec.values = values
	dec.dm = dec.dm[0:0]

	val = val.Elem()
	typ := val.Type()

	if val.Kind() == reflect.Struct && typ != timeType {
		dec.traverseStruct(val, typ, dec.namespace[0:0])
	} else {
		dec.setFieldByType(val, dec.namespace[0:0], 0)
	}

	if len(dec.errs) > 0 {
		err = dec.errs
		dec.errs = nil
	}

	d.dataPool.Put(dec)

	return
}

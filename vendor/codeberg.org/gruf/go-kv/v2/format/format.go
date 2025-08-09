package format

import (
	"reflect"
	"runtime/debug"
	"strconv"
	"sync"
	"unsafe"
)

// Global formatter instance.
var Global Formatter

// FormatFunc defines a function capable of formatting
// the value contained in State{}.P, based on args in
// State{}.A, storing the result in buffer State{}.B.
type FormatFunc func(*State)

// State contains all necessary
// arguments, buffer and value
// data pointer required for a
// FormatFunc operation, in a
// reusable structure if wanted.
type State struct {

	// A contains args
	// passed to this
	// FormatFunc call.
	A Args

	// B is the buffer
	// that values will
	// be formatted into.
	B []byte

	// P contains a ptr
	// to the value type
	// being formatted.
	P unsafe.Pointer

	// stores pointers to the
	// recent interface values
	// we have visited. to prevent
	// possible recursion of
	// runtime defined data.
	ifaces ptr_ring
}

// ptr_ring size.
const ringsz = 16

// ptr_ring is a ring buffer of pointers,
// purposely stored as uintptrs as all we
// need them for is integer comparisons and
// we don't want to hold-up the GC.
type ptr_ring struct {
	p [ringsz]uintptr
	n uint8
}

func (p *ptr_ring) contains(ptr unsafe.Pointer) bool {
	for _, eptr := range p.p {
		if uintptr(ptr) == eptr {
			return true
		}
	}
	return false
}

func (p *ptr_ring) set(ptr unsafe.Pointer) {
	p.p[p.n%ringsz] = uintptr(ptr)
	p.n++
}

func (p *ptr_ring) clear() {
	p.p = [ringsz]uintptr{}
	p.n = 0
}

// Formatter provides access to value formatting
// provided by this library. It encompasses a set
// of configurable default arguments for when none
// are set, and an internal concurrency-safe cache
// of FormatFuncs to passed value type.
type Formatter struct {

	// Defaults defines the default
	// set of arguments to use when
	// none are supplied to calls to
	// Append() and AppendState().
	Defaults Args

	// internal
	// format func
	// cache map.
	fns sync.Map
}

// LoadFor returns a FormatFunc for the given value type.
func (fmt *Formatter) LoadFor(value any) FormatFunc {
	rtype := reflect.TypeOf(value)
	flags := reflect_iface_elem_flags(rtype)
	t := new_typenode(rtype, flags)
	return fmt.loadOrStore(t)
}

// Append calls AppendState() with a newly allocated State{}, returning byte buffer.
func (fmt *Formatter) Append(buf []byte, value any, args Args) []byte {
	s := new(State)
	s.A = args
	s.B = buf
	fmt.AppendState(s, value)
	return s.B
}

// AppendState will format the given value into the given
// State{}'s byte buffer, using currently-set arguments.
func (fmt *Formatter) AppendState(s *State, value any) {
	switch {
	case s.A != zeroArgs:
		break
	case fmt.Defaults != zeroArgs:
		// use fmt defaults.
		s.A = fmt.Defaults
	default:
		// global defaults.
		s.A = defaultArgs
	}
	s.P = unpack_eface(value)
	s.ifaces.clear()
	s.ifaces.set(s.P)
	fmt.LoadFor(value)(s)
}

func (fmt *Formatter) loadOrGet(t typenode) FormatFunc {
	// Look for existing stored
	// func under this type key.
	v, _ := fmt.fns.Load(t.key())
	fn, _ := v.(FormatFunc)

	if fn == nil {
		// Load format func
		// for typecontext.
		fn = fmt.get(t)
		if fn == nil {
			panic("unreachable")
		}
	}

	return fn
}

func (fmt *Formatter) loadOrStore(t typenode) FormatFunc {
	// Get cache key.
	key := t.key()

	// Look for existing stored
	// func under this type key.
	v, _ := fmt.fns.Load(key)
	fn, _ := v.(FormatFunc)

	if fn == nil {
		// Load format func
		// for typecontext.
		fn = fmt.get(t)
		if fn == nil {
			panic("unreachable")
		}

		// Store in map under type.
		fmt.fns.Store(key, fn)
	}

	return fn
}

var (
	// reflectTypeType is the reflected type of the reflect type,
	// used in fmt.get() to prevent iter of internal ABI structs.
	reflectTypeType = reflect.TypeOf(reflect.TypeOf(0))

	// stringable int types.
	byteType = typeof[byte]()
	runeType = typeof[rune]()

	// stringable slice types.
	bytesType = typeof[[]byte]()
	runesType = typeof[[]rune]()
)

func (fmt *Formatter) get(t typenode) (fn FormatFunc) {
	if t.rtype == nil {
		// catch nil type.
		return appendNil
	}

	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			panic(r) // keep panicking
		} else if fn == nil {
			panic("nil func")
		}

		// Don't allow method functions for map keys,
		// to prevent situation of the method receiver
		// attempting to modify stored map key itself.
		if t.flags&flagKeyType != 0 {
			return
		}

		// Check if type supports known method receiver.
		if methodFn := getMethodType(t); methodFn != nil {

			// Keep ptr to existing
			// non-method format fn.
			noMethodFn := fn

			// Wrap 'fn' to switch
			// between method / none.
			fn = func(s *State) {
				if s.A.NoMethod() {
					noMethodFn(s)
				} else {
					methodFn(s)
				}
			}
		}
	}()

	if t.rtype == reflectTypeType {
		// DO NOT iterate down internal ABI
		// types, some are in non-GC memory.
		return getPointerType(t)
	}

	if !t.visit() {
		// On type recursion simply
		// format as raw pointer.
		return getPointerType(t)
	}

	// Get func for type kind.
	switch t.rtype.Kind() {
	case reflect.Interface:
		return fmt.getInterfaceType(t)
	case reflect.String:
		return getStringType(t)
	case reflect.Bool:
		return getBoolType(t)
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		return getIntType(t)
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		return getUintType(t)
	case reflect.Float32,
		reflect.Float64:
		return getFloatType(t)
	case reflect.Complex64,
		reflect.Complex128:
		return getComplexType(t)
	case reflect.Pointer:
		return fmt.derefPointerType(t)
	case reflect.Array:
		elem := t.rtype.Elem()
		switch fn := fmt.iterArrayType(t); {
		case elem.AssignableTo(byteType):
			return wrapByteArray(t, fn)
		case elem.AssignableTo(runeType):
			return wrapRuneArray(t, fn)
		default:
			return fn
		}
	case reflect.Slice:
		switch fn := fmt.iterSliceType(t); {
		case t.rtype.AssignableTo(bytesType):
			return wrapByteSlice(t, fn)
		case t.rtype.AssignableTo(runesType):
			return wrapRuneSlice(t, fn)
		default:
			return fn
		}
	case reflect.Struct:
		return fmt.iterStructType(t)
	case reflect.Map:
		return fmt.iterMapType(t)
	default:
		return getPointerType(t)
	}
}

func (fmt *Formatter) getInterfaceType(t typenode) FormatFunc {
	if t.rtype.NumMethod() == 0 {
		return func(s *State) {
			// Unpack empty interface.
			eface := *(*any)(s.P)
			s.P = unpack_eface(eface)

			// Get reflected type information.
			rtype := reflect.TypeOf(eface)
			if rtype == nil {
				appendNil(s)
				return
			}

			// Check for ptr recursion.
			if s.ifaces.contains(s.P) {
				getPointerType(t)(s)
				return
			}

			// Store value ptr.
			s.ifaces.set(s.P)

			// Wrap in our typenode for before load.
			flags := reflect_iface_elem_flags(rtype)
			t := new_typenode(rtype, flags)

			// Load + pass to func.
			fmt.loadOrStore(t)(s)
		}
	} else {
		return func(s *State) {
			// Unpack interface-with-method ptr.
			iface := *(*interface{ M() })(s.P)
			s.P = unpack_eface(iface)

			// Get reflected type information.
			rtype := reflect.TypeOf(iface)
			if rtype == nil {
				appendNil(s)
				return
			}

			// Check for ptr recursion.
			if s.ifaces.contains(s.P) {
				getPointerType(t)(s)
				return
			}

			// Store value ptr.
			s.ifaces.set(s.P)

			// Wrap in our typenode for before load.
			flags := reflect_iface_elem_flags(rtype)
			t := new_typenode(rtype, flags)

			// Load + pass to func.
			fmt.loadOrStore(t)(s)
		}
	}
}

func getStringType(t typenode) FormatFunc {
	return with_typestr_ptrs(t, func(s *State) {
		appendString(s, *(*string)(s.P))
	})
}

func getBoolType(t typenode) FormatFunc {
	return with_typestr_ptrs(t, func(s *State) {
		s.B = strconv.AppendBool(s.B, *(*bool)(s.P))
	})
}

func getIntType(t typenode) FormatFunc {
	switch t.rtype.Bits() {
	case 8:
		return with_typestr_ptrs(t, func(s *State) {
			appendInt(s, int64(*(*int8)(s.P)))
		})
	case 16:
		return with_typestr_ptrs(t, func(s *State) {
			appendInt(s, int64(*(*int16)(s.P)))
		})
	case 32:
		return with_typestr_ptrs(t, func(s *State) {
			switch {
			case s.A.AsNumber():
				// fallthrough
			case s.A.AsQuotedASCII():
				s.B = strconv.AppendQuoteRuneToASCII(s.B, *(*rune)(s.P))
				return
			case s.A.AsText() || s.A.AsQuotedText():
				s.B = strconv.AppendQuoteRune(s.B, *(*rune)(s.P))
				return
			}
			appendInt(s, int64(*(*int32)(s.P)))
		})
	case 64:
		return with_typestr_ptrs(t, func(s *State) {
			appendInt(s, int64(*(*int64)(s.P)))
		})
	default:
		panic("unreachable")
	}
}

func getUintType(t typenode) FormatFunc {
	switch t.rtype.Bits() {
	case 8:
		return with_typestr_ptrs(t, func(s *State) {
			switch {
			case s.A.AsNumber():
				// fallthrough
			case s.A.AsText() || s.A.AsQuotedText() || s.A.AsQuotedASCII():
				s.B = AppendQuoteByte(s.B, *(*byte)(s.P))
				return
			}
			appendUint(s, uint64(*(*uint8)(s.P)))
		})
	case 16:
		return with_typestr_ptrs(t, func(s *State) {
			appendUint(s, uint64(*(*uint16)(s.P)))
		})
	case 32:
		return with_typestr_ptrs(t, func(s *State) {
			appendUint(s, uint64(*(*uint32)(s.P)))
		})
	case 64:
		return with_typestr_ptrs(t, func(s *State) {
			appendUint(s, uint64(*(*uint64)(s.P)))
		})
	default:
		panic("unreachable")
	}
}

func getFloatType(t typenode) FormatFunc {
	switch t.rtype.Bits() {
	case 32:
		return with_typestr_ptrs(t, func(s *State) {
			appendFloat(s, float64(*(*float32)(s.P)), 32)
		})
	case 64:
		return with_typestr_ptrs(t, func(s *State) {
			appendFloat(s, float64(*(*float64)(s.P)), 64)
		})
	default:
		panic("unreachable")
	}
}

func getComplexType(t typenode) FormatFunc {
	switch t.rtype.Bits() {
	case 64:
		return with_typestr_ptrs(t, func(s *State) {
			v := *(*complex64)(s.P)
			r, i := real(v), imag(v)
			appendComplex(s, float64(r), float64(i), 32)
		})
	case 128:
		return with_typestr_ptrs(t, func(s *State) {
			v := *(*complex128)(s.P)
			r, i := real(v), imag(v)
			appendComplex(s, float64(r), float64(i), 64)
		})
	default:
		panic("unreachable")
	}
}

func getPointerType(t typenode) FormatFunc {
	switch t.indirect() {
	case true:
		return with_typestr_ptrs(t, func(s *State) {
			s.P = *(*unsafe.Pointer)(s.P)
			appendPointer(s, s.P)
		})
	case false:
		return with_typestr_ptrs(t, func(s *State) {
			appendPointer(s, s.P)
		})
	default:
		panic("unreachable")
	}
}

func with_typestr_ptrs(t typenode, fn FormatFunc) FormatFunc {
	if fn == nil {
		panic("nil func")
	}

	// Check for type wrapping.
	if !t.needs_typestr() {
		return fn
	}

	// Get type string with pointers.
	typestr := t.typestr_with_ptrs()

	// Wrap format func to include
	// type information when needed.
	return func(s *State) {
		if s.A.WithType() {
			s.B = append(s.B, "("+typestr+")("...)
			fn(s)
			s.B = append(s.B, ")"...)
		} else {
			fn(s)
		}
	}
}

func appendString(s *State, v string) {
	switch {
	case s.A.WithType():
		if len(v) > SingleTermLine || !IsSafeASCII(v) {
			// Requires quoting AND escaping
			s.B = strconv.AppendQuote(s.B, v)
		} else if ContainsDoubleQuote(v) {
			// Contains double quotes, needs escaping
			s.B = append(s.B, '"')
			s.B = AppendEscape(s.B, v)
			s.B = append(s.B, '"')
		} else {
			// All else, needs quotes
			s.B = append(s.B, '"')
			s.B = append(s.B, v...)
			s.B = append(s.B, '"')
		}
	case s.A.Logfmt():
		if len(v) > SingleTermLine || !IsSafeASCII(v) {
			// Requires quoting AND escaping
			s.B = strconv.AppendQuote(s.B, v)
		} else if ContainsDoubleQuote(v) {
			// Contains double quotes, needs escaping
			s.B = append(s.B, '"')
			s.B = AppendEscape(s.B, v)
			s.B = append(s.B, '"')
		} else if len(v) == 0 || ContainsSpaceOrTab(v) {
			// Contains space / empty, needs quotes
			s.B = append(s.B, '"')
			s.B = append(s.B, v...)
			s.B = append(s.B, '"')
		} else {
			// All else write as-is
			s.B = append(s.B, v...)
		}
	case s.A.AsQuotedText():
		s.B = strconv.AppendQuote(s.B, v)
	case s.A.AsQuotedASCII():
		s.B = strconv.AppendQuoteToASCII(s.B, v)
	default:
		s.B = append(s.B, v...)
	}
}

func appendInt(s *State, v int64) {
	args := s.A.Int

	// Set argument defaults.
	if args == zeroArgs.Int {
		args = defaultArgs.Int
	}

	// Add any padding.
	if args.Pad > 0 {
		const zeros = `00000000000000000000`
		if args.Pad > len(zeros) {
			panic("cannot pad > " + zeros)
		}

		if v == 0 {
			s.B = append(s.B, zeros[:args.Pad]...)
			return
		}

		// Get absolute.
		abs := abs64(v)

		// Get number of required chars.
		chars := int(v / int64(args.Base))
		if v%int64(args.Base) != 0 {
			chars++
		}

		if abs != v {
			// If this is a negative value,
			// prepend minus ourselves and
			// set value as the absolute.
			s.B = append(s.B, '-')
			v = abs
		}

		// Prepend required zeros.
		n := args.Pad - chars
		s.B = append(s.B, zeros[:n]...)
	}

	// Append value as signed integer w/ args.
	s.B = strconv.AppendInt(s.B, v, args.Base)
}

func appendUint(s *State, v uint64) {
	args := s.A.Int

	// Set argument defaults.
	if args == zeroArgs.Int {
		args = defaultArgs.Int
	}

	// Add any padding.
	if args.Pad > 0 {
		const zeros = `00000000000000000000`
		if args.Pad > len(zeros) {
			panic("cannot pad > " + zeros)
		}

		if v == 0 {
			s.B = append(s.B, zeros[:args.Pad]...)
			return
		}

		// Get number of required chars.
		chars := int(v / uint64(args.Base))
		if v%uint64(args.Base) != 0 {
			chars++
		}

		// Prepend required zeros.
		n := args.Pad - chars
		s.B = append(s.B, zeros[:n]...)
	}

	// Append value as unsigned integer w/ args.
	s.B = strconv.AppendUint(s.B, v, args.Base)
}

func appendFloat(s *State, v float64, bits int) {
	args := s.A.Float

	// Set argument defaults.
	if args == zeroArgs.Float {
		args = defaultArgs.Float
	}

	// Append value as float${bit} w/ args.
	s.B = strconv.AppendFloat(s.B, float64(v),
		args.Fmt, args.Prec, bits)
}

func appendComplex(s *State, r, i float64, bits int) {
	args := s.A.Complex

	// Set argument defaults.
	if args == zeroArgs.Complex {
		args = defaultArgs.Complex
	}

	// Append real value as float${bit} w/ args.
	s.B = strconv.AppendFloat(s.B, float64(r),
		args.Real.Fmt, args.Real.Prec, bits)
	s.B = append(s.B, '+')

	// Append imag value as float${bit} w/ args.
	s.B = strconv.AppendFloat(s.B, float64(i),
		args.Imag.Fmt, args.Imag.Prec, bits)
	s.B = append(s.B, 'i')
}

func appendPointer(s *State, v unsafe.Pointer) {
	if v != nil {
		s.B = append(s.B, "0x"...)
		s.B = strconv.AppendUint(s.B, uint64(uintptr(v)), 16)
	} else {
		appendNil(s)
	}
}

func appendNilType(s *State, typestr string) {
	if s.A.WithType() {
		s.B = append(s.B, "("+typestr+")(<nil>)"...)
	} else {
		s.B = append(s.B, "<nil>"...)
	}
}

func appendNil(s *State) {
	s.B = append(s.B, "<nil>"...)
}

func abs64(i int64) int64 {
	u := uint64(i >> 63)
	return (i ^ int64(u)) + int64(u&1)
}

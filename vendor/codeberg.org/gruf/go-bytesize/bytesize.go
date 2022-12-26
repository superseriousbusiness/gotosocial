package bytesize

import (
	"errors"
	"math/bits"
	_ "strconv"
	"unsafe"
)

const (
	// SI units
	KB Size = 1e3
	MB Size = 1e6
	GB Size = 1e9
	TB Size = 1e12
	PB Size = 1e15
	EB Size = 1e18

	// IEC units
	KiB Size = 1024
	MiB Size = KiB * 1024
	GiB Size = MiB * 1024
	TiB Size = GiB * 1024
	PiB Size = TiB * 1024
	EiB Size = PiB * 1024
)

var (
	// ErrInvalidUnit is returned when an invalid IEC/SI is provided.
	ErrInvalidUnit = errors.New("bytesize: invalid unit")

	// ErrInvalidFormat is returned when an invalid size value is provided.
	ErrInvalidFormat = errors.New("bytesize: invalid format")

	// iecpows is a precomputed table of 1024^n.
	iecpows = [...]float64{
		float64(KiB),
		float64(MiB),
		float64(GiB),
		float64(TiB),
		float64(PiB),
		float64(EiB),
	}

	// sipows is a precomputed table of 1000^n.
	sipows = [...]float64{
		float64(KB),
		float64(MB),
		float64(GB),
		float64(TB),
		float64(PB),
		float64(EB),
	}

	// bvals is a precomputed table of IEC unit values.
	iecvals = [...]float64{
		'k': float64(KiB),
		'K': float64(KiB),
		'M': float64(MiB),
		'G': float64(GiB),
		'T': float64(TiB),
		'P': float64(PiB),
		'E': float64(EiB),
	}

	// sivals is a precomputed table of SI unit values.
	sivals = [...]float64{
		// ASCII numbers _aren't_ valid SI unit values,
		// BUT if the space containing a possible unit
		// char is checked with this table -- it is valid
		// to provide no unit char so unit=1 works.
		'0': 1,
		'1': 1,
		'2': 1,
		'3': 1,
		'4': 1,
		'5': 1,
		'6': 1,
		'7': 1,
		'8': 1,
		'9': 1,

		'k': float64(KB),
		'K': float64(KB),
		'M': float64(MB),
		'G': float64(GB),
		'T': float64(TB),
		'P': float64(PB),
		'E': float64(EB),
	}
)

// Size is a casting for uint64 types that provides formatting
// methods for byte sizes in both IEC and SI units.
type Size uint64

// ParseSize will parse a valid Size from given string. Both IEC and SI units are supported.
func ParseSize(s string) (Size, error) {
	// Parse units from string
	unit, l, err := parseUnit(s)
	if err != nil {
		return 0, err
	}

	// Parse remaining string as float
	f, n, err := atof64(s[:l])
	if err != nil || n != l {
		return 0, ErrInvalidFormat
	}

	return Size(f * unit), nil
}

// Set implements flag.Value{}.
func (sz *Size) Set(in string) error {
	s, err := ParseSize(in)
	if err != nil {
		return err
	}
	*sz = s
	return nil
}

// MarshalText implements encoding.TextMarshaler{}.
func (sz *Size) MarshalText() ([]byte, error) {
	const maxLen = 7 // max IEC string length
	return sz.AppendFormatIEC(make([]byte, 0, maxLen)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler{}.
func (sz *Size) UnmarshalText(text []byte) error {
	return sz.Set(*(*string)(unsafe.Pointer(&text)))
}

// AppendFormat defaults to using Size.AppendFormatIEC().
func (sz Size) AppendFormat(dst []byte) []byte {
	return sz.AppendFormatIEC(dst) // default
}

// AppendFormatSI will append SI formatted size to 'dst'.
func (sz Size) AppendFormatSI(dst []byte) []byte {
	if uint64(sz) < 1000 {
		dst = itoa(dst, uint64(sz))
		dst = append(dst, 'B')
		return dst
	} // above is fast-path, .appendFormat() is outlined
	return sz.appendFormat(dst, 1000, &sipows, "B")
}

// AppendFormatIEC will append IEC formatted size to 'dst'.
func (sz Size) AppendFormatIEC(dst []byte) []byte {
	if uint64(sz) < 1024 {
		dst = itoa(dst, uint64(sz))
		dst = append(dst, 'B')
		return dst
	} // above is fast-path, .appendFormat() is outlined
	return sz.appendFormat(dst, 1024, &iecpows, "iB")
}

// appendFormat will append formatted Size to 'dst', depending on base, powers table and single unit suffix.
func (sz Size) appendFormat(dst []byte, base uint64, pows *[6]float64, sunit string) []byte {
	const (
		// min "small" unit threshold
		min = 0.75

		// binary unit chars.
		units = `kMGTPE`
	)

	// Larger number: get value of
	// i / unit size. We have a 'min'
	// threshold after which we prefer
	// using the unit 1 down
	n := bits.Len64(uint64(sz)) / 10
	f := float64(sz) / pows[n-1]
	if f < min {
		f *= float64(base)
		n--
	}

	// Append formatted float with units
	dst = ftoa(dst, f)
	dst = append(dst, units[n-1])
	dst = append(dst, sunit...)
	return dst
}

// StringSI returns an SI unit string format of Size.
func (sz Size) StringSI() string {
	const maxLen = 6 // max SI string length
	b := sz.AppendFormatSI(make([]byte, 0, maxLen))
	return *(*string)(unsafe.Pointer(&b))
}

// StringIEC returns an IEC unit string format of Size.
func (sz Size) StringIEC() string {
	const maxLen = 7 // max IEC string length
	b := sz.AppendFormatIEC(make([]byte, 0, maxLen))
	return *(*string)(unsafe.Pointer(&b))
}

// String returns a string format of Size, defaults to IEC unit format.
func (sz Size) String() string {
	return sz.StringIEC()
}

// parseUnit will parse the byte size unit from string 's'.
func parseUnit(s string) (float64, int, error) {
	// Check for string
	if len(s) < 1 {
		return 0, 0, ErrInvalidFormat
	}

	// Strip 'byte' unit suffix
	if l := len(s) - 1; s[l] == 'B' {
		s = s[:l]

		if len(s) < 1 {
			// No remaining str before unit suffix
			return 0, 0, ErrInvalidFormat
		}
	}

	// Strip IEC binary unit suffix
	if l := len(s) - 1; s[l] == 'i' {
		s = s[:l]

		if len(s) < 1 {
			// No remaining str before unit suffix
			return 0, 0, ErrInvalidFormat
		}

		// Location of unit char.
		l := len(s) - 1
		c := int(s[l])

		// Check valid unit char was provided
		if len(iecvals) < c || iecvals[c] == 0 {
			return 0, 0, ErrInvalidUnit
		}

		// Return parsed IEC unit size
		return iecvals[c], l, nil
	}

	// Location of unit char.
	l := len(s) - 1
	c := int(s[l])

	switch {
	// Check valid unit char provided
	case len(sivals) < c || sivals[c] == 0:
		return 0, 0, ErrInvalidUnit

	// No unit char (only ascii number)
	case sivals[c] == 1:
		l++
	}

	// Return parsed SI unit size
	return sivals[c], l, nil
}

// ftoa appends string formatted 'f' to 'dst', assumed < ~800.
func ftoa(dst []byte, f float64) []byte {
	switch i := uint64(f); {
	// Append with 2 d.p.
	case i < 10:
		f *= 10

		// Calculate next dec. value
		d1 := uint8(uint64(f) % 10)

		f *= 10

		// Calculate next dec. value
		d2 := uint8(uint64(f) % 10)

		// Round the final value
		if uint64(f*10)%10 > 4 {
			d2++

			// Overflow, incr 'd1'
			if d2 == 10 {
				d2 = 0
				d1++

				// Overflow, incr 'i'
				if d1 == 10 {
					d1 = 0
					i++
				}
			}
		}

		// Append decimal value
		dst = itoa(dst, i)
		dst = append(dst,
			'.',
			'0'+d1,
			'0'+d2,
		)

	// Append with 1 d.p.
	case i < 100:
		f *= 10

		// Calculate next dec. value
		d1 := uint8(uint64(f) % 10)

		// Round the final value
		if uint64(f*10)%10 > 4 {
			d1++

			// Overflow, incr 'i'
			if d1 == 10 {
				d1 = 0
				i++
			}
		}

		// Append decimal value
		dst = itoa(dst, i)
		dst = append(dst, '.', '0'+d1)

	// No decimal places
	default:
		dst = itoa(dst, i)
	}

	return dst
}

// itoa appends string formatted 'i' to 'dst'.
func itoa(dst []byte, i uint64) []byte {
	// Assemble uint in reverse order.
	var b [4]byte
	bp := len(b) - 1

	// Append integer
	for i >= 10 {
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	} // i < 10
	b[bp] = byte('0' + i)

	return append(dst, b[bp:]...)
}

// We use the following internal strconv function usually
// used internally to parse float values, as we know that
// are value passed will always be of 64bit type, and knowing
// the returned float string length is very helpful!
//
//go:linkname atof64 strconv.atof64
func atof64(string) (float64, int, error)

package bytesize

import (
	"errors"
	"math/bits"
	"unsafe"
)

var (
	// ErrInvalidUnit is returned when an invalid IEC/SI is provided.
	ErrInvalidUnit = errors.New("bytesize: invalid unit")

	// ErrInvalidFormat is returned when an invalid size value is provided.
	ErrInvalidFormat = errors.New("bytesize: invalid format")

	// bunits are the binary unit chars.
	units = `kMGTPE`

	// iecpows is a precomputed table of 1024^n.
	iecpows = [...]float64{
		float64(1024),                                    // KiB
		float64(1024 * 1024),                             // MiB
		float64(1024 * 1024 * 1024),                      // GiB
		float64(1024 * 1024 * 1024 * 1024),               // TiB
		float64(1024 * 1024 * 1024 * 1024 * 1024),        // PiB
		float64(1024 * 1024 * 1024 * 1024 * 1024 * 1024), // EiB
	}

	// sipows is a precomputed table of 1000^n.
	sipows = [...]float64{
		float64(1e3),  // KB
		float64(1e6),  // MB
		float64(1e9),  // GB
		float64(1e12), // TB
		float64(1e15), // PB
		float64(1e18), // EB
	}

	// bvals is a precomputed table of IEC unit values.
	iecvals = [...]uint64{
		'k': 1024,
		'K': 1024,
		'M': 1024 * 1024,
		'G': 1024 * 1024 * 1024,
		'T': 1024 * 1024 * 1024 * 1024,
		'P': 1024 * 1024 * 1024 * 1024 * 1024,
		'E': 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
	}

	// sivals is a precomputed table of SI unit values.
	sivals = [...]uint64{
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

		'k': 1e3,
		'M': 1e6,
		'G': 1e9,
		'T': 1e12,
		'P': 1e15,
		'E': 1e18,
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

	return Size(uint64(f) * unit), nil
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
	const min = 0.75

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
	b := sz.AppendFormatSI(make([]byte, 0, 6))
	return *(*string)(unsafe.Pointer(&b))
}

// StringIEC returns an IEC unit string format of Size.
func (sz Size) StringIEC() string {
	b := sz.AppendFormatIEC(make([]byte, 0, 7))
	return *(*string)(unsafe.Pointer(&b))
}

// String returns a string format of Size, defaults to IEC unit format.
func (sz Size) String() string {
	return sz.StringIEC()
}

// parseUnit will parse the byte size unit from string 's'.
func parseUnit(s string) (uint64, int, error) {
	var isIEC bool

	// Check for string
	if len(s) < 1 {
		return 0, 0, ErrInvalidFormat
	}

	// Strip 'byte' unit suffix
	if l := len(s) - 1; s[l] == 'B' {
		s = s[:l]

		// Check str remains
		if len(s) < 1 {
			return 0, 0, ErrInvalidFormat
		}
	}

	// Strip IEC binary unit suffix
	if l := len(s) - 1; s[l] == 'i' {
		s = s[:l]
		isIEC = true

		// Check str remains
		if len(s) < 1 {
			return 0, 0, ErrInvalidFormat
		}
	}

	// Location of unit char.
	l := len(s) - 1

	var unit uint64
	switch c := int(s[l]); {
	// Determine IEC unit in use
	case isIEC && c < len(iecvals):
		unit = iecvals[c]
		if unit == 0 {
			return 0, 0, ErrInvalidUnit
		}

	// Determine SI unit in use
	case c < len(sivals):
		unit = sivals[c]
		switch unit {
		case 0:
			return 0, 0, ErrInvalidUnit
		case 1:
			l++
		}
	}

	return unit, l, nil
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
	// Assemble int in reverse order.
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

//go:linkname atof64 strconv.atof64
func atof64(string) (float64, int, error)

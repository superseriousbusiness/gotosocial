package split

import (
	"strconv"
	"time"
	"unsafe"

	"codeberg.org/gruf/go-bytesize"
)

// Signed defines a signed
// integer generic type parameter.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Unsigned defines an unsigned
// integer generic type paramter.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Float defines a float-type generic parameter.
type Float interface{ ~float32 | ~float64 }

// SplitFunc will split input string on commas, taking into account string quoting
// and stripping extra whitespace, passing each split to the given function hook.
func SplitFunc(str string, fn func(string) error) error {
	return (&Splitter{}).SplitFunc(str, fn)
}

// SplitStrings will pass string input to SplitFunc(), compiling a slice of strings.
func SplitStrings[String ~string](str string) ([]String, error) {
	var slice []String

	// Simply append each split string to slice
	if err := SplitFunc(str, func(s string) error {
		slice = append(slice, String(s))
		return nil
	}); err != nil {
		return nil, err
	}

	return slice, nil
}

// SplitBools will pass string input to SplitFunc(), parsing and compiling a slice of bools.
func SplitBools[Bool ~bool](str string) ([]Bool, error) {
	var slice []Bool

	// Parse each bool split from input string
	if err := SplitFunc(str, func(s string) error {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		slice = append(slice, Bool(b))
		return nil
	}); err != nil {
		return nil, err
	}

	return slice, nil
}

// SplitInts will pass string input to SplitFunc(), parsing and compiling a slice of signed integers.
func SplitInts[Int Signed](str string) ([]Int, error) {
	// Determine bits from param type size
	bits := int(unsafe.Sizeof(Int(0)) * 8)

	var slice []Int

	// Parse each int split from input string
	if err := SplitFunc(str, func(s string) error {
		i, err := strconv.ParseInt(s, 10, bits)
		if err != nil {
			return err
		}
		slice = append(slice, Int(i))
		return nil
	}); err != nil {
		return nil, err
	}

	return slice, nil
}

// SplitUints will pass string input to SplitFunc(), parsing and compiling a slice of unsigned integers.
func SplitUints[Uint Unsigned](str string) ([]Uint, error) {
	// Determine bits from param type size
	bits := int(unsafe.Sizeof(Uint(0)) * 8)

	var slice []Uint

	// Parse each uint split from input string
	if err := SplitFunc(str, func(s string) error {
		u, err := strconv.ParseUint(s, 10, bits)
		if err != nil {
			return err
		}
		slice = append(slice, Uint(u))
		return nil
	}); err != nil {
		return nil, err
	}

	return slice, nil
}

// SplitFloats will pass string input to SplitFunc(), parsing and compiling a slice of floats.
func SplitFloats[Float_ Float](str string) ([]Float_, error) {
	// Determine bits from param type size
	bits := int(unsafe.Sizeof(Float_(0)) * 8)

	var slice []Float_

	// Parse each float split from input string
	if err := SplitFunc(str, func(s string) error {
		f, err := strconv.ParseFloat(s, bits)
		if err != nil {
			return err
		}
		slice = append(slice, Float_(f))
		return nil
	}); err != nil {
		return nil, err
	}

	return slice, nil
}

// SplitSizes will pass string input to SplitFunc(), parsing and compiling a slice of byte sizes.
func SplitSizes(str string) ([]bytesize.Size, error) {
	var slice []bytesize.Size

	// Parse each size split from input string
	if err := SplitFunc(str, func(s string) error {
		sz, err := bytesize.ParseSize(s)
		if err != nil {
			return err
		}
		slice = append(slice, sz)
		return nil
	}); err != nil {
		return nil, err
	}

	return slice, nil
}

// SplitDurations will pass string input to SplitFunc(), parsing and compiling a slice of durations.
func SplitDurations(str string) ([]time.Duration, error) {
	var slice []time.Duration

	// Parse each duration split from input string
	if err := SplitFunc(str, func(s string) error {
		d, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		slice = append(slice, d)
		return nil
	}); err != nil {
		return nil, err
	}

	return slice, nil
}

// SplitTimes will pass string input to SplitFunc(), parsing and compiling a slice of times.
func SplitTimes(str string, format string) ([]time.Time, error) {
	var slice []time.Time

	// Parse each time split from input string
	if err := SplitFunc(str, func(s string) error {
		t, err := time.Parse(s, format)
		if err != nil {
			return err
		}
		slice = append(slice, t)
		return nil
	}); err != nil {
		return nil, err
	}

	return slice, nil
}

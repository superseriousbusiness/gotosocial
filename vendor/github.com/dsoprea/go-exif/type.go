package exif

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

type TagTypePrimitive uint16

func (typeType TagTypePrimitive) String() string {
	return TypeNames[typeType]
}

func (tagType TagTypePrimitive) Size() int {
	if tagType == TypeByte {
		return 1
	} else if tagType == TypeAscii || tagType == TypeAsciiNoNul {
		return 1
	} else if tagType == TypeShort {
		return 2
	} else if tagType == TypeLong {
		return 4
	} else if tagType == TypeRational {
		return 8
	} else if tagType == TypeSignedLong {
		return 4
	} else if tagType == TypeSignedRational {
		return 8
	} else {
		log.Panicf("can not determine tag-value size for type (%d): [%s]", tagType, TypeNames[tagType])

		// Never called.
		return 0
	}
}

const (
	TypeByte           TagTypePrimitive = 1
	TypeAscii          TagTypePrimitive = 2
	TypeShort          TagTypePrimitive = 3
	TypeLong           TagTypePrimitive = 4
	TypeRational       TagTypePrimitive = 5
	TypeUndefined      TagTypePrimitive = 7
	TypeSignedLong     TagTypePrimitive = 9
	TypeSignedRational TagTypePrimitive = 10

	// TypeAsciiNoNul is just a pseudo-type, for our own purposes.
	TypeAsciiNoNul TagTypePrimitive = 0xf0
)

var (
	typeLogger = log.NewLogger("exif.type")
)

var (
	// TODO(dustin): Rename TypeNames() to typeNames() and add getter.
	TypeNames = map[TagTypePrimitive]string{
		TypeByte:           "BYTE",
		TypeAscii:          "ASCII",
		TypeShort:          "SHORT",
		TypeLong:           "LONG",
		TypeRational:       "RATIONAL",
		TypeUndefined:      "UNDEFINED",
		TypeSignedLong:     "SLONG",
		TypeSignedRational: "SRATIONAL",

		TypeAsciiNoNul: "_ASCII_NO_NUL",
	}

	TypeNamesR = map[string]TagTypePrimitive{}
)

var (
	// ErrNotEnoughData is used when there isn't enough data to accomodate what
	// we're trying to parse (sizeof(type) * unit_count).
	ErrNotEnoughData = errors.New("not enough data for type")

	// ErrWrongType is used when we try to parse anything other than the
	// current type.
	ErrWrongType = errors.New("wrong type, can not parse")

	// ErrUnhandledUnknownTag is used when we try to parse a tag that's
	// recorded as an "unknown" type but not a documented tag (therefore
	// leaving us not knowning how to read it).
	ErrUnhandledUnknownTypedTag = errors.New("not a standard unknown-typed tag")
)

type Rational struct {
	Numerator   uint32
	Denominator uint32
}

type SignedRational struct {
	Numerator   int32
	Denominator int32
}

func TagTypeSize(tagType TagTypePrimitive) int {

	// DEPRECATED(dustin): `(TagTypePrimitive).Size()` should be used, directly.

	return tagType.Size()
}

// Format returns a stringified value for the given bytes. Automatically
// calculates count based on type size.
func Format(rawBytes []byte, tagType TagTypePrimitive, justFirst bool, byteOrder binary.ByteOrder) (value string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): !! Add tests

	typeSize := tagType.Size()

	if len(rawBytes)%typeSize != 0 {
		log.Panicf("byte-count (%d) does not align for [%s] type with a size of (%d) bytes", len(rawBytes), TypeNames[tagType], typeSize)
	}

	// unitCount is the calculated unit-count. This should equal the original
	// value from the tag (pre-resolution).
	unitCount := uint32(len(rawBytes) / typeSize)

	// Truncate the items if it's not bytes or a string and we just want the first.

	valueSuffix := ""
	if justFirst == true && unitCount > 1 && tagType != TypeByte && tagType != TypeAscii && tagType != TypeAsciiNoNul {
		unitCount = 1
		valueSuffix = "..."
	}

	if tagType == TypeByte {
		items, err := parser.ParseBytes(rawBytes, unitCount)
		log.PanicIf(err)

		return DumpBytesToString(items), nil
	} else if tagType == TypeAscii {
		phrase, err := parser.ParseAscii(rawBytes, unitCount)
		log.PanicIf(err)

		return phrase, nil
	} else if tagType == TypeAsciiNoNul {
		phrase, err := parser.ParseAsciiNoNul(rawBytes, unitCount)
		log.PanicIf(err)

		return phrase, nil
	} else if tagType == TypeShort {
		items, err := parser.ParseShorts(rawBytes, unitCount, byteOrder)
		log.PanicIf(err)

		if len(items) > 0 {
			if justFirst == true {
				return fmt.Sprintf("%v%s", items[0], valueSuffix), nil
			} else {
				return fmt.Sprintf("%v", items), nil
			}
		} else {
			return "", nil
		}
	} else if tagType == TypeLong {
		items, err := parser.ParseLongs(rawBytes, unitCount, byteOrder)
		log.PanicIf(err)

		if len(items) > 0 {
			if justFirst == true {
				return fmt.Sprintf("%v%s", items[0], valueSuffix), nil
			} else {
				return fmt.Sprintf("%v", items), nil
			}
		} else {
			return "", nil
		}
	} else if tagType == TypeRational {
		items, err := parser.ParseRationals(rawBytes, unitCount, byteOrder)
		log.PanicIf(err)

		if len(items) > 0 {
			parts := make([]string, len(items))
			for i, r := range items {
				parts[i] = fmt.Sprintf("%d/%d", r.Numerator, r.Denominator)
			}

			if justFirst == true {
				return fmt.Sprintf("%v%s", parts[0], valueSuffix), nil
			} else {
				return fmt.Sprintf("%v", parts), nil
			}
		} else {
			return "", nil
		}
	} else if tagType == TypeSignedLong {
		items, err := parser.ParseSignedLongs(rawBytes, unitCount, byteOrder)
		log.PanicIf(err)

		if len(items) > 0 {
			if justFirst == true {
				return fmt.Sprintf("%v%s", items[0], valueSuffix), nil
			} else {
				return fmt.Sprintf("%v", items), nil
			}
		} else {
			return "", nil
		}
	} else if tagType == TypeSignedRational {
		items, err := parser.ParseSignedRationals(rawBytes, unitCount, byteOrder)
		log.PanicIf(err)

		parts := make([]string, len(items))
		for i, r := range items {
			parts[i] = fmt.Sprintf("%d/%d", r.Numerator, r.Denominator)
		}

		if len(items) > 0 {
			if justFirst == true {
				return fmt.Sprintf("%v%s", parts[0], valueSuffix), nil
			} else {
				return fmt.Sprintf("%v", parts), nil
			}
		} else {
			return "", nil
		}
	} else {
		// Affects only "unknown" values, in general.
		log.Panicf("value of type [%s] can not be formatted into string", tagType.String())

		// Never called.
		return "", nil
	}
}

func EncodeStringToBytes(tagType TagTypePrimitive, valueString string) (value interface{}, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if tagType == TypeUndefined {
		// TODO(dustin): Circle back to this.
		log.Panicf("undefined-type values are not supported")
	}

	if tagType == TypeByte {
		return []byte(valueString), nil
	} else if tagType == TypeAscii || tagType == TypeAsciiNoNul {
		// Whether or not we're putting an NUL on the end is only relevant for
		// byte-level encoding. This function really just supports a user
		// interface.

		return valueString, nil
	} else if tagType == TypeShort {
		n, err := strconv.ParseUint(valueString, 10, 16)
		log.PanicIf(err)

		return uint16(n), nil
	} else if tagType == TypeLong {
		n, err := strconv.ParseUint(valueString, 10, 32)
		log.PanicIf(err)

		return uint32(n), nil
	} else if tagType == TypeRational {
		parts := strings.SplitN(valueString, "/", 2)

		numerator, err := strconv.ParseUint(parts[0], 10, 32)
		log.PanicIf(err)

		denominator, err := strconv.ParseUint(parts[1], 10, 32)
		log.PanicIf(err)

		return Rational{
			Numerator:   uint32(numerator),
			Denominator: uint32(denominator),
		}, nil
	} else if tagType == TypeSignedLong {
		n, err := strconv.ParseInt(valueString, 10, 32)
		log.PanicIf(err)

		return int32(n), nil
	} else if tagType == TypeSignedRational {
		parts := strings.SplitN(valueString, "/", 2)

		numerator, err := strconv.ParseInt(parts[0], 10, 32)
		log.PanicIf(err)

		denominator, err := strconv.ParseInt(parts[1], 10, 32)
		log.PanicIf(err)

		return SignedRational{
			Numerator:   int32(numerator),
			Denominator: int32(denominator),
		}, nil
	}

	log.Panicf("from-string encoding for type not supported; this shouldn't happen: [%s]", tagType.String())
	return nil, nil
}

func init() {
	for typeId, typeName := range TypeNames {
		TypeNamesR[typeName] = typeId
	}
}

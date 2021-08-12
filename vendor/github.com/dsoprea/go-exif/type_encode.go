package exif

import (
	"bytes"
	"reflect"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

var (
	typeEncodeLogger = log.NewLogger("exif.type_encode")
)

// EncodedData encapsulates the compound output of an encoding operation.
type EncodedData struct {
	Type    TagTypePrimitive
	Encoded []byte

	// TODO(dustin): Is this really necessary? We might have this just to correlate to the incoming stream format (raw bytes and a unit-count both for incoming and outgoing).
	UnitCount uint32
}

type ValueEncoder struct {
	byteOrder binary.ByteOrder
}

func NewValueEncoder(byteOrder binary.ByteOrder) *ValueEncoder {
	return &ValueEncoder{
		byteOrder: byteOrder,
	}
}

func (ve *ValueEncoder) encodeBytes(value []uint8) (ed EncodedData, err error) {
	ed.Type = TypeByte
	ed.Encoded = []byte(value)
	ed.UnitCount = uint32(len(value))

	return ed, nil
}

func (ve *ValueEncoder) encodeAscii(value string) (ed EncodedData, err error) {
	ed.Type = TypeAscii

	ed.Encoded = []byte(value)
	ed.Encoded = append(ed.Encoded, 0)

	ed.UnitCount = uint32(len(ed.Encoded))

	return ed, nil
}

// encodeAsciiNoNul returns a string encoded as a byte-string without a trailing
// NUL byte.
//
// Note that:
//
// 1. This type can not be automatically encoded using `Encode()`. The default
//    mode is to encode *with* a trailing NUL byte using `encodeAscii`. Only
//    certain undefined-type tags using an unterminated ASCII string and these
//    are exceptional in nature.
//
// 2. The presence of this method allows us to completely test the complimentary
//    no-nul parser.
//
func (ve *ValueEncoder) encodeAsciiNoNul(value string) (ed EncodedData, err error) {
	ed.Type = TypeAsciiNoNul
	ed.Encoded = []byte(value)
	ed.UnitCount = uint32(len(ed.Encoded))

	return ed, nil
}

func (ve *ValueEncoder) encodeShorts(value []uint16) (ed EncodedData, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ed.UnitCount = uint32(len(value))
	ed.Encoded = make([]byte, ed.UnitCount*2)

	for i := uint32(0); i < ed.UnitCount; i++ {
		ve.byteOrder.PutUint16(ed.Encoded[i*2:(i+1)*2], value[i])
	}

	ed.Type = TypeShort

	return ed, nil
}

func (ve *ValueEncoder) encodeLongs(value []uint32) (ed EncodedData, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ed.UnitCount = uint32(len(value))
	ed.Encoded = make([]byte, ed.UnitCount*4)

	for i := uint32(0); i < ed.UnitCount; i++ {
		ve.byteOrder.PutUint32(ed.Encoded[i*4:(i+1)*4], value[i])
	}

	ed.Type = TypeLong

	return ed, nil
}

func (ve *ValueEncoder) encodeRationals(value []Rational) (ed EncodedData, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ed.UnitCount = uint32(len(value))
	ed.Encoded = make([]byte, ed.UnitCount*8)

	for i := uint32(0); i < ed.UnitCount; i++ {
		ve.byteOrder.PutUint32(ed.Encoded[i*8+0:i*8+4], value[i].Numerator)
		ve.byteOrder.PutUint32(ed.Encoded[i*8+4:i*8+8], value[i].Denominator)
	}

	ed.Type = TypeRational

	return ed, nil
}

func (ve *ValueEncoder) encodeSignedLongs(value []int32) (ed EncodedData, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ed.UnitCount = uint32(len(value))

	b := bytes.NewBuffer(make([]byte, 0, 8*ed.UnitCount))

	for i := uint32(0); i < ed.UnitCount; i++ {
		err := binary.Write(b, ve.byteOrder, value[i])
		log.PanicIf(err)
	}

	ed.Type = TypeSignedLong
	ed.Encoded = b.Bytes()

	return ed, nil
}

func (ve *ValueEncoder) encodeSignedRationals(value []SignedRational) (ed EncodedData, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ed.UnitCount = uint32(len(value))

	b := bytes.NewBuffer(make([]byte, 0, 8*ed.UnitCount))

	for i := uint32(0); i < ed.UnitCount; i++ {
		err := binary.Write(b, ve.byteOrder, value[i].Numerator)
		log.PanicIf(err)

		err = binary.Write(b, ve.byteOrder, value[i].Denominator)
		log.PanicIf(err)
	}

	ed.Type = TypeSignedRational
	ed.Encoded = b.Bytes()

	return ed, nil
}

// Encode returns bytes for the given value, infering type from the actual
// value. This does not support `TypeAsciiNoNull` (all strings are encoded as
// `TypeAscii`).
func (ve *ValueEncoder) Encode(value interface{}) (ed EncodedData, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): This is redundant with EncodeWithType. Refactor one to use the other.

	switch value.(type) {
	case []byte:
		ed, err = ve.encodeBytes(value.([]byte))
		log.PanicIf(err)
	case string:
		ed, err = ve.encodeAscii(value.(string))
		log.PanicIf(err)
	case []uint16:
		ed, err = ve.encodeShorts(value.([]uint16))
		log.PanicIf(err)
	case []uint32:
		ed, err = ve.encodeLongs(value.([]uint32))
		log.PanicIf(err)
	case []Rational:
		ed, err = ve.encodeRationals(value.([]Rational))
		log.PanicIf(err)
	case []int32:
		ed, err = ve.encodeSignedLongs(value.([]int32))
		log.PanicIf(err)
	case []SignedRational:
		ed, err = ve.encodeSignedRationals(value.([]SignedRational))
		log.PanicIf(err)
	default:
		log.Panicf("value not encodable: [%s] [%v]", reflect.TypeOf(value), value)
	}

	return ed, nil
}

// EncodeWithType returns bytes for the given value, using the given `TagType`
// value to determine how to encode. This supports `TypeAsciiNoNul`.
func (ve *ValueEncoder) EncodeWithType(tt TagType, value interface{}) (ed EncodedData, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): This is redundant with Encode. Refactor one to use the other.

	switch tt.Type() {
	case TypeByte:
		ed, err = ve.encodeBytes(value.([]byte))
		log.PanicIf(err)
	case TypeAscii:
		ed, err = ve.encodeAscii(value.(string))
		log.PanicIf(err)
	case TypeAsciiNoNul:
		ed, err = ve.encodeAsciiNoNul(value.(string))
		log.PanicIf(err)
	case TypeShort:
		ed, err = ve.encodeShorts(value.([]uint16))
		log.PanicIf(err)
	case TypeLong:
		ed, err = ve.encodeLongs(value.([]uint32))
		log.PanicIf(err)
	case TypeRational:
		ed, err = ve.encodeRationals(value.([]Rational))
		log.PanicIf(err)
	case TypeSignedLong:
		ed, err = ve.encodeSignedLongs(value.([]int32))
		log.PanicIf(err)
	case TypeSignedRational:
		ed, err = ve.encodeSignedRationals(value.([]SignedRational))
		log.PanicIf(err)
	default:
		log.Panicf("value not encodable (with type): %v [%v]", tt, value)
	}

	return ed, nil
}

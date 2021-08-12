package exif

// NOTE(dustin): Most of this file encapsulates deprecated functionality and awaits being dumped in a future release.

import (
	"fmt"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

type TagType struct {
	tagType   TagTypePrimitive
	name      string
	byteOrder binary.ByteOrder
}

func NewTagType(tagType TagTypePrimitive, byteOrder binary.ByteOrder) TagType {
	name, found := TypeNames[tagType]
	if found == false {
		log.Panicf("tag-type not valid: 0x%04x", tagType)
	}

	return TagType{
		tagType:   tagType,
		name:      name,
		byteOrder: byteOrder,
	}
}

func (tt TagType) String() string {
	return fmt.Sprintf("TagType<NAME=[%s]>", tt.name)
}

func (tt TagType) Name() string {
	return tt.name
}

func (tt TagType) Type() TagTypePrimitive {
	return tt.tagType
}

func (tt TagType) ByteOrder() binary.ByteOrder {
	return tt.byteOrder
}

func (tt TagType) Size() int {

	// DEPRECATED(dustin): `(TagTypePrimitive).Size()` should be used, directly.

	return tt.Type().Size()
}

// valueIsEmbedded will return a boolean indicating whether the value should be
// found directly within the IFD entry or an offset to somewhere else.
func (tt TagType) valueIsEmbedded(unitCount uint32) bool {
	return (tt.tagType.Size() * int(unitCount)) <= 4
}

func (tt TagType) readRawEncoded(valueContext ValueContext) (rawBytes []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	unitSizeRaw := uint32(tt.tagType.Size())

	if tt.valueIsEmbedded(valueContext.UnitCount()) == true {
		byteLength := unitSizeRaw * valueContext.UnitCount()
		return valueContext.RawValueOffset()[:byteLength], nil
	} else {
		return valueContext.AddressableData()[valueContext.ValueOffset() : valueContext.ValueOffset()+valueContext.UnitCount()*unitSizeRaw], nil
	}
}

func (tt TagType) ParseBytes(data []byte, unitCount uint32) (value []uint8, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(*Parser).ParseBytes()` should be used.

	value, err = parser.ParseBytes(data, unitCount)
	log.PanicIf(err)

	return value, nil
}

// ParseAscii returns a string and auto-strips the trailing NUL character.
func (tt TagType) ParseAscii(data []byte, unitCount uint32) (value string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(*Parser).ParseAscii()` should be used.

	value, err = parser.ParseAscii(data, unitCount)
	log.PanicIf(err)

	return value, nil
}

// ParseAsciiNoNul returns a string without any consideration for a trailing NUL
// character.
func (tt TagType) ParseAsciiNoNul(data []byte, unitCount uint32) (value string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(*Parser).ParseAsciiNoNul()` should be used.

	value, err = parser.ParseAsciiNoNul(data, unitCount)
	log.PanicIf(err)

	return value, nil
}

func (tt TagType) ParseShorts(data []byte, unitCount uint32) (value []uint16, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(*Parser).ParseShorts()` should be used.

	value, err = parser.ParseShorts(data, unitCount, tt.byteOrder)
	log.PanicIf(err)

	return value, nil
}

func (tt TagType) ParseLongs(data []byte, unitCount uint32) (value []uint32, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(*Parser).ParseLongs()` should be used.

	value, err = parser.ParseLongs(data, unitCount, tt.byteOrder)
	log.PanicIf(err)

	return value, nil
}

func (tt TagType) ParseRationals(data []byte, unitCount uint32) (value []Rational, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(*Parser).ParseRationals()` should be used.

	value, err = parser.ParseRationals(data, unitCount, tt.byteOrder)
	log.PanicIf(err)

	return value, nil
}

func (tt TagType) ParseSignedLongs(data []byte, unitCount uint32) (value []int32, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(*Parser).ParseSignedLongs()` should be used.

	value, err = parser.ParseSignedLongs(data, unitCount, tt.byteOrder)
	log.PanicIf(err)

	return value, nil
}

func (tt TagType) ParseSignedRationals(data []byte, unitCount uint32) (value []SignedRational, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(*Parser).ParseSignedRationals()` should be used.

	value, err = parser.ParseSignedRationals(data, unitCount, tt.byteOrder)
	log.PanicIf(err)

	return value, nil
}

func (tt TagType) ReadByteValues(valueContext ValueContext) (value []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(ValueContext).ReadBytes()` should be used.

	value, err = valueContext.ReadBytes()
	log.PanicIf(err)

	return value, nil
}

func (tt TagType) ReadAsciiValue(valueContext ValueContext) (value string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(ValueContext).ReadAscii()` should be used.

	value, err = valueContext.ReadAscii()
	log.PanicIf(err)

	return value, nil
}

func (tt TagType) ReadAsciiNoNulValue(valueContext ValueContext) (value string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(ValueContext).ReadAsciiNoNul()` should be used.

	value, err = valueContext.ReadAsciiNoNul()
	log.PanicIf(err)

	return value, nil
}

func (tt TagType) ReadShortValues(valueContext ValueContext) (value []uint16, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(ValueContext).ReadShorts()` should be used.

	value, err = valueContext.ReadShorts()
	log.PanicIf(err)

	return value, nil
}

func (tt TagType) ReadLongValues(valueContext ValueContext) (value []uint32, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(ValueContext).ReadLongs()` should be used.

	value, err = valueContext.ReadLongs()
	log.PanicIf(err)

	return value, nil
}

func (tt TagType) ReadRationalValues(valueContext ValueContext) (value []Rational, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(ValueContext).ReadRationals()` should be used.

	value, err = valueContext.ReadRationals()
	log.PanicIf(err)

	return value, nil
}

func (tt TagType) ReadSignedLongValues(valueContext ValueContext) (value []int32, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(ValueContext).ReadSignedLongs()` should be used.

	value, err = valueContext.ReadSignedLongs()
	log.PanicIf(err)

	return value, nil
}

func (tt TagType) ReadSignedRationalValues(valueContext ValueContext) (value []SignedRational, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(ValueContext).ReadSignedRationals()` should be used.

	value, err = valueContext.ReadSignedRationals()
	log.PanicIf(err)

	return value, nil
}

// ResolveAsString resolves the given value and returns a flat string.
//
// Where the type is not ASCII, `justFirst` indicates whether to just stringify
// the first item in the slice (or return an empty string if the slice is
// empty).
//
// Since this method lacks the information to process unknown-type tags (e.g.
// byte-order, tag-ID, IFD type), it will return an error if attempted. See
// `Undefined()`.
func (tt TagType) ResolveAsString(valueContext ValueContext, justFirst bool) (value string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if justFirst == true {
		value, err = valueContext.FormatFirst()
		log.PanicIf(err)
	} else {
		value, err = valueContext.Format()
		log.PanicIf(err)
	}

	return value, nil
}

// Resolve knows how to resolve the given value.
//
// Since this method lacks the information to process unknown-type tags (e.g.
// byte-order, tag-ID, IFD type), it will return an error if attempted. See
// `Undefined()`.
func (tt TagType) Resolve(valueContext *ValueContext) (values interface{}, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `(ValueContext).Values()` should be used.

	values, err = valueContext.Values()
	log.PanicIf(err)

	return values, nil
}

// Encode knows how to encode the given value to a byte slice.
func (tt TagType) Encode(value interface{}) (encoded []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ve := NewValueEncoder(tt.byteOrder)

	ed, err := ve.EncodeWithType(tt, value)
	log.PanicIf(err)

	return ed.Encoded, err
}

func (tt TagType) FromString(valueString string) (value interface{}, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// DEPRECATED(dustin): `EncodeStringToBytes()` should be used.

	value, err = EncodeStringToBytes(tt.tagType, valueString)
	log.PanicIf(err)

	return value, nil
}

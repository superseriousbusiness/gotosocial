package exif

import (
	"bytes"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

type Parser struct {
}

func (p *Parser) ParseBytes(data []byte, unitCount uint32) (value []uint8, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	count := int(unitCount)

	if len(data) < (TypeByte.Size() * count) {
		log.Panic(ErrNotEnoughData)
	}

	value = []uint8(data[:count])

	return value, nil
}

// ParseAscii returns a string and auto-strips the trailing NUL character.
func (p *Parser) ParseAscii(data []byte, unitCount uint32) (value string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	count := int(unitCount)

	if len(data) < (TypeAscii.Size() * count) {
		log.Panic(ErrNotEnoughData)
	}

	if len(data) == 0 || data[count-1] != 0 {
		s := string(data[:count])
		typeLogger.Warningf(nil, "ascii not terminated with nul as expected: [%v]", s)

		return s, nil
	} else {
		// Auto-strip the NUL from the end. It serves no purpose outside of
		// encoding semantics.

		return string(data[:count-1]), nil
	}
}

// ParseAsciiNoNul returns a string without any consideration for a trailing NUL
// character.
func (p *Parser) ParseAsciiNoNul(data []byte, unitCount uint32) (value string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	count := int(unitCount)

	if len(data) < (TypeAscii.Size() * count) {
		log.Panic(ErrNotEnoughData)
	}

	return string(data[:count]), nil
}

func (p *Parser) ParseShorts(data []byte, unitCount uint32, byteOrder binary.ByteOrder) (value []uint16, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	count := int(unitCount)

	if len(data) < (TypeShort.Size() * count) {
		log.Panic(ErrNotEnoughData)
	}

	value = make([]uint16, count)
	for i := 0; i < count; i++ {
		value[i] = byteOrder.Uint16(data[i*2:])
	}

	return value, nil
}

func (p *Parser) ParseLongs(data []byte, unitCount uint32, byteOrder binary.ByteOrder) (value []uint32, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	count := int(unitCount)

	if len(data) < (TypeLong.Size() * count) {
		log.Panic(ErrNotEnoughData)
	}

	value = make([]uint32, count)
	for i := 0; i < count; i++ {
		value[i] = byteOrder.Uint32(data[i*4:])
	}

	return value, nil
}

func (p *Parser) ParseRationals(data []byte, unitCount uint32, byteOrder binary.ByteOrder) (value []Rational, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	count := int(unitCount)

	if len(data) < (TypeRational.Size() * count) {
		log.Panic(ErrNotEnoughData)
	}

	value = make([]Rational, count)
	for i := 0; i < count; i++ {
		value[i].Numerator = byteOrder.Uint32(data[i*8:])
		value[i].Denominator = byteOrder.Uint32(data[i*8+4:])
	}

	return value, nil
}

func (p *Parser) ParseSignedLongs(data []byte, unitCount uint32, byteOrder binary.ByteOrder) (value []int32, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	count := int(unitCount)

	if len(data) < (TypeSignedLong.Size() * count) {
		log.Panic(ErrNotEnoughData)
	}

	b := bytes.NewBuffer(data)

	value = make([]int32, count)
	for i := 0; i < count; i++ {
		err := binary.Read(b, byteOrder, &value[i])
		log.PanicIf(err)
	}

	return value, nil
}

func (p *Parser) ParseSignedRationals(data []byte, unitCount uint32, byteOrder binary.ByteOrder) (value []SignedRational, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	count := int(unitCount)

	if len(data) < (TypeSignedRational.Size() * count) {
		log.Panic(ErrNotEnoughData)
	}

	b := bytes.NewBuffer(data)

	value = make([]SignedRational, count)
	for i := 0; i < count; i++ {
		err = binary.Read(b, byteOrder, &value[i].Numerator)
		log.PanicIf(err)

		err = binary.Read(b, byteOrder, &value[i].Denominator)
		log.PanicIf(err)
	}

	return value, nil
}

package exif

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"encoding/binary"
	"io/ioutil"

	"github.com/dsoprea/go-logging"
)

const (
	// ExifAddressableAreaStart is the absolute offset in the file that all
	// offsets are relative to.
	ExifAddressableAreaStart = uint32(0x0)

	// ExifDefaultFirstIfdOffset is essentially the number of bytes in addition
	// to `ExifAddressableAreaStart` that you have to move in order to escape
	// the rest of the header and get to the earliest point where we can put
	// stuff (which has to be the first IFD). This is the size of the header
	// sequence containing the two-character byte-order, two-character fixed-
	// bytes, and the four bytes describing the first-IFD offset.
	ExifDefaultFirstIfdOffset = uint32(2 + 2 + 4)
)

var (
	exifLogger = log.NewLogger("exif.exif")

	// EncodeDefaultByteOrder is the default byte-order for encoding operations.
	EncodeDefaultByteOrder = binary.BigEndian

	// Default byte order for tests.
	TestDefaultByteOrder = binary.BigEndian

	BigEndianBoBytes    = [2]byte{'M', 'M'}
	LittleEndianBoBytes = [2]byte{'I', 'I'}

	ByteOrderLookup = map[[2]byte]binary.ByteOrder{
		BigEndianBoBytes:    binary.BigEndian,
		LittleEndianBoBytes: binary.LittleEndian,
	}

	ByteOrderLookupR = map[binary.ByteOrder][2]byte{
		binary.BigEndian:    BigEndianBoBytes,
		binary.LittleEndian: LittleEndianBoBytes,
	}

	ExifFixedBytesLookup = map[binary.ByteOrder][2]byte{
		binary.LittleEndian: {0x2a, 0x00},
		binary.BigEndian:    {0x00, 0x2a},
	}
)

var (
	ErrNoExif          = errors.New("no exif data")
	ErrExifHeaderError = errors.New("exif header error")
)

// SearchAndExtractExif returns a slice from the beginning of the EXIF data to
// end of the file (it's not practical to try and calculate where the data
// actually ends; it needs to be formally parsed).
func SearchAndExtractExif(data []byte) (rawExif []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	// Search for the beginning of the EXIF information. The EXIF is near the
	// beginning of our/most JPEGs, so this has a very low cost.

	foundAt := -1
	for i := 0; i < len(data); i++ {
		if _, err := ParseExifHeader(data[i:]); err == nil {
			foundAt = i
			break
		} else if log.Is(err, ErrNoExif) == false {
			return nil, err
		}
	}

	if foundAt == -1 {
		return nil, ErrNoExif
	}

	return data[foundAt:], nil
}

// SearchFileAndExtractExif returns a slice from the beginning of the EXIF data
// to the end of the file (it's not practical to try and calculate where the
// data actually ends).
func SearchFileAndExtractExif(filepath string) (rawExif []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	// Open the file.

	f, err := os.Open(filepath)
	log.PanicIf(err)

	defer f.Close()

	data, err := ioutil.ReadAll(f)
	log.PanicIf(err)

	rawExif, err = SearchAndExtractExif(data)
	log.PanicIf(err)

	return rawExif, nil
}

type ExifHeader struct {
	ByteOrder      binary.ByteOrder
	FirstIfdOffset uint32
}

func (eh ExifHeader) String() string {
	return fmt.Sprintf("ExifHeader<BYTE-ORDER=[%v] FIRST-IFD-OFFSET=(0x%02x)>", eh.ByteOrder, eh.FirstIfdOffset)
}

// ParseExifHeader parses the bytes at the very top of the header.
//
// This will panic with ErrNoExif on any data errors so that we can double as
// an EXIF-detection routine.
func ParseExifHeader(data []byte) (eh ExifHeader, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// Good reference:
	//
	//      CIPA DC-008-2016; JEITA CP-3451D
	//      -> http://www.cipa.jp/std/documents/e/DC-008-Translation-2016-E.pdf

	if len(data) < 2 {
		exifLogger.Warningf(nil, "Not enough data for EXIF header (1): (%d)", len(data))
		return eh, ErrNoExif
	}

	byteOrderBytes := [2]byte{data[0], data[1]}

	byteOrder, found := ByteOrderLookup[byteOrderBytes]
	if found == false {
		// exifLogger.Warningf(nil, "EXIF byte-order not recognized: [%v]", byteOrderBytes)
		return eh, ErrNoExif
	}

	if len(data) < 4 {
		exifLogger.Warningf(nil, "Not enough data for EXIF header (2): (%d)", len(data))
		return eh, ErrNoExif
	}

	fixedBytes := [2]byte{data[2], data[3]}
	expectedFixedBytes := ExifFixedBytesLookup[byteOrder]
	if fixedBytes != expectedFixedBytes {
		// exifLogger.Warningf(nil, "EXIF header fixed-bytes should be [%v] but are: [%v]", expectedFixedBytes, fixedBytes)
		return eh, ErrNoExif
	}

	if len(data) < 2 {
		exifLogger.Warningf(nil, "Not enough data for EXIF header (3): (%d)", len(data))
		return eh, ErrNoExif
	}

	firstIfdOffset := byteOrder.Uint32(data[4:8])

	eh = ExifHeader{
		ByteOrder:      byteOrder,
		FirstIfdOffset: firstIfdOffset,
	}

	return eh, nil
}

// Visit recursively invokes a callback for every tag.
func Visit(rootIfdName string, ifdMapping *IfdMapping, tagIndex *TagIndex, exifData []byte, visitor RawTagVisitor) (eh ExifHeader, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	eh, err = ParseExifHeader(exifData)
	log.PanicIf(err)

	ie := NewIfdEnumerate(ifdMapping, tagIndex, exifData, eh.ByteOrder)

	err = ie.Scan(rootIfdName, eh.FirstIfdOffset, visitor, true)
	log.PanicIf(err)

	return eh, nil
}

// Collect recursively builds a static structure of all IFDs and tags.
func Collect(ifdMapping *IfdMapping, tagIndex *TagIndex, exifData []byte) (eh ExifHeader, index IfdIndex, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	eh, err = ParseExifHeader(exifData)
	log.PanicIf(err)

	ie := NewIfdEnumerate(ifdMapping, tagIndex, exifData, eh.ByteOrder)

	index, err = ie.Collect(eh.FirstIfdOffset, true)
	log.PanicIf(err)

	return eh, index, nil
}

// BuildExifHeader constructs the bytes that go in the very beginning.
func BuildExifHeader(byteOrder binary.ByteOrder, firstIfdOffset uint32) (headerBytes []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	b := new(bytes.Buffer)

	// This is the point in the data that all offsets are relative to.
	boBytes := ByteOrderLookupR[byteOrder]
	_, err = b.WriteString(string(boBytes[:]))
	log.PanicIf(err)

	fixedBytes := ExifFixedBytesLookup[byteOrder]

	_, err = b.Write(fixedBytes[:])
	log.PanicIf(err)

	err = binary.Write(b, byteOrder, firstIfdOffset)
	log.PanicIf(err)

	return b.Bytes(), nil
}

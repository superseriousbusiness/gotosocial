package photoshopinfo

import (
	"fmt"
	"io"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

var (
	defaultByteOrder = binary.BigEndian
)

// Photoshop30InfoRecord is the data for one parsed Photoshop-info record.
type Photoshop30InfoRecord struct {
	// RecordType is the record-type.
	RecordType string

	// ImageResourceId is the image resource-ID.
	ImageResourceId uint16

	// Name is the name of the record. It is optional and will be an empty-
	// string if not present.
	Name string

	// Data is the raw record data.
	Data []byte
}

// String returns a descriptive string.
func (pir Photoshop30InfoRecord) String() string {
	return fmt.Sprintf("RECORD-TYPE=[%s] IMAGE-RESOURCE-ID=[0x%04x] NAME=[%s] DATA-SIZE=(%d)", pir.RecordType, pir.ImageResourceId, pir.Name, len(pir.Data))
}

// ReadPhotoshop30InfoRecord parses a single photoshop-info record.
func ReadPhotoshop30InfoRecord(r io.Reader) (pir Photoshop30InfoRecord, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	recordType := make([]byte, 4)
	_, err = io.ReadFull(r, recordType)
	if err != nil {
		if err == io.EOF {
			return pir, err
		}

		log.Panic(err)
	}

	// TODO(dustin): Move BigEndian to constant/config.

	irId := uint16(0)
	err = binary.Read(r, defaultByteOrder, &irId)
	log.PanicIf(err)

	nameSize := uint8(0)
	err = binary.Read(r, defaultByteOrder, &nameSize)
	log.PanicIf(err)

	// Add an extra byte if the two length+data size is odd to make the total
	// bytes read even.
	doAddPadding := (1+nameSize)%2 == 1
	if doAddPadding == true {
		nameSize++
	}

	name := make([]byte, nameSize)
	_, err = io.ReadFull(r, name)
	log.PanicIf(err)

	// If the last byte is padding, truncate it.
	if doAddPadding == true {
		name = name[:nameSize-1]
	}

	dataSize := uint32(0)
	err = binary.Read(r, defaultByteOrder, &dataSize)
	log.PanicIf(err)

	data := make([]byte, dataSize+dataSize%2)
	_, err = io.ReadFull(r, data)
	log.PanicIf(err)

	data = data[:dataSize]

	pir = Photoshop30InfoRecord{
		RecordType:      string(recordType),
		ImageResourceId: irId,
		Name:            string(name),
		Data:            data,
	}

	return pir, nil
}

// ReadPhotoshop30Info parses a sequence of photoship-info records from the stream.
func ReadPhotoshop30Info(r io.Reader) (pirIndex map[uint16]Photoshop30InfoRecord, err error) {
	pirIndex = make(map[uint16]Photoshop30InfoRecord)

	for {
		pir, err := ReadPhotoshop30InfoRecord(r)
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Panic(err)
		}

		pirIndex[pir.ImageResourceId] = pir
	}

	return pirIndex, nil
}

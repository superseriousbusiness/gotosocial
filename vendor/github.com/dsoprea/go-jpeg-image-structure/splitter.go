package jpegstructure

import (
	"bufio"
	"bytes"
	"io"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

// JpegSplitter uses the Go stream splitter to divide the JPEG stream into
// segments.
type JpegSplitter struct {
	lastMarkerId   byte
	lastMarkerName string
	counter        int
	lastIsScanData bool
	visitor        interface{}

	currentOffset int
	segments      *SegmentList

	scandataOffset int
}

// NewJpegSplitter returns a new JpegSplitter.
func NewJpegSplitter(visitor interface{}) *JpegSplitter {
	return &JpegSplitter{
		segments: NewSegmentList(nil),
		visitor:  visitor,
	}
}

// Segments returns all found segments.
func (js *JpegSplitter) Segments() *SegmentList {
	return js.segments
}

// MarkerId returns the ID of the last processed marker.
func (js *JpegSplitter) MarkerId() byte {
	return js.lastMarkerId
}

// MarkerName returns the name of the last-processed marker.
func (js *JpegSplitter) MarkerName() string {
	return js.lastMarkerName
}

// Counter returns the number of processed segments.
func (js *JpegSplitter) Counter() int {
	return js.counter
}

// IsScanData returns whether the last processed segment was scan-data.
func (js *JpegSplitter) IsScanData() bool {
	return js.lastIsScanData
}

func (js *JpegSplitter) processScanData(data []byte) (advanceBytes int, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// Search through the segment, past all 0xff's therein, until we encounter
	// the EOI segment.

	dataLength := -1
	for i := js.scandataOffset; i < len(data); i++ {
		thisByte := data[i]

		if i == 0 {
			continue
		}

		lastByte := data[i-1]
		if lastByte != 0xff {
			continue
		}

		if thisByte == 0x00 || thisByte >= 0xd0 && thisByte <= 0xd8 {
			continue
		}

		// After all of the other checks, this means that we're on the EOF
		// segment.
		if thisByte != MARKER_EOI {
			continue
		}

		dataLength = i - 1
		break
	}

	if dataLength == -1 {
		// On the next pass, start on the last byte of this pass, just in case
		// the first byte of the two-byte sequence is here.
		js.scandataOffset = len(data) - 1

		jpegLogger.Debugf(nil, "Scan-data not fully available (%d).", len(data))
		return 0, nil
	}

	js.lastIsScanData = true
	js.lastMarkerId = 0
	js.lastMarkerName = ""

	// Note that we don't increment the counter since this isn't an actual
	// segment.

	jpegLogger.Debugf(nil, "End of scan-data.")

	err = js.handleSegment(0x0, "!SCANDATA", 0x0, data[:dataLength])
	log.PanicIf(err)

	return dataLength, nil
}

func (js *JpegSplitter) readSegment(data []byte) (count int, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if js.counter == 0 {
		// Verify magic bytes.

		if len(data) < 3 {
			jpegLogger.Debugf(nil, "Not enough (1)")
			return 0, nil
		}

		if data[0] == jpegMagic2000[0] && data[1] == jpegMagic2000[1] && data[2] == jpegMagic2000[2] {
			// TODO(dustin): Revisit JPEG2000 support.
			log.Panicf("JPEG2000 not supported")
		}

		if data[0] != jpegMagicStandard[0] || data[1] != jpegMagicStandard[1] || data[2] != jpegMagicStandard[2] {
			log.Panicf("file does not look like a JPEG: (%02x) (%02x) (%02x)", data[0], data[1], data[2])
		}
	}

	chunkLength := len(data)

	jpegLogger.Debugf(nil, "SPLIT: LEN=(%d) COUNTER=(%d)", chunkLength, js.counter)

	if js.scanDataIsNext() == true {
		// If the last segment was the SOS, we're currently sitting on scan data.
		// Search for the EOI marker afterward in order to know how much data
		// there is. Return this as its own token.
		//
		// REF: https://stackoverflow.com/questions/26715684/parsing-jpeg-sos-marker

		advanceBytes, err := js.processScanData(data)
		log.PanicIf(err)

		// This will either return 0 and implicitly request that we need more
		// data and then need to run again or will return an actual byte count
		// to progress by.

		return advanceBytes, nil
	} else if js.lastMarkerId == MARKER_EOI {
		// We have more data following the EOI, which is unexpected. There
		// might be non-standard cruft at the end of the file. Terminate the
		// parse because the file-structure is, technically, complete at this
		// point.

		return 0, io.EOF
	} else {
		js.lastIsScanData = false
	}

	// If we're here, we're supposed to be sitting on the 0xff bytes at the
	// beginning of a segment (just before the marker).

	if data[0] != 0xff {
		log.Panicf("not on new segment marker @ (%d): (%02X)", js.currentOffset, data[0])
	}

	i := 0
	found := false
	for ; i < chunkLength; i++ {
		jpegLogger.Debugf(nil, "Prefix check: (%d) %02X", i, data[i])

		if data[i] != 0xff {
			found = true
			break
		}
	}

	jpegLogger.Debugf(nil, "Skipped over leading 0xFF bytes: (%d)", i)

	if found == false || i >= chunkLength {
		jpegLogger.Debugf(nil, "Not enough (3)")
		return 0, nil
	}

	markerId := data[i]

	js.lastMarkerName = markerNames[markerId]

	sizeLen, found := markerLen[markerId]
	jpegLogger.Debugf(nil, "MARKER-ID=%x SIZELEN=%v FOUND=%v", markerId, sizeLen, found)

	i++

	b := bytes.NewBuffer(data[i:])
	payloadLength := 0

	// marker-ID + size => 2 + <dynamic>
	headerSize := 2 + sizeLen

	if found == false {

		// It's not one of the static-length markers. Read the length.
		//
		// The length is an unsigned 16-bit network/big-endian.

		// marker-ID + size => 2 + 2
		headerSize = 2 + 2

		if i+2 >= chunkLength {
			jpegLogger.Debugf(nil, "Not enough (4)")
			return 0, nil
		}

		l := uint16(0)
		err = binary.Read(b, binary.BigEndian, &l)
		log.PanicIf(err)

		if l <= 2 {
			log.Panicf("length of size read for non-special marker (%02x) is unexpectedly not more than two.", markerId)
		}

		// (l includes the bytes of the length itself.)
		payloadLength = int(l) - 2
		jpegLogger.Debugf(nil, "DataLength (dynamically-sized segment): (%d)", payloadLength)

		i += 2
	} else if sizeLen > 0 {

		// Accommodates the non-zero markers in our marker index, which only
		// represent J2C extensions.
		//
		// The length is an unsigned 32-bit network/big-endian.

		// TODO(dustin): !! This needs to be tested, but we need an image.

		if sizeLen != 4 {
			log.Panicf("known non-zero marker is not four bytes, which is not currently handled: M=(%x)", markerId)
		}

		if i+4 >= chunkLength {
			jpegLogger.Debugf(nil, "Not enough (5)")
			return 0, nil
		}

		l := uint32(0)
		err = binary.Read(b, binary.BigEndian, &l)
		log.PanicIf(err)

		payloadLength = int(l) - 4
		jpegLogger.Debugf(nil, "DataLength (four-byte-length segment): (%u)", l)

		i += 4
	}

	jpegLogger.Debugf(nil, "PAYLOAD-LENGTH: %d", payloadLength)

	payload := data[i:]

	if payloadLength < 0 {
		log.Panicf("payload length less than zero: (%d)", payloadLength)
	}

	i += int(payloadLength)

	if i > chunkLength {
		jpegLogger.Debugf(nil, "Not enough (6)")
		return 0, nil
	}

	jpegLogger.Debugf(nil, "Found whole segment.")

	js.lastMarkerId = markerId

	payloadWindow := payload[:payloadLength]
	err = js.handleSegment(markerId, js.lastMarkerName, headerSize, payloadWindow)
	log.PanicIf(err)

	js.counter++

	jpegLogger.Debugf(nil, "Returning advance of (%d)", i)

	return i, nil
}

func (js *JpegSplitter) scanDataIsNext() bool {
	return js.lastMarkerId == MARKER_SOS
}

// Split is the base splitting function that satisfies `bufio.SplitFunc`.
func (js *JpegSplitter) Split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	for len(data) > 0 {
		currentAdvance, err := js.readSegment(data)
		if err != nil {
			if err == io.EOF {
				// We've encountered an EOI marker.
				return 0, nil, err
			}

			log.Panic(err)
		}

		if currentAdvance == 0 {
			if len(data) > 0 && atEOF == true {
				// Provide a little context in the error message.

				if js.scanDataIsNext() == true {
					// Yes, we've ran into this.

					log.Panicf("scan-data is unbounded; EOI not encountered before EOF")
				} else {
					log.Panicf("partial segment data encountered before scan-data")
				}
			}

			// We don't have enough data for another segment.
			break
		}

		data = data[currentAdvance:]
		advance += currentAdvance
	}

	return advance, nil, nil
}

func (js *JpegSplitter) parseSof(data []byte) (sof *SofSegment, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	stream := bytes.NewBuffer(data)
	buffer := bufio.NewReader(stream)

	bitsPerSample, err := buffer.ReadByte()
	log.PanicIf(err)

	height := uint16(0)
	err = binary.Read(buffer, binary.BigEndian, &height)
	log.PanicIf(err)

	width := uint16(0)
	err = binary.Read(buffer, binary.BigEndian, &width)
	log.PanicIf(err)

	componentCount, err := buffer.ReadByte()
	log.PanicIf(err)

	sof = &SofSegment{
		BitsPerSample:  bitsPerSample,
		Width:          width,
		Height:         height,
		ComponentCount: componentCount,
	}

	return sof, nil
}

func (js *JpegSplitter) parseAppData(markerId byte, data []byte) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	return nil
}

func (js *JpegSplitter) handleSegment(markerId byte, markerName string, headerSize int, payload []byte) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	cloned := make([]byte, len(payload))
	copy(cloned, payload)

	s := &Segment{
		MarkerId:   markerId,
		MarkerName: markerName,
		Offset:     js.currentOffset,
		Data:       cloned,
	}

	jpegLogger.Debugf(nil, "Encountered marker (0x%02x) [%s] at offset (%d)", markerId, markerName, js.currentOffset)

	js.currentOffset += headerSize + len(payload)

	js.segments.Add(s)

	sv, ok := js.visitor.(SegmentVisitor)
	if ok == true {
		err = sv.HandleSegment(js.lastMarkerId, js.lastMarkerName, js.counter, js.lastIsScanData)
		log.PanicIf(err)
	}

	if markerId >= MARKER_SOF0 && markerId <= MARKER_SOF15 {
		ssv, ok := js.visitor.(SofSegmentVisitor)
		if ok == true {
			sof, err := js.parseSof(payload)
			log.PanicIf(err)

			err = ssv.HandleSof(sof)
			log.PanicIf(err)
		}
	} else if markerId >= MARKER_APP0 && markerId <= MARKER_APP15 {
		err := js.parseAppData(markerId, payload)
		log.PanicIf(err)
	}

	return nil
}

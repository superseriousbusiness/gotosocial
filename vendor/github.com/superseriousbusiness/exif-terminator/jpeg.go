/*
   exif-terminator
   Copyright (C) 2022 SuperSeriousBusiness admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package terminator

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	exif "github.com/dsoprea/go-exif/v3"
	jpegstructure "github.com/superseriousbusiness/go-jpeg-image-structure/v2"
)

var markerLen = map[byte]int{
	0x00: 0,
	0x01: 0,
	0xd0: 0,
	0xd1: 0,
	0xd2: 0,
	0xd3: 0,
	0xd4: 0,
	0xd5: 0,
	0xd6: 0,
	0xd7: 0,
	0xd8: 0,
	0xd9: 0,
	0xda: 0,

	// J2C
	0x30: 0,
	0x31: 0,
	0x32: 0,
	0x33: 0,
	0x34: 0,
	0x35: 0,
	0x36: 0,
	0x37: 0,
	0x38: 0,
	0x39: 0,
	0x3a: 0,
	0x3b: 0,
	0x3c: 0,
	0x3d: 0,
	0x3e: 0,
	0x3f: 0,
	0x4f: 0,
	0x92: 0,
	0x93: 0,

	// J2C extensions
	0x74: 4,
	0x75: 4,
	0x77: 4,
}

type jpegVisitor struct {
	js                *jpegstructure.JpegSplitter
	writer            io.Writer
	expectedFileSize  int
	writtenTotalBytes int
}

// HandleSegment satisfies the visitor interface{} of the jpegstructure library.
//
// We don't really care about many of the parameters, since all we're interested
// in here is the very last segment that was scanned.
func (v *jpegVisitor) HandleSegment(segmentMarker byte, _ string, _ int, _ bool) error {
	// get the most recent segment scanned (ie., last in the segments list)
	segmentList := v.js.Segments()
	segments := segmentList.Segments()
	mostRecentSegment := segments[len(segments)-1]

	// check if we've written the expected number of bytes by EOI
	if segmentMarker == jpegstructure.MARKER_EOI {
		// take account of the last 2 bytes taken up by the EOI
		eoiLength := 2

		// this is the total file size we will
		// have written including the EOI
		willHaveWritten := v.writtenTotalBytes + eoiLength

		if willHaveWritten < v.expectedFileSize {
			// if we won't have written enough,
			// pad the final segment before EOI
			// so that we meet expected file size
			missingBytes := make([]byte, v.expectedFileSize-willHaveWritten)
			if _, err := v.writer.Write(missingBytes); err != nil {
				return err
			}
		}
	}

	// process the segment
	return v.writeSegment(mostRecentSegment)
}

func (v *jpegVisitor) writeSegment(s *jpegstructure.Segment) error {
	var writtenSegmentData int
	w := v.writer

	defer func() {
		// whatever happens, when we finished then evict data from the segment;
		// once we've written it we don't want it in memory anymore
		s.Data = s.Data[:0]
	}()

	// The scan-data will have a marker-ID of (0) because it doesn't have a marker-ID or length.
	if s.MarkerId != 0 {
		markerIDWritten, err := w.Write([]byte{0xff, s.MarkerId})
		if err != nil {
			return err
		}
		writtenSegmentData += markerIDWritten

		sizeLen, found := markerLen[s.MarkerId]
		if !found || sizeLen == 2 {
			sizeLen = 2
			l := uint16(len(s.Data) + sizeLen)

			if err := binary.Write(w, binary.BigEndian, &l); err != nil {
				return err
			}

			writtenSegmentData += 2
		} else if sizeLen == 4 {
			l := uint32(len(s.Data) + sizeLen)

			if err := binary.Write(w, binary.BigEndian, &l); err != nil {
				return err
			}

			writtenSegmentData += 4
		} else if sizeLen != 0 {
			return fmt.Errorf("not a supported marker-size: MARKER-ID=(0x%02x) MARKER-SIZE-LEN=(%d)", s.MarkerId, sizeLen)
		}
	}

	if !s.IsExif() {
		// if this isn't exif data just copy it over and bail
		writtenNormalData, err := w.Write(s.Data)
		if err != nil {
			return err
		}

		writtenSegmentData += writtenNormalData
		v.writtenTotalBytes += writtenSegmentData
		return nil
	}

	ifd, _, err := s.Exif()
	if err != nil {
		return err
	}

	// amount of bytes we've writtenExifData into the exif body, we'll update this as we go
	var writtenExifData int

	if orientationEntries, err := ifd.FindTagWithName("Orientation"); err == nil && len(orientationEntries) == 1 {
		// If we have an orientation entry, we don't want to completely obliterate the exif data.
		// Instead, we want to surgically obliterate everything *except* the orientation tag, so
		// that the image will still be rotated correctly when shown in client applications etc.
		//
		// To accomplish this, we're going to extract just the bytes that we need and write them
		// in according to the exif specification, then fill in the rest of the space with empty
		// bytes.
		//
		// First we need to write the exif prefix for this segment.
		//
		// Then we write the exif header which contains the byte order and offset of the first ifd.
		//
		// Then we write the ifd0 entry which contains the orientation data.
		//
		// After that we just fill.

		newExifData := &bytes.Buffer{}
		byteOrder := ifd.ByteOrder()

		// 1. Write exif prefix.
		// https://www.ozhiker.com/electronics/pjmt/jpeg_info/app_segments.html
		prefix := []byte{'E', 'x', 'i', 'f', 0, 0}
		if err := binary.Write(newExifData, byteOrder, &prefix); err != nil {
			return err
		}
		writtenExifData += len(prefix)

		// 2. Write exif header, taking the existing byte order.
		exifHeader, err := exif.BuildExifHeader(byteOrder, exif.ExifDefaultFirstIfdOffset)
		if err != nil {
			return err
		}
		hWritten, err := newExifData.Write(exifHeader)
		if err != nil {
			return err
		}
		writtenExifData += hWritten

		// 3. Write in the new ifd
		//
		// An ifd with one orientation entry is structured like this:
		// 		2 bytes: the number of entries in the ifd	uint16(1)
		// 		2 bytes: the tag id							uint16(274)
		// 		2 bytes: the tag type						uint16(3)
		//      4 bytes: the tag count						uint32(1)
		// 		4 bytes: the tag value offset:				uint32(one of the below with padding on the end)
		// 			1 = Horizontal (normal)
		// 			2 = Mirror horizontal
		// 			3 = Rotate 180
		// 			4 = Mirror vertical
		// 			5 = Mirror horizontal and rotate 270 CW
		// 			6 = Rotate 90 CW
		// 			7 = Mirror horizontal and rotate 90 CW
		// 			8 = Rotate 270 CW
		//
		// see https://web.archive.org/web/20190624045241if_/http://www.cipa.jp:80/std/documents/e/DC-008-Translation-2019-E.pdf - p24-25
		orientationEntry := orientationEntries[0]

		ifdCount := uint16(1) // we're only adding one entry into the ifd
		if err := binary.Write(newExifData, byteOrder, &ifdCount); err != nil {
			return err
		}
		writtenExifData += 2

		tagID := orientationEntry.TagId()
		if err := binary.Write(newExifData, byteOrder, &tagID); err != nil {
			return err
		}
		writtenExifData += 2

		tagType := uint16(orientationEntry.TagType())
		if err := binary.Write(newExifData, byteOrder, &tagType); err != nil {
			return err
		}
		writtenExifData += 2

		tagCount := orientationEntry.UnitCount()
		if err := binary.Write(newExifData, byteOrder, &tagCount); err != nil {
			return err
		}
		writtenExifData += 4

		valueOffset, err := orientationEntry.GetRawBytes()
		if err != nil {
			return err
		}

		vWritten, err := newExifData.Write(valueOffset)
		if err != nil {
			return err
		}
		writtenExifData += vWritten

		valuePad := make([]byte, 4-vWritten)
		pWritten, err := newExifData.Write(valuePad)
		if err != nil {
			return err
		}
		writtenExifData += pWritten

		// write all the new data into the writer from the segment
		writtenNewExifData, err := io.Copy(w, newExifData)
		if err != nil {
			return err
		}

		writtenSegmentData += int(writtenNewExifData)
	}

	// fill in any remaining exif body with blank bytes
	blank := make([]byte, len(s.Data)-writtenExifData)
	writtenPadding, err := w.Write(blank)
	if err != nil {
		return err
	}

	writtenSegmentData += writtenPadding
	v.writtenTotalBytes += writtenSegmentData
	return nil
}

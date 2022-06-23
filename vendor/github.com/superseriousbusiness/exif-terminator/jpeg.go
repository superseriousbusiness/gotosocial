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
	js     *jpegstructure.JpegSplitter
	writer io.Writer
}

// HandleSegment satisfies the visitor interface{} of the jpegstructure library.
//
// We don't really care about any of the parameters, since all we're interested
// in here is the very last segment that was scanned.
func (v *jpegVisitor) HandleSegment(_ byte, _ string, _ int, _ bool) error {
	// all we want to do here is get the last segment that was scanned, and then manipulate it
	segmentList := v.js.Segments()
	segments := segmentList.Segments()
	lastSegment := segments[len(segments)-1]
	return v.writeSegment(lastSegment)
}

func (v *jpegVisitor) writeSegment(s *jpegstructure.Segment) error {
	w := v.writer

	defer func() {
		// whatever happens, when we finished then evict data from the segment;
		// once we've written it we don't want it in memory anymore
		s.Data = s.Data[:0]
	}()

	// The scan-data will have a marker-ID of (0) because it doesn't have a marker-ID or length.
	if s.MarkerId != 0 {
		if _, err := w.Write([]byte{0xff, s.MarkerId}); err != nil {
			return err
		}

		sizeLen, found := markerLen[s.MarkerId]
		if !found || sizeLen == 2 {
			sizeLen = 2
			l := uint16(len(s.Data) + sizeLen)

			if err := binary.Write(w, binary.BigEndian, &l); err != nil {
				return err
			}

		} else if sizeLen == 4 {
			l := uint32(len(s.Data) + sizeLen)

			if err := binary.Write(w, binary.BigEndian, &l); err != nil {
				return err
			}

		} else if sizeLen != 0 {
			return fmt.Errorf("not a supported marker-size: MARKER-ID=(0x%02x) MARKER-SIZE-LEN=(%d)", s.MarkerId, sizeLen)
		}
	}

	if !s.IsExif() {
		// if this isn't exif data just copy it over and bail
		_, err := w.Write(s.Data)
		return err
	}

	ifd, _, err := s.Exif()
	if err != nil {
		return err
	}

	// amount of bytes we've written into the exif body
	var written int

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
		// After that we just fill fill fill.

		newData := &bytes.Buffer{}

		// 1. Write exif prefix.
		// https://www.ozhiker.com/electronics/pjmt/jpeg_info/app_segments.html
		prefix := []byte{'E', 'x', 'i', 'f', 0, 0}
		if err := binary.Write(newData, ifd.ByteOrder(), &prefix); err != nil {
			return err
		}
		written += 6

		// 2. Write exif header, taking the existing byte order.
		exifHeader, err := exif.BuildExifHeader(ifd.ByteOrder(), exif.ExifDefaultFirstIfdOffset)
		if err != nil {
			return err
		}
		hWritten, err := newData.Write(exifHeader)
		if err != nil {
			return err
		}
		written += hWritten

		// https://web.archive.org/web/20190624045241if_/http://www.cipa.jp:80/std/documents/e/DC-008-Translation-2019-E.pdf
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
		orientationEntry := orientationEntries[0]

		ifdCount := uint16(1) // we're only adding one entry into the ifd
		if err := binary.Write(newData, ifd.ByteOrder(), &ifdCount); err != nil {
			return err
		}
		written += 2

		tagID := orientationEntry.TagId()
		if err := binary.Write(newData, ifd.ByteOrder(), &tagID); err != nil {
			return err
		}
		written += 2

		tagType := orientationEntry.TagType()
		if err := binary.Write(newData, ifd.ByteOrder(), &tagType); err != nil {
			return err
		}
		written += 2

		tagCount := orientationEntry.UnitCount()
		if err := binary.Write(newData, ifd.ByteOrder(), &tagCount); err != nil {
			return err
		}
		written += 4

		valueOffset, err := orientationEntry.GetRawBytes()
		if err != nil {
			return err
		}

		vWritten, err := newData.Write(valueOffset)
		if err != nil {
			return err
		}
		written += vWritten

		valuePad := make([]byte, 4-vWritten)
		pWritten, err := newData.Write(valuePad)
		if err != nil {
			return err
		}
		written += pWritten

		// write everything in
		if _, err := io.Copy(w, newData); err != nil {
			return err
		}
	}

	// fill in the (remaining) exif body with blank bytes
	blank := make([]byte, len(s.Data)-written)
	if _, err := w.Write(blank); err != nil {
		return err
	}

	return nil
}

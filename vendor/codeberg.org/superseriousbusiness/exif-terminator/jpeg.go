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
	"encoding/binary"
	"fmt"
	"io"

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
// We don't really care about many of the parameters, since all we're interested
// in here is the very last segment that was scanned.
func (v *jpegVisitor) HandleSegment(segmentMarker byte, _ string, _ int, _ bool) error {
	segmentList := v.js.Segments()
	segments := segmentList.Segments()
	mostRecentSegment := segments[len(segments)-1]
	return v.writeSegment(mostRecentSegment)
}

func (v *jpegVisitor) writeSegment(s *jpegstructure.Segment) error {
	w := v.writer

	defer func() {
		// whatever happens, when we finished then evict data from the segment;
		// once we've written it we don't want it in memory anymore
		s.Data = nil
	}()

	if s.IsExif() {
		// Segment contains exif data, terminate!
		if err := terminateEXIF(s); err != nil {
			return err
		}
	}

	// The scan-data will have a marker-ID of (0) because it doesn't have a marker-ID or length.
	if s.MarkerId != 0 {
		_, err := w.Write([]byte{0xff, s.MarkerId})
		if err != nil {
			return err
		}

		sizeLen, found := markerLen[s.MarkerId]
		if !found || sizeLen == 2 {
			l := uint16(len(s.Data) + 2)
			b := make([]byte, 2)
			binary.BigEndian.PutUint16(b, l)
			if _, err := w.Write(b); err != nil {
				return err
			}
		} else if sizeLen == 4 {
			l := uint32(len(s.Data) + 4)
			b := make([]byte, 4)
			binary.BigEndian.PutUint32(b, l)
			if _, err := w.Write(b); err != nil {
				return err
			}
		} else if sizeLen != 0 {
			return fmt.Errorf("not a supported marker-size: MARKER-ID=(0x%02x) MARKER-SIZE-LEN=(%d)", s.MarkerId, sizeLen)
		}
	}

	_, err := w.Write(s.Data)
	return err
}

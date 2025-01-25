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
	"errors"
	"io"
)

const (
	riffHeader = "RIFF"
	webpHeader = "WEBP"
	exifFourcc = "EXIF"
	xmpFourcc  = "XMP "
)

var (
	errNoRiffHeader = errors.New("no RIFF header")
	errNoWebpHeader = errors.New("not a WEBP file")
	errInvalidChunk = errors.New("invalid chunk")
)

type webpVisitor struct {
	writer     io.Writer
	doneHeader bool
}

func (v *webpVisitor) split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// parse/write the header first
	if !v.doneHeader {

		// const rifHeaderSize = 12
		if len(data) < 12 {
			if atEOF {
				err = errNoRiffHeader
			}
			return
		}

		if string(data[:4]) != riffHeader {
			err = errNoRiffHeader
			return
		}

		if string(data[8:12]) != webpHeader {
			err = errNoWebpHeader
			return
		}

		if _, err = v.writer.Write(data[:12]); err != nil {
			return
		}

		advance += 12
		data = data[12:]
		v.doneHeader = true
	}

	for {
		// need enough for
		// fourcc and size
		if len(data) < 8 {
			return
		}

		size := int64(binary.LittleEndian.Uint32(data[4:]))

		if (size & 1) != 0 {
			// odd chunk size:
			// extra padding byte
			size++
		}

		// wait until there is enough
		if int64(len(data)) < 8+size {
			return
		}

		// replace exif/xmp with blank
		switch string(data[:4]) {
		case exifFourcc, xmpFourcc:
			clear(data[8 : 8+size])
		}

		if _, err = v.writer.Write(data[:8+size]); err != nil {
			return
		}

		advance += 8 + int(size)
		data = data[8+size:]
	}
}

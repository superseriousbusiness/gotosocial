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
	riffHeaderSize = 4 * 3
)

var (
	riffHeader = [4]byte{'R', 'I', 'F', 'F'}
	webpHeader = [4]byte{'W', 'E', 'B', 'P'}
	exifFourcc = [4]byte{'E', 'X', 'I', 'F'}
	xmpFourcc  = [4]byte{'X', 'M', 'P', ' '}

	errNoRiffHeader = errors.New("no RIFF header")
	errNoWebpHeader = errors.New("not a WEBP file")
)

type webpVisitor struct {
	writer     io.Writer
	doneHeader bool
}

func fourCC(b []byte) [4]byte {
	return [4]byte{b[0], b[1], b[2], b[3]}
}

func (v *webpVisitor) split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// parse/write the header first
	if !v.doneHeader {
		if len(data) < riffHeaderSize {
			// need the full header
			return
		}
		if fourCC(data) != riffHeader {
			err = errNoRiffHeader
			return
		}
		if fourCC(data[8:]) != webpHeader {
			err = errNoWebpHeader
			return
		}
		if _, err = v.writer.Write(data[:riffHeaderSize]); err != nil {
			return
		}
		advance += riffHeaderSize
		data = data[riffHeaderSize:]
		v.doneHeader = true
	}

	// need enough for fourcc and size
	if len(data) < 8 {
		return
	}
	size := int64(binary.LittleEndian.Uint32(data[4:]))
	if (size & 1) != 0 {
		// odd chunk size - extra padding byte
		size++
	}
	// wait until there is enough
	if int64(len(data)-8) < size {
		return
	}

	fourcc := fourCC(data)
	rawChunkData := data[8 : 8+size]
	if fourcc == exifFourcc || fourcc == xmpFourcc {
		// replace exif/xmp with blank
		rawChunkData = make([]byte, size)
	}

	if _, err = v.writer.Write(data[:8]); err == nil {
		if _, err = v.writer.Write(rawChunkData); err == nil {
			advance += 8 + int(size)
		}
	}

	return
}

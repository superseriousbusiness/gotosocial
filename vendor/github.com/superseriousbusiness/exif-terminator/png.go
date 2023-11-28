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
	"io"

	pngstructure "github.com/superseriousbusiness/go-png-image-structure/v2"
)

type pngVisitor struct {
	ps               *pngstructure.PngSplitter
	writer           io.Writer
	lastWrittenChunk int
}

func (v *pngVisitor) split(data []byte, atEOF bool) (int, []byte, error) {
	// execute the ps split function to read in data
	advance, token, err := v.ps.Split(data, atEOF)
	if err != nil {
		return advance, token, err
	}

	// if we haven't written anything at all yet, then write the png header back into the writer first
	if v.lastWrittenChunk == -1 {
		if _, err := v.writer.Write(pngstructure.PngSignature[:]); err != nil {
			return advance, token, err
		}
	}

	// Check if the splitter now has
	// any new chunks in it for us.
	chunkSlice, err := v.ps.Chunks()
	if err != nil {
		return advance, token, err
	}

	// Write each chunk by passing it
	// through our custom write func,
	// which strips out exif and fixes
	// the CRC of each chunk.
	chunks := chunkSlice.Chunks()
	for i, chunk := range chunks {
		if i <= v.lastWrittenChunk {
			// Skip already
			// written chunks.
			continue
		}

		// Write this new chunk.
		if err := v.writeChunk(chunk); err != nil {
			return advance, token, err
		}
		v.lastWrittenChunk = i

		// Zero data; here you
		// go garbage collector.
		chunk.Data = nil
	}

	return advance, token, err
}

func (v *pngVisitor) writeChunk(chunk *pngstructure.Chunk) error {
	if err := binary.Write(v.writer, binary.BigEndian, chunk.Length); err != nil {
		return err
	}

	if _, err := v.writer.Write([]byte(chunk.Type)); err != nil {
		return err
	}

	if chunk.Type == pngstructure.EXifChunkType {
		// Replace exif data with zero bytes.
		zeros := make([]byte, len(chunk.Data))
		chunk.Data = zeros
	}

	if _, err := v.writer.Write(chunk.Data); err != nil {
		return err
	}

	// Fix CRC of each chunk.
	chunk.UpdateCrc32()
	if err := binary.Write(v.writer, binary.BigEndian, chunk.Crc); err != nil {
		return err
	}

	return nil
}

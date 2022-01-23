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

	pngstructure "github.com/dsoprea/go-png-image-structure/v2"
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

	// check if the splitter has any new chunks in it that we haven't written yet
	chunkSlice := v.ps.Chunks()
	chunks := chunkSlice.Chunks()
	for i, chunk := range chunks {
		// look through all the chunks in the splitter
		if i > v.lastWrittenChunk {
			// we've got a chunk we haven't written yet! write it...
			if err := v.writeChunk(chunk); err != nil {
				return advance, token, err
			}
			// then remove the data
			chunk.Data = chunk.Data[:0]
			// and update
			v.lastWrittenChunk = i
		}
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
		blank := make([]byte, len(chunk.Data))
		if _, err := v.writer.Write(blank); err != nil {
			return err
		}
	} else {
		if _, err := v.writer.Write(chunk.Data); err != nil {
			return err
		}
	}

	if err := binary.Write(v.writer, binary.BigEndian, chunk.Crc); err != nil {
		return err
	}

	return nil
}

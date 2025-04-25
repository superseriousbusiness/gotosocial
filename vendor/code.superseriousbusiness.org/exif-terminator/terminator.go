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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	jpegstructure "code.superseriousbusiness.org/go-jpeg-image-structure/v2"
	pngstructure "code.superseriousbusiness.org/go-png-image-structure/v2"
)

func Terminate(in io.Reader, mediaType string) (io.Reader, error) {
	// To avoid keeping too much stuff
	// in memory we want to pipe data
	// directly to the reader.
	pr, pw := io.Pipe()

	// Setup scanner to terminate exif into pipe writer.
	scanner, err := terminatingScanner(pw, in, mediaType)
	if err != nil {
		_ = pw.Close()
		return nil, err
	}

	go func() {
		var err error

		defer func() {
			// Always close writer, using returned
			// scanner error (if any). If err is nil
			// then the standard io.EOF will be used.
			// (this will not overwrite existing).
			pw.CloseWithError(err)
		}()

		// Scan through input.
		for scanner.Scan() {
		}

		// Set error on return.
		err = scanner.Err()
	}()

	return pr, nil
}

func TerminateInto(out io.Writer, in io.Reader, mediaType string) error {
	// Setup scanner to terminate exif from 'in' to 'out'.
	scanner, err := terminatingScanner(out, in, mediaType)
	if err != nil {
		return err
	}

	// Scan through input.
	for scanner.Scan() {
	}

	// Return scan errors.
	return scanner.Err()
}

func terminatingScanner(out io.Writer, in io.Reader, mediaType string) (*bufio.Scanner, error) {
	scanner := bufio.NewScanner(in)

	// 40mb buffer size should be enough
	// to scan through most file chunks
	// without running into issues, they're
	// usually chunked smaller than this...
	scanner.Buffer(nil, 40*1024*1024)

	switch mediaType {
	case "image/jpeg", "jpeg", "jpg":
		v := &jpegVisitor{
			writer: out,
		}

		// Provide the visitor to the splitter so
		// that it triggers on every section scan.
		js := jpegstructure.NewJpegSplitter(v)

		// The visitor also needs to read back the
		// list of segments: for this it needs to
		// know what jpeg splitter it's attached to,
		// so give it a pointer to the splitter.
		v.js = js

		// Jpeg visitor's 'split' function
		// satisfies bufio.SplitFunc{}.
		scanner.Split(js.Split)

	case "image/webp", "webp":
		// Webp visitor's 'split' function
		// satisfies bufio.SplitFunc{}.
		scanner.Split((&webpVisitor{
			writer: out,
		}).split)

	case "image/png", "png":
		// For pngs we need to skip the header bytes, so read
		// them in and check we're really dealing with a png.
		header := make([]byte, len(pngstructure.PngSignature))
		if _, headerError := in.Read(header); headerError != nil {
			return nil, headerError
		} else if !bytes.Equal(header, pngstructure.PngSignature[:]) {
			return nil, errors.New("could not decode png: invalid header")
		}

		// Don't bother checking CRC;
		// we're overwriting it anyway.
		ps := pngstructure.NewPngSplitter()
		ps.DoCheckCrc(false)

		// Png visitor's 'split' function
		// satisfies bufio.SplitFunc{}.
		scanner.Split((&pngVisitor{
			ps:               ps,
			writer:           out,
			lastWrittenChunk: -1,
		}).split)

	default:
		return nil, fmt.Errorf("mediaType %s cannot be processed", mediaType)
	}

	return scanner, nil
}

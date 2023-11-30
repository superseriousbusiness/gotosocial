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

	jpegstructure "github.com/superseriousbusiness/go-jpeg-image-structure/v2"
	pngstructure "github.com/superseriousbusiness/go-png-image-structure/v2"
)

func Terminate(in io.Reader, fileSize int, mediaType string) (io.Reader, error) {
	// To avoid keeping too much stuff
	// in memory we want to pipe data
	// directly to the reader.
	pipeReader, pipeWriter := io.Pipe()

	// We don't know ahead of time how long
	// segments might be: they could be as
	// large as the file itself, so we need
	// a buffer with generous overhead.
	scanner := bufio.NewScanner(in)
	scanner.Buffer([]byte{}, fileSize)

	var err error
	switch mediaType {
	case "image/jpeg", "jpeg", "jpg":
		err = terminateJpeg(scanner, pipeWriter, fileSize)

	case "image/webp", "webp":
		err = terminateWebp(scanner, pipeWriter)

	case "image/png", "png":
		// For pngs we need to skip the header bytes, so read
		// them in and check we're really dealing with a png.
		header := make([]byte, len(pngstructure.PngSignature))
		if _, headerError := in.Read(header); headerError != nil {
			err = headerError
			break
		}

		if !bytes.Equal(header, pngstructure.PngSignature[:]) {
			err = errors.New("could not decode png: invalid header")
			break
		}

		err = terminatePng(scanner, pipeWriter)
	default:
		err = fmt.Errorf("mediaType %s cannot be processed", mediaType)
	}

	return pipeReader, err
}

func terminateJpeg(scanner *bufio.Scanner, writer *io.PipeWriter, expectedFileSize int) error {
	v := &jpegVisitor{
		writer:           writer,
		expectedFileSize: expectedFileSize,
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

	go scanAndClose(scanner, writer)
	return nil
}

func terminateWebp(scanner *bufio.Scanner, writer *io.PipeWriter) error {
	v := &webpVisitor{
		writer: writer,
	}

	// Webp visitor's 'split' function
	// satisfies bufio.SplitFunc{}.
	scanner.Split(v.split)

	go scanAndClose(scanner, writer)
	return nil
}

func terminatePng(scanner *bufio.Scanner, writer *io.PipeWriter) error {
	ps := pngstructure.NewPngSplitter()

	// Don't bother checking CRC;
	// we're overwriting it anyway.
	ps.DoCheckCrc(false)

	v := &pngVisitor{
		ps:               ps,
		writer:           writer,
		lastWrittenChunk: -1,
	}

	// Png visitor's 'split' function
	// satisfies bufio.SplitFunc{}.
	scanner.Split(v.split)

	go scanAndClose(scanner, writer)
	return nil
}

// scanAndClose scans through the given scanner until there's
// nothing left to scan, and then closes the writer so that the
// reader on the other side of the pipe knows that we're done.
//
// Any error encountered when scanning will be logged by terminator.
//
// Due to the nature of io.Pipe, writing won't actually work
// until the pipeReader starts being read by the caller, which
// is why this function should always be called asynchronously.
func scanAndClose(scanner *bufio.Scanner, writer *io.PipeWriter) {
	var err error

	defer func() {
		// Always close writer, using returned
		// scanner error (if any). If err is nil
		// then the standard io.EOF will be used.
		// (this will not overwrite existing).
		writer.CloseWithError(err)
	}()

	for scanner.Scan() {
	}

	// Set error on return.
	err = scanner.Err()
}

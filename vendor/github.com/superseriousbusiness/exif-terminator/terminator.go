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

	pngstructure "github.com/dsoprea/go-png-image-structure/v2"
	jpegstructure "github.com/superseriousbusiness/go-jpeg-image-structure/v2"
)

func Terminate(in io.Reader, fileSize int, mediaType string) (io.Reader, error) {
	// to avoid keeping too much stuff in memory we want to pipe data directly
	pipeReader, pipeWriter := io.Pipe()

	// we don't know ahead of time how long segments might be: they could be as large as
	// the file itself, so unfortunately we need to allocate a buffer here that'scanner as large
	// as the file
	scanner := bufio.NewScanner(in)
	scanner.Buffer([]byte{}, fileSize)
	var err error

	switch mediaType {
	case "image/jpeg", "jpeg", "jpg":
		err = terminateJpeg(scanner, pipeWriter, fileSize)
	case "image/webp", "webp":
		err = terminateWebp(scanner, pipeWriter)
	case "image/png", "png":
		// for pngs we need to skip the header bytes, so read them in
		// and check we're really dealing with a png here
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

func terminateJpeg(scanner *bufio.Scanner, writer io.WriteCloser, expectedFileSize int) error {
	// jpeg visitor is where the spicy hack of streaming the de-exifed data is contained
	v := &jpegVisitor{
		writer:           writer,
		expectedFileSize: expectedFileSize,
	}

	// provide the visitor to the splitter so that it triggers on every section scan
	js := jpegstructure.NewJpegSplitter(v)

	// the visitor also needs to read back the list of segments: for this it needs
	// to know what jpeg splitter it's attached to, so give it a pointer to the splitter
	v.js = js

	// use the jpeg splitters 'split' function, which satisfies the bufio.SplitFunc interface
	scanner.Split(js.Split)

	scanAndClose(scanner, writer)
	return nil
}

func terminateWebp(scanner *bufio.Scanner, writer io.WriteCloser) error {
	v := &webpVisitor{
		writer: writer,
	}

	// use the webp visitor's 'split' function, which satisfies the bufio.SplitFunc interface
	scanner.Split(v.split)

	scanAndClose(scanner, writer)
	return nil
}

func terminatePng(scanner *bufio.Scanner, writer io.WriteCloser) error {
	ps := pngstructure.NewPngSplitter()

	v := &pngVisitor{
		ps:               ps,
		writer:           writer,
		lastWrittenChunk: -1,
	}

	// use the png visitor's 'split' function, which satisfies the bufio.SplitFunc interface
	scanner.Split(v.split)

	scanAndClose(scanner, writer)
	return nil
}

func scanAndClose(scanner *bufio.Scanner, writer io.WriteCloser) {
	// scan asynchronously until there's nothing left to scan, and then close the writer
	// so that the reader on the other side knows that we're done
	//
	// due to the nature of io.Pipe, writing won't actually work
	// until the pipeReader starts being read by the caller, which
	// is why we do this asynchronously
	go func() {
		defer writer.Close()
		for scanner.Scan() {
		}
		if scanner.Err() != nil {
			logger.Error(scanner.Err())
		}
	}()
}

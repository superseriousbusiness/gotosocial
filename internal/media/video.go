/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package media

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"os"

	"github.com/abema/go-mp4"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

var thumbFill = color.RGBA{42, 43, 47, 0} // the color to fill video thumbnails with

func decodeVideo(r io.Reader, contentType string) (*mediaMeta, error) {
	// We'll need a readseeker to decode the video. We can get a readseeker
	// without burning too much mem by first copying the reader into a temp file.
	// First create the file in the temporary directory...
	tempFile, err := os.CreateTemp(os.TempDir(), "gotosocial-")
	if err != nil {
		return nil, fmt.Errorf("could not create temporary file while decoding video: %w", err)
	}
	tempFileName := tempFile.Name()

	// Make sure to clean up the temporary file when we're done with it
	defer func() {
		if err := tempFile.Close(); err != nil {
			log.Errorf("could not close file %s: %s", tempFileName, err)
		}
		if err := os.Remove(tempFileName); err != nil {
			log.Errorf("could not remove file %s: %s", tempFileName, err)
		}
	}()

	// Now copy the entire reader we've been provided into the
	// temporary file; we won't use the reader again after this.
	if _, err := io.Copy(tempFile, r); err != nil {
		return nil, fmt.Errorf("could not copy video reader into temporary file %s: %w", tempFileName, err)
	}

	// define some vars we need to pull the width/height out of the video
	var (
		height      int
		width       int
		readHandler = getReadHandler(&height, &width)
	)

	// do the actual decoding here, providing the temporary file we created as readseeker
	if _, err := mp4.ReadBoxStructure(tempFile, readHandler); err != nil {
		return nil, fmt.Errorf("parsing video data: %w", err)
	}

	// width + height should now be updated by the readHandler
	return &mediaMeta{
		width:  width,
		height: height,
		size:   height * width,
		aspect: float64(width) / float64(height),
	}, nil
}

// getReadHandler returns a handler function that updates the underling
// values of the given height and width int pointers to the hightest and
// widest points of the video.
func getReadHandler(height *int, width *int) func(h *mp4.ReadHandle) (interface{}, error) {
	return func(rh *mp4.ReadHandle) (interface{}, error) {
		if rh.BoxInfo.Type == mp4.BoxTypeTkhd() {
			box, _, err := rh.ReadPayload()
			if err != nil {
				return nil, fmt.Errorf("could not read mp4 payload: %w", err)
			}

			tkhd, ok := box.(*mp4.Tkhd)
			if !ok {
				return nil, errors.New("box was not of type *mp4.Tkhd")
			}

			// if height + width of this box are greater than what
			// we have stored, then update our stored values
			if h := int(tkhd.GetHeight()); h > *height {
				*height = h
			}

			if w := int(tkhd.GetWidth()); w > *width {
				*width = w
			}
		}

		if rh.BoxInfo.IsSupportedType() {
			return rh.Expand()
		}

		return nil, nil
	}
}

func deriveThumbnailFromVideo(height int, width int) (*mediaMeta, error) {
	// create a rectangle with the same dimensions as the video
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// fill the rectangle with our desired fill color
	draw.Draw(img, img.Bounds(), &image.Uniform{thumbFill}, image.Point{}, draw.Src)

	// we can get away with using extremely poor quality for this monocolor thumbnail
	out := &bytes.Buffer{}
	if err := jpeg.Encode(out, img, &jpeg.Options{Quality: 1}); err != nil {
		return nil, fmt.Errorf("error encoding video thumbnail: %w", err)
	}

	return &mediaMeta{
		width:  width,
		height: height,
		size:   width * height,
		aspect: float64(width) / float64(height),
		small:  out.Bytes(),
	}, nil
}

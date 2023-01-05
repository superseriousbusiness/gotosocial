/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"os"

	"github.com/abema/go-mp4"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
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

	var (
		width     int
		height    int
		duration  float32
		framerate float32
		bitrate   uint64
	)

	// probe the video file to extract useful metadata from it; for methodology, see:
	// https://github.com/abema/go-mp4/blob/7d8e5a7c5e644e0394261b0cf72fef79ce246d31/mp4tool/probe/probe.go#L85-L154
	info, err := mp4.Probe(tempFile)
	if err != nil {
		return nil, fmt.Errorf("could not probe temporary video file %s: %w", tempFileName, err)
	}

	for _, tr := range info.Tracks {
		if tr.AVC == nil {
			continue
		}

		if w := int(tr.AVC.Width); w > width {
			width = w
		}

		if h := int(tr.AVC.Height); h > height {
			height = h
		}

		if br := tr.Samples.GetBitrate(tr.Timescale); br > bitrate {
			bitrate = br
		} else if br := info.Segments.GetBitrate(tr.TrackID, tr.Timescale); br > bitrate {
			bitrate = br
		}

		if d := float32(tr.Duration) / float32(tr.Timescale); d > duration {
			duration = d
			framerate = float32(len(tr.Samples)) / duration
		}
	}

	var errs gtserror.MultiError
	if width == 0 {
		errs = append(errs, "video width could not be discovered")
	}

	if height == 0 {
		errs = append(errs, "video height could not be discovered")
	}

	if duration == 0 {
		errs = append(errs, "video duration could not be discovered")
	}

	if framerate == 0 {
		errs = append(errs, "video framerate could not be discovered")
	}

	if bitrate == 0 {
		errs = append(errs, "video bitrate could not be discovered")
	}

	if errs != nil {
		return nil, errs.Combine()
	}

	return &mediaMeta{
		width:     width,
		height:    height,
		duration:  duration,
		framerate: framerate,
		bitrate:   bitrate,
		size:      height * width,
		aspect:    float32(width) / float32(height),
	}, nil
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
		aspect: float32(width) / float32(height),
		small:  out.Bytes(),
	}, nil
}

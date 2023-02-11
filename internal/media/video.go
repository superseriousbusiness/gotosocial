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
	"fmt"
	"io"

	"github.com/abema/go-mp4"
	"github.com/superseriousbusiness/gotosocial/internal/iotools"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type gtsVideo struct {
	frame     *gtsImage
	duration  float32 // in seconds
	bitrate   uint64
	framerate float32
}

// decodeVideoFrame decodes and returns an image from a single frame in the given video stream.
// (note: currently this only returns a blank image resized to fit video dimensions).
func decodeVideoFrame(r io.Reader) (*gtsVideo, error) {
	// we need a readseeker to decode the video...
	tfs, err := iotools.TempFileSeeker(r)
	if err != nil {
		return nil, fmt.Errorf("error creating temp file seeker: %w", err)
	}
	defer func() {
		if err := tfs.Close(); err != nil {
			log.Errorf(nil, "error closing temp file seeker: %s", err)
		}
	}()

	// probe the video file to extract useful metadata from it; for methodology, see:
	// https://github.com/abema/go-mp4/blob/7d8e5a7c5e644e0394261b0cf72fef79ce246d31/mp4tool/probe/probe.go#L85-L154
	info, err := mp4.Probe(tfs)
	if err != nil {
		return nil, fmt.Errorf("error during mp4 probe: %w", err)
	}

	var (
		width        int
		height       int
		videoBitrate uint64
		audioBitrate uint64
		video        gtsVideo
	)

	for _, tr := range info.Tracks {
		if tr.AVC == nil {
			// audio track
			if br := tr.Samples.GetBitrate(tr.Timescale); br > audioBitrate {
				audioBitrate = br
			} else if br := info.Segments.GetBitrate(tr.TrackID, tr.Timescale); br > audioBitrate {
				audioBitrate = br
			}

			if d := float64(tr.Duration) / float64(tr.Timescale); d > float64(video.duration) {
				video.duration = float32(d)
			}
			continue
		}

		// video track
		if w := int(tr.AVC.Width); w > width {
			width = w
		}

		if h := int(tr.AVC.Height); h > height {
			height = h
		}

		if br := tr.Samples.GetBitrate(tr.Timescale); br > videoBitrate {
			videoBitrate = br
		} else if br := info.Segments.GetBitrate(tr.TrackID, tr.Timescale); br > videoBitrate {
			videoBitrate = br
		}

		if d := float64(tr.Duration) / float64(tr.Timescale); d > float64(video.duration) {
			video.framerate = float32(len(tr.Samples)) / float32(d)
			video.duration = float32(d)
		}
	}

	// overall bitrate should be audio + video combined
	// (since they're both playing at the same time)
	video.bitrate = audioBitrate + videoBitrate

	// Check for empty video metadata.
	var empty []string
	if width == 0 {
		empty = append(empty, "width")
	}
	if height == 0 {
		empty = append(empty, "height")
	}
	if video.duration == 0 {
		empty = append(empty, "duration")
	}
	if video.framerate == 0 {
		empty = append(empty, "framerate")
	}
	if video.bitrate == 0 {
		empty = append(empty, "bitrate")
	}
	if len(empty) > 0 {
		return nil, fmt.Errorf("error determining video metadata: %v", empty)
	}

	// Create new empty "frame" image.
	// TODO: decode frame from video file.
	video.frame = blankImage(width, height)

	return &video, nil
}

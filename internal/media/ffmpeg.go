// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package media

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path"
	"strconv"
	"strings"

	"codeberg.org/gruf/go-byteutil"

	"codeberg.org/gruf/go-ffmpreg/wasm"
	_ffmpeg "github.com/superseriousbusiness/gotosocial/internal/media/ffmpeg"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/tetratelabs/wazero"
)

// ffmpegClearMetadata generates a copy (in-place) of input media with all metadata cleared.
func ffmpegClearMetadata(ctx context.Context, filepath string, ext string) error {
	// Get directory from filepath.
	dirpath := path.Dir(filepath)

	// Generate output file path with ext.
	outpath := filepath + "_cleaned." + ext

	// Clear metadata with ffmpeg.
	if err := ffmpeg(ctx, dirpath,
		"-loglevel", "error",
		"-i", filepath,
		"-map_metadata", "-1",
		"-codec", "copy",
		"-y",
		outpath,
	); err != nil {
		return err
	}

	// Move the new output file path to original location.
	if err := os.Rename(outpath, filepath); err != nil {
		return gtserror.Newf("error renaming %s: %w", outpath, err)
	}

	return nil
}

// ffmpegGenerateThumb generates a thumbnail jpeg from input media of any type, useful for any media.
func ffmpegGenerateThumb(ctx context.Context, filepath string, width, height int) (string, error) {
	// Get directory from filepath.
	dirpath := path.Dir(filepath)

	// Generate output frame file path.
	outpath := filepath + "_thumb.jpg"

	// Generate thumb with ffmpeg.
	if err := ffmpeg(ctx, dirpath,
		"-loglevel", "error",
		"-i", filepath,
		"-filter:v", "thumbnail=n=10",
		"-filter:v", "scale="+strconv.Itoa(width)+":"+strconv.Itoa(height),
		"-qscale:v", "12", // ~ 70% quality
		"-frames:v", "1",
		"-y",
		outpath,
	); err != nil {
		return "", err
	}

	return outpath, nil
}

// ffmpegGenerateStatic generates a static png from input image of any type, useful for emoji.
func ffmpegGenerateStatic(ctx context.Context, filepath string) (string, error) {
	// Get directory from filepath.
	dirpath := path.Dir(filepath)

	// Generate output static file path.
	outpath := filepath + "_static.png"

	// Generate static with ffmpeg.
	if err := ffmpeg(ctx, dirpath,
		"-loglevel", "error",
		"-i", filepath,
		"-codec:v", "png", // specifically NOT 'apng'
		"-frames:v", "1", // in case animated, only take 1 frame
		"-y",
		outpath,
	); err != nil {
		return "", err
	}

	return outpath, nil
}

// ffmpeg calls `ffmpeg [args...]` (WASM) with directory path mounted in runtime.
func ffmpeg(ctx context.Context, dirpath string, args ...string) error {
	var stderr byteutil.Buffer
	rc, err := _ffmpeg.Ffmpeg(ctx, wasm.Args{
		Stderr: &stderr,
		Args:   args,
		Config: func(modcfg wazero.ModuleConfig) wazero.ModuleConfig {
			fscfg := wazero.NewFSConfig()
			fscfg = fscfg.WithDirMount(dirpath, dirpath)
			modcfg = modcfg.WithFSConfig(fscfg)
			return modcfg
		},
	})
	if err != nil {
		return gtserror.Newf("error running: %w", err)
	} else if rc != 0 {
		return gtserror.Newf("non-zero return code %d (%s)", rc, stderr.B)
	}
	return nil
}

// ffprobe calls `ffprobe` (WASM) on filepath, returning parsed JSON output.
func ffprobe(ctx context.Context, filepath string) (*result, error) {
	var stdout byteutil.Buffer

	// Get directory from filepath.
	dirpath := path.Dir(filepath)

	// Run ffprobe on our given file at path.
	_, err := _ffmpeg.Ffprobe(ctx, wasm.Args{
		Stdout: &stdout,

		Args: []string{
			"-i", filepath,
			"-loglevel", "quiet",
			"-print_format", "json",
			"-show_streams",
			"-show_format",
			"-show_error",
		},

		Config: func(modcfg wazero.ModuleConfig) wazero.ModuleConfig {
			fscfg := wazero.NewFSConfig()
			fscfg = fscfg.WithReadOnlyDirMount(dirpath, dirpath)
			modcfg = modcfg.WithFSConfig(fscfg)
			return modcfg
		},
	})
	if err != nil {
		return nil, gtserror.Newf("error running: %w", err)
	}

	var result ffprobeResult

	// Unmarshal the ffprobe output as our result type.
	if err := json.Unmarshal(stdout.B, &result); err != nil {
		return nil, gtserror.Newf("error unmarshaling json: %w", err)
	}

	// Convert raw result data.
	res, err := result.Process()
	if err != nil {
		return nil, err
	}

	return res, nil
}

// result contains parsed ffprobe result
// data in a more useful data format.
type result struct {
	format   string
	audio    []audioStream
	video    []videoStream
	bitrate  uint64
	duration float64
}

type stream struct {
	codec string
}

type audioStream struct {
	stream
}

type videoStream struct {
	stream
	width     int
	height    int
	framerate float32
}

// GetFileType determines file type and extension to use for media data. This
// function helps to abstract away the horrible complexities that are possible
// general media container types and and possible sub-types within that.
func (res *result) GetFileType() (gtsmodel.FileType, string) {
	switch res.format {
	case "mpeg":
		return gtsmodel.FileTypeVideo, "mpeg"
	case "mov,mp4,m4a,3gp,3g2,mj2":
		for _, stream := range res.video {
			// check for motion video
			// (static is album art).
			if stream.framerate > 0 {
				return gtsmodel.FileTypeVideo, "mp4"
			}
		}
		for _, stream := range res.audio {
			switch stream.codec { //nolint
			case "aac":
				return gtsmodel.FileTypeAudio, "m4a"
			}
		}
	case "apng":
		return gtsmodel.FileTypeImage, "apng"
	case "png_pipe":
		return gtsmodel.FileTypeImage, "png"
	case "image2", "jpeg_pipe":
		return gtsmodel.FileTypeImage, "jpeg"
	case "webp_pipe":
		return gtsmodel.FileTypeImage, "webp"
	case "gif":
		return gtsmodel.FileTypeImage, "gif"
	case "mp3":
		for _, stream := range res.audio {
			switch stream.codec {
			case "mp2":
				return gtsmodel.FileTypeAudio, "mp2"
			case "mp3":
				return gtsmodel.FileTypeAudio, "mp3"
			}
		}
	case "asf":
		for _, stream := range res.video {
			// check for motion video
			// (static is album art).
			if stream.framerate > 0 {
				return gtsmodel.FileTypeVideo, "wmv"
			}
		}
		switch { //nolint
		case len(res.audio) > 0:
			return gtsmodel.FileTypeAudio, "wma"
		}
	case "ogg":
		for _, stream := range res.video {
			// check for motion video
			// (static is album art).
			if stream.framerate > 0 {
				return gtsmodel.FileTypeVideo, "ogv"
			}
		}
		switch { //nolint
		case len(res.audio) > 0:
			return gtsmodel.FileTypeAudio, "ogg"
		}
	case "matroska,webm":
		for _, stream := range res.video {
			// check for motion video
			// (static is album art).
			if stream.framerate > 0 {
				return gtsmodel.FileTypeVideo, "mkv"
			}
		}
		switch { //nolint
		case len(res.audio) > 0:
			return gtsmodel.FileTypeAudio, "mka"
		}
	case "avi":
		return gtsmodel.FileTypeVideo, "avi"
	case "mpegts":
		return gtsmodel.FileTypeVideo, "mts"
	}
	return gtsmodel.FileTypeUnknown, res.format
}

// ImageMeta extracts image metadata contained within ffprobe'd media result streams.
func (res *result) ImageMeta() (width int, height int, err error) {
	for _, stream := range res.video {
		if stream.width > width {
			width = stream.width
		}
		if stream.height > height {
			height = stream.height
		}
	}
	if width == 0 || height == 0 {
		err = errors.New("invalid image stream(s)")
	}
	return
}

// VideoMeta extracts video metadata contained within ffprobe'd media result streams.
func (res *result) VideoMeta() (width, height int, framerate float32, err error) {
	for _, stream := range res.video {
		if stream.width > width {
			width = stream.width
		}
		if stream.height > height {
			height = stream.height
		}
		if fr := float32(stream.framerate); fr > 0 {
			if framerate == 0 || fr < framerate {
				framerate = fr
			}
		}
	}
	if width == 0 || height == 0 || framerate == 0 {
		err = errors.New("invalid video stream(s)")
	}
	return
}

// Process converts raw ffprobe result data into our more usable result{} type.
func (res *ffprobeResult) Process() (*result, error) {
	if res.Error != nil {
		return nil, res.Error
	}

	if res.Format == nil {
		return nil, errors.New("missing format data")
	}

	var r result
	var err error

	// Copy over container format.
	r.format = res.Format.FormatName

	// Parsed media bitrate (if it was set).
	if str := res.Format.BitRate; str != "" {
		r.bitrate, err = strconv.ParseUint(str, 10, 64)
		if err != nil {
			return nil, gtserror.Newf("invalid bitrate %s: %w", str, err)
		}
	}

	// Parse media duration (if it was set).
	if str := res.Format.Duration; str != "" {
		r.duration, err = strconv.ParseFloat(str, 32)
		if err != nil {
			return nil, gtserror.Newf("invalid duration %s: %w", str, err)
		}
	}

	// Preallocate streams to max possible lengths.
	r.audio = make([]audioStream, 0, len(res.Streams))
	r.video = make([]videoStream, 0, len(res.Streams))

	// Convert streams to separate types.
	for _, s := range res.Streams {
		switch s.CodecType {
		case "audio":
			// Append audio stream data to result.
			r.audio = append(r.audio, audioStream{
				stream: stream{codec: s.CodecName},
			})
		case "video":
			var framerate float32

			// Parse stream framerate (if non-zero).
			if str := s.AvgFrameRate; str != "" {
				var num, den float32

				// Check for inequality (numerator / denominator).
				if p := strings.SplitN(str, "/", 2); len(p) == 2 {
					n, _ := strconv.ParseFloat(p[0], 32)
					d, _ := strconv.ParseFloat(p[1], 32)
					num, den = float32(n), float32(d)
				} else {
					n, _ := strconv.ParseFloat(p[0], 32)
					num = float32(n)
				}

				// Set final framerate.
				framerate = (num / den)
			}

			// Append video stream data to result.
			r.video = append(r.video, videoStream{
				stream:    stream{codec: s.CodecName},
				width:     s.Width,
				height:    s.Height,
				framerate: framerate,
			})
		}
	}

	return &r, nil
}

// ffprobeResult contains parsed JSON data from
// result of calling `ffprobe` on a media file.
type ffprobeResult struct {

	// ffprobe JSON data
	Streams []ffprobeStream `json:"streams"`
	Format  *ffprobeFormat  `json:"format"`
	Error   *ffprobeError   `json:"error"`

	// further processed details
	Audio []*ffprobeStream `json:"-"`
	Video []*ffprobeStream `json:"-"`
}

type ffprobeStream struct {
	CodecName    string `json:"codec_name"`
	CodecType    string `json:"codec_type"`
	AvgFrameRate string `json:"avg_frame_rate"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	// + unused fields.
}

type ffprobeFormat struct {
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"`
	BitRate    string `json:"bit_rate"`
	// + unused fields
}

type ffprobeError struct {
	Code   int    `json:"code"`
	String string `json:"string"`
}

func isUnsupportedTypeErr(err error) bool {
	ffprobeErr, ok := err.(*ffprobeError)
	return ok && ffprobeErr.Code == -1094995529
}

func (err *ffprobeError) Error() string {
	return err.String + " (" + strconv.Itoa(err.Code) + ")"
}

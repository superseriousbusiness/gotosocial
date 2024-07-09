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
	ffmpeglib "codeberg.org/gruf/go-ffmpreg/ffmpeg"
	ffprobelib "codeberg.org/gruf/go-ffmpreg/ffprobe"
	"codeberg.org/gruf/go-ffmpreg/wasm"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/tetratelabs/wazero"
)

// ffmpegClearMetadata ...
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

// ffmpegGenerateThumb ...
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
		outpath,
	); err != nil {
		return "", err
	}

	return outpath, nil
}

// ffmpegGenerateStatic ...
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
		outpath,
	); err != nil {
		return "", gtserror.Newf("%w (filepath=%q outpath=%q)", err, filepath, outpath)
	}

	return outpath, nil
}

// ffmpeg calls `ffmpeg [args...]` (WASM) with directory path mounted in runtime.
func ffmpeg(ctx context.Context, dirpath string, args ...string) error {
	var stderr byteutil.Buffer
	rc, err := ffmpeglib.Run(ctx, wasm.Args{
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
func ffprobe(ctx context.Context, filepath string) (*ffprobeResult, error) {
	var stdout byteutil.Buffer

	// Get directory from filepath.
	dirpath := path.Dir(filepath)

	// Run ffprobe on our given file at path.
	rc, err := ffprobelib.Run(ctx, wasm.Args{
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
	} else if rc != 0 {
		return nil, gtserror.Newf("non-zero return code %d", rc)
	}

	var result ffprobeResult

	// Unmarshal the ffprobe output as our result type.
	if err := json.Unmarshal(stdout.B, &result); err != nil {
		return nil, gtserror.Newf("error unmarshaling json: %w", err)
	}

	// Check for ffprobe result error.
	if err := result.Error; err != nil {
		return nil, err
	}

	// Ensure valid result data.
	if len(result.Streams) == 0 ||
		result.Format == nil {
		return nil, gtserror.Newf("invalid result data: %s", stdout.B)
	}

	return &result, nil
}

// ffprobeResult contains parsed JSON data from
// result of calling `ffprobe` on a media file.
type ffprobeResult struct {
	Streams []ffprobeStream `json:"streams"`
	Format  *ffprobeFormat  `json:"format"`
	Error   *ffprobeError   `json:"error"`
}

// ImageMeta extracts image metadata contained within ffprobe'd media result streams.
func (res *ffprobeResult) ImageMeta() (width int, height int, err error) {
	for _, stream := range res.Streams {
		if stream.Width > width {
			width = stream.Width
		}
		if stream.Height > height {
			height = stream.Height
		}
	}
	if width == 0 || height == 0 {
		err = errors.New("invalid image stream(s)")
	}
	return
}

// VideoMeta extracts video metadata contained within ffprobe'd media result streams.
func (res *ffprobeResult) VideoMeta() (width, height int, framerate float32, err error) {
	for _, stream := range res.Streams {
		if stream.Width > width {
			width = stream.Width
		}
		if stream.Height > height {
			height = stream.Height
		}
		if fr := stream.GetFrameRate(); fr > 0 {
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

type ffprobeStream struct {
	CodecName    string `json:"codec_name"`
	AvgFrameRate string `json:"avg_frame_rate"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	// + unused fields.
}

// GetFrameRate calculates float32 framerate value from stream json string.
func (str *ffprobeStream) GetFrameRate() float32 {
	if str.AvgFrameRate != "" {
		var (
			// numerator
			num float32

			// denominator
			den float32
		)

		// Check for a provided inequality, i.e. numerator / denominator.
		if p := strings.SplitN(str.AvgFrameRate, "/", 2); len(p) == 2 {
			n, _ := strconv.ParseFloat(p[0], 32)
			d, _ := strconv.ParseFloat(p[1], 32)
			num, den = float32(n), float32(d)
		} else {
			n, _ := strconv.ParseFloat(p[0], 32)
			num = float32(n)
		}

		return num / den
	}
	return 0
}

type ffprobeFormat struct {
	Filename   string `json:"filename"`
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"`
	BitRate    string `json:"bit_rate"`
	// + unused fields
}

// GetFileType determines file type and extension to use for media data.
func (fmt *ffprobeFormat) GetFileType() (gtsmodel.FileType, string) {
	switch fmt.FormatName {
	case "mov,mp4,m4a,3gp,3g2,mj2":
		return gtsmodel.FileTypeVideo, "mp4"
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
		return gtsmodel.FileTypeAudio, "mp3"
	case "ogg":
		return gtsmodel.FileTypeAudio, "ogg"
	default:
		return gtsmodel.FileTypeUnknown, fmt.FormatName
	}
}

// GetDuration calculates float32 framerate value from format json string.
func (fmt *ffprobeFormat) GetDuration() float32 {
	if fmt.Duration != "" {
		dur, _ := strconv.ParseFloat(fmt.Duration, 32)
		return float32(dur)
	}
	return 0
}

// GetBitRate calculates uint64 bitrate value from format json string.
func (fmt *ffprobeFormat) GetBitRate() uint64 {
	if fmt.BitRate != "" {
		r, _ := strconv.ParseUint(fmt.BitRate, 10, 64)
		return r
	}
	return 0
}

type ffprobeError struct {
	Code   int    `json:"code"`
	String string `json:"string"`
}

func (err *ffprobeError) Error() string {
	return err.String + " (" + strconv.Itoa(err.Code) + ")"
}

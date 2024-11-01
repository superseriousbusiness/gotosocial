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

	_ffmpeg "github.com/superseriousbusiness/gotosocial/internal/media/ffmpeg"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/tetratelabs/wazero"
)

// ffmpegClearMetadata generates a copy of input media with all metadata cleared.
// NOTE: given that we are not performing an encode, this only clears global level metadata,
// any metadata encoded into the media stream itself will not be cleared. This is the best we
// can do without absolutely tanking performance by requiring transcodes :(
func ffmpegClearMetadata(ctx context.Context, outpath, inpath string) error {
	return ffmpeg(ctx, inpath, outpath,

		// Only log errors.
		"-loglevel", "error",

		// Input file path.
		"-i", inpath,

		// Drop all metadata.
		"-map_metadata", "-1",

		// Copy input codecs,
		// i.e. no transcode.
		"-codec", "copy",

		// Overwrite.
		"-y",

		// Output.
		outpath,
	)
}

// ffmpegGenerateWebpThumb generates a thumbnail webp from input media of any type, useful for any media.
func ffmpegGenerateWebpThumb(ctx context.Context, inpath, outpath string, width, height int, pixfmt string) error {
	// Generate thumb with ffmpeg.
	return ffmpeg(ctx, inpath, outpath,

		// Only log errors.
		"-loglevel", "error",

		// Input file.
		"-i", inpath,

		// Encode using libwebp.
		// (NOT as libwebp_anim).
		"-codec:v", "libwebp",

		// Only one frame
		"-frames:v", "1",

		// Scale to dimensions
		// (scale filter: https://ffmpeg.org/ffmpeg-filters.html#scale)
		"-filter:v", "scale="+strconv.Itoa(width)+":"+strconv.Itoa(height)+","+

			// Attempt to use original pixel format
			// (format filter: https://ffmpeg.org/ffmpeg-filters.html#format)
			"format=pix_fmts="+pixfmt,

		// Quality not specified,
		// i.e. use default which
		// should be 75% webp quality.
		// (codec options: https://ffmpeg.org/ffmpeg-codecs.html#toc-Codec-Options)
		// (libwebp codec: https://ffmpeg.org/ffmpeg-codecs.html#Options-36)
		// "-qscale:v", "75",

		// Overwrite.
		"-y",

		// Output.
		outpath,
	)
}

// ffmpegGenerateStatic generates a static png from input image of any type, useful for emoji.
func ffmpegGenerateStatic(ctx context.Context, inpath string) (string, error) {
	var outpath string

	// Generate thumb output path REPLACING extension.
	if i := strings.IndexByte(inpath, '.'); i != -1 {
		outpath = inpath[:i] + "_static.png"
	} else {
		return "", gtserror.New("input file missing extension")
	}

	// Generate static with ffmpeg.
	if err := ffmpeg(ctx, inpath, outpath,

		// Only log errors.
		"-loglevel", "error",

		// Input file.
		"-i", inpath,

		// Only first frame.
		"-frames:v", "1",

		// Encode using png.
		// (NOT as apng).
		"-codec:v", "png",

		// Overwrite.
		"-y",

		// Output.
		outpath,
	); err != nil {
		return "", err
	}

	return outpath, nil
}

// ffmpeg calls `ffmpeg [args...]` (WASM) with in + out paths mounted in runtime.
func ffmpeg(ctx context.Context, inpath string, outpath string, args ...string) error {
	var stderr byteutil.Buffer
	rc, err := _ffmpeg.Ffmpeg(ctx, _ffmpeg.Args{
		Stderr: &stderr,
		Args:   args,
		Config: func(modcfg wazero.ModuleConfig) wazero.ModuleConfig {
			fscfg := wazero.NewFSConfig()

			// Needs read-only access to
			// /dev/urandom for some types.
			urandom := &allowFiles{
				{
					abs:  "/dev/urandom",
					flag: os.O_RDONLY,
					perm: 0,
				},
			}
			fscfg = fscfg.WithFSMount(urandom, "/dev")

			// In+out dirs are always the same (tmp),
			// so we can share one file system for
			// both + grant different perms to inpath
			// (read only) and outpath (read+write).
			shared := &allowFiles{
				{
					abs:  inpath,
					flag: os.O_RDONLY,
					perm: 0,
				},
				{
					abs:  outpath,
					flag: os.O_RDWR | os.O_CREATE | os.O_TRUNC,
					perm: 0666,
				},
			}
			fscfg = fscfg.WithFSMount(shared, path.Dir(inpath))

			return modcfg.WithFSConfig(fscfg)
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

	// Run ffprobe on our given file at path.
	_, err := _ffmpeg.Ffprobe(ctx, _ffmpeg.Args{
		Stdout: &stdout,

		Args: []string{
			// Don't show any excess logging
			// information, all goes in JSON.
			"-loglevel", "quiet",

			// Print in compact JSON format.
			"-print_format", "json=compact=1",

			// Show error in our
			// chosen format type.
			"-show_error",

			// Show specifically container format, total duration and bitrate.
			"-show_entries", "format=format_name,duration,bit_rate" + ":" +

				// Show specifically stream codec names, types, frame rate, duration, dimens, and pixel format.
				"stream=codec_name,codec_type,r_frame_rate,duration_ts,width,height,pix_fmt" + ":" +

				// Show orientation tag.
				"tags=orientation" + ":" +

				// Show rotation data.
				"side_data=rotation",

			// Limit to reading the first
			// 1s of data looking for "rotation"
			// side_data tags (expensive part).
			"-read_intervals", "%+1",

			// Input file.
			"-i", filepath,
		},

		Config: func(modcfg wazero.ModuleConfig) wazero.ModuleConfig {
			fscfg := wazero.NewFSConfig()

			// Needs read-only access
			// to file being probed.
			in := &allowFiles{
				{
					abs:  filepath,
					flag: os.O_RDONLY,
					perm: 0,
				},
			}
			fscfg = fscfg.WithFSMount(in, path.Dir(filepath))

			return modcfg.WithFSConfig(fscfg)
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

const (
	// possible orientation values
	// specified in "orientation"
	// tag of images.
	//
	// FlipH      = flips horizontally
	// FlipV      = flips vertically
	// Transpose  = flips horizontally and rotates 90 counter-clockwise.
	// Transverse = flips vertically and rotates 90 counter-clockwise.
	orientationUnspecified = 0
	orientationNormal      = 1
	orientationFlipH       = 2
	orientationRotate180   = 3
	orientationFlipV       = 4
	orientationTranspose   = 5
	orientationRotate270   = 6
	orientationTransverse  = 7
	orientationRotate90    = 8
)

// result contains parsed ffprobe result
// data in a more useful data format.
type result struct {
	format      string
	audio       []audioStream
	video       []videoStream
	duration    float64
	bitrate     uint64
	orientation int
}

type stream struct {
	codec string
}

type audioStream struct {
	stream
}

type videoStream struct {
	stream
	pixfmt    string
	width     int
	height    int
	framerate float32
}

// GetFileType determines file type and extension to use for media data. This
// function helps to abstract away the horrible complexities that are possible
// media container (i.e. the file) types and and possible sub-types within that.
//
// Note the checks for (len(res.video) > 0) may catch some audio files with embedded
// album art as video, but i blame that on the hellscape that is media filetypes.
func (res *result) GetFileType() (gtsmodel.FileType, string, string) {
	switch res.format {
	case "mpeg":
		return gtsmodel.FileTypeVideo,
			"video/mpeg", "mpeg"
	case "mjpeg":
		return gtsmodel.FileTypeVideo,
			"video/x-motion-jpeg", "mjpeg"
	case "mov,mp4,m4a,3gp,3g2,mj2":
		switch {
		case len(res.video) > 0:
			if len(res.audio) == 0 &&
				res.duration <= 30 {
				// Short, soundless
				// video file aka gifv.
				return gtsmodel.FileTypeGifv,
					"video/mp4", "mp4"
			} else {
				// Video file (with or without audio).
				return gtsmodel.FileTypeVideo,
					"video/mp4", "mp4"
			}
		case len(res.audio) > 0 &&
			res.audio[0].codec == "aac":
			// m4a only supports [aac] audio.
			return gtsmodel.FileTypeAudio,
				"audio/mp4", "m4a"
		}
	case "apng":
		return gtsmodel.FileTypeImage,
			"image/apng", "apng"
	case "png_pipe":
		return gtsmodel.FileTypeImage,
			"image/png", "png"
	case "image2", "image2pipe", "jpeg_pipe":
		return gtsmodel.FileTypeImage,
			"image/jpeg", "jpeg"
	case "webp", "webp_pipe":
		return gtsmodel.FileTypeImage,
			"image/webp", "webp"
	case "gif":
		return gtsmodel.FileTypeImage,
			"image/gif", "gif"
	case "mp3":
		if len(res.audio) > 0 {
			switch res.audio[0].codec {
			case "mp2":
				return gtsmodel.FileTypeAudio,
					"audio/mp2", "mp2"
			case "mp3":
				return gtsmodel.FileTypeAudio,
					"audio/mp3", "mp3"
			}
		}
	case "asf":
		switch {
		case len(res.video) > 0:
			return gtsmodel.FileTypeVideo,
				"video/x-ms-wmv", "wmv"
		case len(res.audio) > 0:
			return gtsmodel.FileTypeAudio,
				"audio/x-ms-wma", "wma"
		}
	case "ogg":
		if len(res.video) > 0 {
			switch res.video[0].codec {
			case "theora", "dirac": // daala, tarkin
				return gtsmodel.FileTypeVideo,
					"video/ogg", "ogv"
			}
		}
		if len(res.audio) > 0 {
			switch res.audio[0].codec {
			case "opus", "libopus":
				return gtsmodel.FileTypeAudio,
					"audio/opus", "opus"
			default:
				return gtsmodel.FileTypeAudio,
					"audio/ogg", "ogg"
			}
		}
	case "matroska,webm":
		switch {
		case len(res.video) > 0:
			var isWebm bool

			switch res.video[0].codec {
			case "vp8", "vp9", "av1":
				if len(res.audio) > 0 {
					switch res.audio[0].codec {
					case "vorbis", "opus", "libopus":
						// webm only supports [VP8/VP9/AV1] +
						//                    [vorbis/opus]
						isWebm = true
					}
				} else {
					// no audio with correct
					// video codec also fine.
					isWebm = true
				}
			}

			if isWebm {
				// Check valid webm codec config.
				return gtsmodel.FileTypeVideo,
					"video/webm", "webm"
			}

			// All else falls under generic mkv.
			return gtsmodel.FileTypeVideo,
				"video/x-matroska", "mkv"
		case len(res.audio) > 0:
			return gtsmodel.FileTypeAudio,
				"audio/x-matroska", "mka"
		}
	case "avi":
		return gtsmodel.FileTypeVideo,
			"video/x-msvideo", "avi"
	case "flac":
		return gtsmodel.FileTypeAudio,
			"audio/flac", "flac"
	}
	return gtsmodel.FileTypeUnknown,
		"", res.format
}

// ImageMeta extracts image metadata contained within ffprobe'd media result streams.
func (res *result) ImageMeta() (width int, height int, framerate float32) {
	for _, stream := range res.video {
		// Use widest found width.
		if stream.width > width {
			width = stream.width
		}

		// Use tallest found height.
		if stream.height > height {
			height = stream.height
		}

		// Use lowest non-zero (valid) framerate.
		if fr := float32(stream.framerate); fr > 0 {
			if framerate == 0 || fr < framerate {
				framerate = fr
			}
		}
	}

	// If image is rotated by
	// any odd multiples of 90,
	// flip width / height to
	// get the correct scale.
	switch res.orientation {
	case orientationRotate90,
		orientationRotate270,
		orientationTransverse,
		orientationTranspose:
		width, height = height, width
	}

	return
}

// PixFmt returns the first valid pixel format
// contained among the result vidoe streams.
func (res *result) PixFmt() string {
	for _, str := range res.video {
		if str.pixfmt != "" {
			return str.pixfmt
		}
	}
	return ""
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

	// Check extra packet / frame information
	// for provided orientation (if provided).
	for _, pf := range res.PacketsAndFrames {

		// Ensure frame contains tags.
		if pf.Tags.Orientation == "" {
			continue
		}

		// Trim any space from orientation value.
		str := strings.TrimSpace(pf.Tags.Orientation)

		// Parse as integer value.
		orient, _ := strconv.Atoi(str)
		if orient < 0 || orient >= 9 {
			return nil, errors.New("invalid orientation data")
		}

		// Ensure different value has
		// not already been specified.
		if r.orientation != 0 &&
			orient != r.orientation {
			return nil, errors.New("multiple sets of orientation / rotation data")
		}

		// Set new orientation.
		r.orientation = orient
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
			// Parse stream framerate, bearing in
			// mind that some static container formats
			// (e.g. jpeg) still return a framerate, so
			// we also check for a non-1 timebase (dts).
			var framerate float32
			if str := s.RFrameRate; str != "" &&
				s.DurationTS > 1 {
				var num, den uint32
				den = 1

				// Check for inequality (numerator / denominator).
				if p := strings.SplitN(str, "/", 2); len(p) == 2 {
					n, _ := strconv.ParseUint(p[0], 10, 32)
					d, _ := strconv.ParseUint(p[1], 10, 32)
					num, den = uint32(n), uint32(d) // #nosec G115 -- ParseUint is configured to check
				} else {
					n, _ := strconv.ParseUint(p[0], 10, 32)
					num = uint32(n) // #nosec G115 -- ParseUint is configured to check
				}

				// Set final divised framerate.
				framerate = float32(num / den)
			}

			// Check for embedded sidedata
			// which may contain rotation data.
			for _, d := range s.SideDataList {

				// Ensure frame side
				// data IS rotation data.
				if d.Rotation == 0 {
					continue
				}

				// Drop any decimal
				// rotation value.
				rot := int(d.Rotation)

				// Round rotation to multiple of 90.
				// More granularity is not needed.
				if q := rot % 90; q > 45 {
					rot += (90 - q)
				} else {
					rot -= q
				}

				// Drop any value above 360
				// or below -360, these are
				// just repeat full turns.
				//
				// Then convert to
				// orientation value.
				var orient int
				switch rot % 360 {
				case 0:
					orient = orientationNormal
				case 90, -270:
					orient = orientationRotate90
				case 180:
					orient = orientationRotate180
				case 270, -90:
					orient = orientationRotate270
				}

				// Ensure different value has
				// not already been specified.
				if r.orientation != 0 &&
					orient != r.orientation {
					return nil, errors.New("multiple sets of orientation / rotation data")
				}

				// Set new orientation.
				r.orientation = orient
			}

			// Append video stream data to result.
			r.video = append(r.video, videoStream{
				stream:    stream{codec: s.CodecName},
				pixfmt:    s.PixFmt,
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
	PacketsAndFrames []ffprobePacketOrFrame `json:"packets_and_frames"`
	Streams          []ffprobeStream        `json:"streams"`
	Format           *ffprobeFormat         `json:"format"`
	Error            *ffprobeError          `json:"error"`
}

type ffprobePacketOrFrame struct {
	Type string      `json:"type"`
	Tags ffprobeTags `json:"tags"`
	// SideDataList []ffprobeSideData `json:"side_data_list"`
}

type ffprobeTags struct {
	Orientation string `json:"orientation"`
}

type ffprobeStream struct {
	CodecName    string            `json:"codec_name"`
	CodecType    string            `json:"codec_type"`
	PixFmt       string            `json:"pix_fmt"`
	RFrameRate   string            `json:"r_frame_rate"`
	DurationTS   uint              `json:"duration_ts"`
	Width        int               `json:"width"`
	Height       int               `json:"height"`
	SideDataList []ffprobeSideData `json:"side_data_list"`
}

type ffprobeSideData struct {
	Rotation float64 `json:"rotation"`
}

type ffprobeFormat struct {
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"`
	BitRate    string `json:"bit_rate"`
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

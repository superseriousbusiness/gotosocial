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
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/buckket/go-blurhash"
	"golang.org/x/image/webp"
)

// thumbSize returns the dimensions to use for an input
// image of given width / height, for its outgoing thumbnail.
// This attempts to maintains the original image aspect ratio.
func thumbSize(max, width, height int, aspect float32) (int, int) {

	switch {
	// Simplest case, within bounds!
	case width < max && height < max:
		return width, height

	// Width is larger side.
	case width > height:

		// i.e. height = newWidth * (height / width)
		height = int(float32(max) / aspect)
		return max, height

	// Height is larger side.
	case height > width:

		// i.e. width = newHeight * (width / height)
		width = int(float32(max) * aspect)
		return width, max

	// Square.
	default:
		return max, max
	}
}

// generateThumb generates a thumbnail for the
// input file at path, resizing it to the given
// dimensions and generating a blurhash if needed.
// This wraps much of the complex thumbnailing
// logic in which where possible we use native
// Go libraries for generating thumbnails, else
// always falling back to slower but much more
// widely supportive ffmpeg.
func generateThumb(
	ctx context.Context,
	filepath string,
	width, height int,
	orientation int,
	pixfmt string,
	needBlurhash bool,
) (
	outpath string,
	mimeType string,
	blurhash string,
	err error,
) {
	var ext string

	// Default type is webp.
	mimeType = "image/webp"

	// Generate thumb output path REPLACING extension.
	if i := strings.IndexByte(filepath, '.'); i != -1 {
		outpath = filepath[:i] + "_thumb.webp"
		ext = filepath[i+1:] // old extension
	} else {
		return "", "", "", gtserror.New("input file missing extension")
	}

	// Check for the few media types we
	// have native Go decoding that allow
	// us to generate thumbs natively.
	switch {

	case ext == "jpeg":
		// Replace the "webp" with "jpeg", as we'll
		// use our native Go thumbnailing generation.
		outpath = outpath[:len(outpath)-4] + "jpeg"
		mimeType = "image/jpeg"

		log.Debug(ctx, "generating thumb from jpeg")
		blurhash, err := generateNativeThumb(
			filepath,
			outpath,
			width,
			height,
			orientation,
			jpeg.Decode,
			needBlurhash,
		)
		return outpath, mimeType, blurhash, err

	// We specifically only allow generating native
	// thumbnails from gif IF it doesn't contain an
	// alpha channel. We'll ultimately be encoding to
	// jpeg which doesn't support transparency layers.
	case ext == "gif" && !containsAlpha(pixfmt):

		// Replace the "webp" with "jpeg", as we'll
		// use our native Go thumbnailing generation.
		outpath = outpath[:len(outpath)-4] + "jpeg"
		mimeType = "image/jpeg"

		log.Debug(ctx, "generating thumb from gif")
		blurhash, err := generateNativeThumb(
			filepath,
			outpath,
			width,
			height,
			orientation,
			gif.Decode,
			needBlurhash,
		)
		return outpath, mimeType, blurhash, err

	// We specifically only allow generating native
	// thumbnails from png IF it doesn't contain an
	// alpha channel. We'll ultimately be encoding to
	// jpeg which doesn't support transparency layers.
	case ext == "png" && !containsAlpha(pixfmt):

		// Replace the "webp" with "jpeg", as we'll
		// use our native Go thumbnailing generation.
		outpath = outpath[:len(outpath)-4] + "jpeg"
		mimeType = "image/jpeg"

		log.Debug(ctx, "generating thumb from png")
		blurhash, err := generateNativeThumb(
			filepath,
			outpath,
			width,
			height,
			orientation,
			png.Decode,
			needBlurhash,
		)
		return outpath, mimeType, blurhash, err

	// We specifically only allow generating native
	// thumbnails from webp IF it doesn't contain an
	// alpha channel. We'll ultimately be encoding to
	// jpeg which doesn't support transparency layers.
	case ext == "webp" && !containsAlpha(pixfmt):

		// Replace the "webp" with "jpeg", as we'll
		// use our native Go thumbnailing generation.
		outpath = outpath[:len(outpath)-4] + "jpeg"
		mimeType = "image/jpeg"

		log.Debug(ctx, "generating thumb from webp")
		blurhash, err := generateNativeThumb(
			filepath,
			outpath,
			width,
			height,
			orientation,
			webp.Decode,
			needBlurhash,
		)
		return outpath, mimeType, blurhash, err
	}

	// The fallback for thumbnail generation, which
	// encompasses most media types is with ffmpeg.
	log.Debug(ctx, "generating thumb with ffmpeg")
	if err := ffmpegGenerateWebpThumb(ctx,
		filepath,
		outpath,
		width,
		height,
		pixfmt,
	); err != nil {
		return outpath, "", "", err
	}

	if needBlurhash {
		// Generate new blurhash from webp output thumb.
		blurhash, err = generateWebpBlurhash(outpath)
		if err != nil {
			return outpath, "", "", gtserror.Newf("error generating blurhash: %w", err)
		}
	}

	return outpath, mimeType, blurhash, nil
}

// generateNativeThumb generates a thumbnail
// using native Go code, using given decode
// function to get image, resize to given dimens,
// and write to output filepath as JPEG. If a
// blurhash is required it will also generate
// this from the image.Image while in-memory.
func generateNativeThumb(
	inpath, outpath string,
	width, height int,
	orientation int,
	decode func(io.Reader) (image.Image, error),
	needBlurhash bool,
) (
	string, // blurhash
	error,
) {
	// Open input file at given path.
	infile, err := os.Open(inpath)
	if err != nil {
		return "", gtserror.Newf("error opening input file %s: %w", inpath, err)
	}

	// Decode image into memory.
	img, err := decode(infile)

	// Done with file.
	_ = infile.Close()

	if err != nil {
		return "", gtserror.Newf("error decoding file %s: %w", inpath, err)
	}

	// Apply orientation BEFORE any resize,
	// as our image dimensions are calculated
	// taking orientation into account.
	switch orientation {
	case orientationFlipH:
		img = flipH(img)
	case orientationFlipV:
		img = flipV(img)
	case orientationRotate90:
		img = rotate90(img)
	case orientationRotate180:
		img = rotate180(img)
	case orientationRotate270:
		img = rotate270(img)
	case orientationTranspose:
		img = transpose(img)
	case orientationTransverse:
		img = transverse(img)
	}

	// Resize image to dimens.
	img = resizeDownLinear(img,
		width, height,
	)

	// Open output file at given path.
	outfile, err := os.Create(outpath)
	if err != nil {
		return "", gtserror.Newf("error opening output file %s: %w", outpath, err)
	}

	// Encode in-memory image to output file.
	// (nil uses defaults, i.e. quality=75).
	err = jpeg.Encode(outfile, img, nil)

	// Done with file.
	_ = outfile.Close()

	if err != nil {
		return "", gtserror.Newf("error encoding image: %w", err)
	}

	if needBlurhash {
		// for generating blurhashes, it's more
		// cost effective to lose detail since
		// it's blurry, so make a tiny version.
		tiny := resizeDownLinear(img, 32, 0)

		// Drop the larger image
		// ref as soon as possible
		// to allow GC to claim.
		img = nil //nolint

		// Generate blurhash for the tiny thumbnail.
		blurhash, err := blurhash.Encode(4, 3, tiny)
		if err != nil {
			return "", gtserror.Newf("error generating blurhash: %w", err)
		}

		return blurhash, nil
	}

	return "", nil
}

// generateWebpBlurhash generates a blurhash for Webp at filepath.
func generateWebpBlurhash(filepath string) (string, error) {
	// Open the file at given path.
	file, err := os.Open(filepath)
	if err != nil {
		return "", gtserror.Newf("error opening input file %s: %w", filepath, err)
	}

	// Decode image from file.
	img, err := webp.Decode(file)

	// Done with file.
	_ = file.Close()

	if err != nil {
		return "", gtserror.Newf("error decoding file %s: %w", filepath, err)
	}

	// for generating blurhashes, it's more
	// cost effective to lose detail since
	// it's blurry, so make a tiny version.
	tiny := resizeDownLinear(img, 32, 0)

	// Drop the larger image
	// ref as soon as possible
	// to allow GC to claim.
	img = nil //nolint

	// Generate blurhash for the tiny thumbnail.
	blurhash, err := blurhash.Encode(4, 3, tiny)
	if err != nil {
		return "", gtserror.Newf("error generating blurhash: %w", err)
	}

	return blurhash, nil
}

// containsAlpha returns whether given pixfmt
// (i.e. colorspace) contains an alpha channel.
//
// generated with: for entry in $(ffprobe.wasm -show_entries pixel_format=name:flags=alpha 2>/dev/null | grep -B1 alpha=1 | sed -n '/name=/{ s|name=||; p }'); do printf '\tcase "%s":\n\t\treturn true\n' "$entry"; done
func containsAlpha(pixfmt string) bool {
	switch pixfmt {
	case "pal8":
		return true
	case "argb":
		return true
	case "rgba":
		return true
	case "abgr":
		return true
	case "bgra":
		return true
	case "yuva420p":
		return true
	case "ya8":
		return true
	case "yuva422p":
		return true
	case "yuva444p":
		return true
	case "yuva420p9be":
		return true
	case "yuva420p9le":
		return true
	case "yuva422p9be":
		return true
	case "yuva422p9le":
		return true
	case "yuva444p9be":
		return true
	case "yuva444p9le":
		return true
	case "yuva420p10be":
		return true
	case "yuva420p10le":
		return true
	case "yuva422p10be":
		return true
	case "yuva422p10le":
		return true
	case "yuva444p10be":
		return true
	case "yuva444p10le":
		return true
	case "yuva420p16be":
		return true
	case "yuva420p16le":
		return true
	case "yuva422p16be":
		return true
	case "yuva422p16le":
		return true
	case "yuva444p16be":
		return true
	case "yuva444p16le":
		return true
	case "rgba64be":
		return true
	case "rgba64le":
		return true
	case "bgra64be":
		return true
	case "bgra64le":
		return true
	case "ya16be":
		return true
	case "ya16le":
		return true
	case "gbrap":
		return true
	case "gbrap16be":
		return true
	case "gbrap16le":
		return true
	case "ayuv64le":
		return true
	case "ayuv64be":
		return true
	case "gbrap12be":
		return true
	case "gbrap12le":
		return true
	case "gbrap10be":
		return true
	case "gbrap10le":
		return true
	case "gbrapf32be":
		return true
	case "gbrapf32le":
		return true
	case "yuva422p12be":
		return true
	case "yuva422p12le":
		return true
	case "yuva444p12be":
		return true
	case "yuva444p12le":
		return true
	default:
		return false
	}
}

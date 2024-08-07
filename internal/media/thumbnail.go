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

	"github.com/buckket/go-blurhash"
	"github.com/disintegration/imaging"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"golang.org/x/image/webp"
)

func generateThumb(
	ctx context.Context,
	filepath string,
	width, height int,
	pixfmt string,
	needBlurhash bool,
) (
	outpath string,
	blurhash string,
	err error,
) {
	var ext string

	// Generate thumb output path REPLACING extension.
	if i := strings.IndexByte(filepath, '.'); i != -1 {
		outpath = filepath[:i] + "_thumb.webp"
		ext = filepath[i+1:] // old extension
	} else {
		return "", "", gtserror.New("input file missing extension")
	}

	// Check for the few media types we
	// have native Go decoding that allow
	// us to generate thumbs natively.
	switch {

	case ext == "jpeg":
		// Replace the "webp" with "jpeg", as we'll
		// use our native Go thumbnailing generation.
		outpath = outpath[:len(outpath)-4] + "jpeg"

		log.Debug(ctx, "generating thumb from jpeg")
		blurhash, err := generateNativeThumb(
			filepath,
			outpath,
			width,
			height,
			jpeg.Decode,
			needBlurhash,
		)
		return outpath, blurhash, err

	case ext == "gif":
		// Replace the "webp" with "jpeg", as we'll
		// use our native Go thumbnailing generation.
		outpath = outpath[:len(outpath)-4] + "jpeg"

		log.Debug(ctx, "generating thumb from gif")
		blurhash, err := generateNativeThumb(
			filepath,
			outpath,
			width,
			height,
			gif.Decode,
			needBlurhash,
		)
		return outpath, blurhash, err

	// We specifically only allow generating native
	// thumbnails from png IF it doesn't contain an
	// alpha channel. We'll ultimately be encoding to
	// jpeg which doesn't support transparency layers.
	case ext == "png" && !containsAlpha(pixfmt):

		// Replace the "webp" with "jpeg", as we'll
		// use our native Go thumbnailing generation.
		outpath = outpath[:len(outpath)-4] + "jpeg"

		log.Debug(ctx, "generating thumb from png")
		blurhash, err := generateNativeThumb(
			filepath,
			outpath,
			width,
			height,
			png.Decode,
			needBlurhash,
		)
		return outpath, blurhash, err

	// We specifically only allow generating native
	// thumbnails from webp IF it doesn't contain an
	// alpha channel. We'll ultimately be encoding to
	// jpeg which doesn't support transparency layers.
	case ext == "webp" && !containsAlpha(pixfmt):

		// Replace the "webp" with "jpeg", as we'll
		// use our native Go thumbnailing generation.
		outpath = outpath[:len(outpath)-4] + "jpeg"

		log.Debug(ctx, "generating thumb from webp")
		blurhash, err := generateNativeThumb(
			filepath,
			outpath,
			width,
			height,
			webp.Decode,
			needBlurhash,
		)
		return outpath, blurhash, err
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
		return outpath, "", err
	}

	if needBlurhash {
		// Generate new blurhash from webp output thumb.
		blurhash, err = generateWebpBlurhash(outpath)
		if err != nil {
			return outpath, "", gtserror.Newf("error generating blurhash: %w", err)
		}
	}

	return outpath, blurhash, err
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

	// Resize image to dimens.
	img = imaging.Resize(img,
		width, height,
		imaging.NearestNeighbor,
	)

	// Open output file at given path.
	outfile, err := os.Create(outpath)
	if err != nil {
		return "", gtserror.Newf("error opening output file %s: %w", outpath, err)
	}

	// Encode in-memory image to output file.
	err = jpeg.Encode(outfile, img, &jpeg.Options{
		Quality: 70, // good enough for thumb
	})

	// Done with file.
	_ = outfile.Close()

	if err != nil {
		return "", gtserror.Newf("error encoding image: %w", err)
	}

	if needBlurhash {
		// for generating blurhashes, it's more cost effective to
		// lose detail since it's blurry, so make a tiny version.
		tiny := imaging.Resize(img, 64, 64, imaging.NearestNeighbor)

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

	// for generating blurhashes, it's more cost effective to
	// lose detail since it's blurry, so make a tiny version.
	tiny := imaging.Resize(img, 64, 64, imaging.NearestNeighbor)

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

// List of pixel formats that have an alpha layer.
// Derived from the following very messy command:
//
//	for res in $(ffprobe -show_entries pixel_format=name:flags=alpha | grep -B1 alpha=1 | grep name); do echo $res | sed 's/name=//g' | sed 's/^/"/g' | sed 's/$/",/g'; done
var alphaPixelFormats = []string{
	"pal8",
	"argb",
	"rgba",
	"abgr",
	"bgra",
	"yuva420p",
	"ya8",
	"yuva422p",
	"yuva444p",
	"yuva420p9be",
	"yuva420p9le",
	"yuva422p9be",
	"yuva422p9le",
	"yuva444p9be",
	"yuva444p9le",
	"yuva420p10be",
	"yuva420p10le",
	"yuva422p10be",
	"yuva422p10le",
	"yuva444p10be",
	"yuva444p10le",
	"yuva420p16be",
	"yuva420p16le",
	"yuva422p16be",
	"yuva422p16le",
	"yuva444p16be",
	"yuva444p16le",
	"rgba64be",
	"rgba64le",
	"bgra64be",
	"bgra64le",
	"ya16be",
	"ya16le",
	"gbrap",
	"gbrap16be",
	"gbrap16le",
	"ayuv64le",
	"ayuv64be",
	"gbrap12be",
	"gbrap12le",
	"gbrap10be",
	"gbrap10le",
	"gbrapf32be",
	"gbrapf32le",
	"yuva422p12be",
	"yuva422p12le",
	"yuva444p12be",
	"yuva444p12le",
}

// containsAlpha returns whether given pixfmt
// (i.e. colorspace) contains an alpha channel.
func containsAlpha(pixfmt string) bool {
	if pixfmt == "" {
		return false
	}
	for _, checkfmt := range alphaPixelFormats {
		if pixfmt == checkfmt {
			return true
		}
	}
	return false
}

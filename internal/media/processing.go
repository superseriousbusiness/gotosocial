package media

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/buckket/go-blurhash"
	"github.com/nfnt/resize"
	"github.com/superseriousbusiness/exifremove/pkg/exifremove"
)

// purgeExif is a little wrapper for the action of removing exif data from an image.
// Only pass pngs or jpegs to this function.
func purgeExif(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("passed image was not valid")
	}

	clean, err := exifremove.Remove(data)
	if err != nil {
		return nil, fmt.Errorf("could not purge exif from image: %s", err)
	}

	if len(clean) == 0 {
		return nil, errors.New("purged image was not valid")
	}
	
	return clean, nil
}

func deriveGif(b []byte, extension string) (*imageAndMeta, error) {
	var g *gif.GIF
	var err error
	switch extension {
	case mimeGif:
		g, err = gif.DecodeAll(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("extension %s not recognised", extension)
	}

	// use the first frame to get the static characteristics
	width := g.Config.Width
	height := g.Config.Height
	size := width * height
	aspect := float64(width) / float64(height)

	return &imageAndMeta{
		image:  b,
		width:  width,
		height: height,
		size:   size,
		aspect: aspect,
	}, nil
}

func deriveImage(b []byte, contentType string) (*imageAndMeta, error) {
	var i image.Image
	var err error

	switch contentType {
	case mimeImageJpeg:
		i, err = jpeg.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case mimeImagePng:
		i, err = png.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("content type %s not recognised", contentType)
	}

	width := i.Bounds().Size().X
	height := i.Bounds().Size().Y
	size := width * height
	aspect := float64(width) / float64(height)

	return &imageAndMeta{
		image:  b,
		width:  width,
		height: height,
		size:   size,
		aspect: aspect,
	}, nil
}

// deriveThumbnail returns a byte slice and metadata for a thumbnail of width x and height y,
// of a given jpeg, png, or gif, or an error if something goes wrong.
//
// Note that the aspect ratio of the image will be retained,
// so it will not necessarily be a square, even if x and y are set as the same value.
func deriveThumbnail(b []byte, contentType string, x uint, y uint) (*imageAndMeta, error) {
	var i image.Image
	var err error

	switch contentType {
	case mimeImageJpeg:
		i, err = jpeg.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case mimeImagePng:
		i, err = png.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case mimeImageGif:
		i, err = gif.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("content type %s not recognised", contentType)
	}

	thumb := resize.Thumbnail(x, y, i, resize.NearestNeighbor)
	width := thumb.Bounds().Size().X
	height := thumb.Bounds().Size().Y
	size := width * height
	aspect := float64(width) / float64(height)

	tiny := resize.Thumbnail(32, 32, thumb, resize.NearestNeighbor)
	bh, err := blurhash.Encode(4, 3, tiny)
	if err != nil {
		return nil, err
	}

	out := &bytes.Buffer{}
	if err := jpeg.Encode(out, thumb, &jpeg.Options{
		Quality: 75,
	}); err != nil {
		return nil, err
	}
	return &imageAndMeta{
		image:    out.Bytes(),
		width:    width,
		height:   height,
		size:     size,
		aspect:   aspect,
		blurhash: bh,
	}, nil
}

// deriveStaticEmojji takes a given gif or png of an emoji, decodes it, and re-encodes it as a static png.
func deriveStaticEmoji(b []byte, contentType string) (*imageAndMeta, error) {
	var i image.Image
	var err error

	switch contentType {
	case mimeImagePng:
		i, err = png.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case mimeImageGif:
		i, err = gif.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("content type %s not allowed for emoji", contentType)
	}

	out := &bytes.Buffer{}
	if err := png.Encode(out, i); err != nil {
		return nil, err
	}
	return &imageAndMeta{
		image: out.Bytes(),
	}, nil
}

type imageAndMeta struct {
	image    []byte
	width    int
	height   int
	size     int
	aspect   float64
	blurhash string
}

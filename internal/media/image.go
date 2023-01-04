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
	"bufio"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"sync"

	"github.com/buckket/go-blurhash"
	"github.com/disintegration/imaging"
	"github.com/superseriousbusiness/gotosocial/internal/iotools"

	// import to init webp encode/decoding.
	_ "golang.org/x/image/webp"
)

var (
	// pngEncoder provides our global PNG encoding with
	// specified compression level, and memory pooled buffers.
	pngEncoder = png.Encoder{
		CompressionLevel: png.DefaultCompression,
		BufferPool:       &pngEncoderBufferPool{},
	}

	// jpegBufferPool is a memory pool of byte buffers for JPEG encoding.
	jpegBufferPool = sync.Pool{
		New: func() any {
			return bufio.NewWriter(nil)
		},
	}
)

// gtsImage is a thin wrapper around the standard library image
// interface to provide our own useful helper functions for image
// size and aspect ratio calculations, streamed encoding to various
// types, and creating reduced size thumbnail images.
type gtsImage struct{ image image.Image }

// blankImage generates a blank image of given dimensions.
func blankImage(width int, height int) *gtsImage {
	// create a rectangle with the same dimensions as the video
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// fill the rectangle with our desired fill color.
	draw.Draw(img, img.Bounds(), &image.Uniform{
		color.RGBA{42, 43, 47, 0},
	}, image.Point{}, draw.Src)

	return &gtsImage{image: img}
}

// decodeImage will decode image from reader stream and return image wrapped in our own gtsImage{} type.
func decodeImage(r io.Reader, opts ...imaging.DecodeOption) (*gtsImage, error) {
	img, err := imaging.Decode(r, opts...)
	if err != nil {
		return nil, err
	}
	return &gtsImage{image: img}, nil
}

// Width returns the image width in pixels.
func (m *gtsImage) Width() uint32 {
	return uint32(m.image.Bounds().Size().X)
}

// Height returns the image height in pixels.
func (m *gtsImage) Height() uint32 {
	return uint32(m.image.Bounds().Size().Y)
}

// Size returns the total number of image pixels.
func (m *gtsImage) Size() uint64 {
	return uint64(m.image.Bounds().Size().X) *
		uint64(m.image.Bounds().Size().Y)
}

// AspectRatio returns the image ratio of width:height.
func (m *gtsImage) AspectRatio() float32 {
	return float32(m.image.Bounds().Size().X) /
		float32(m.image.Bounds().Size().Y)
}

// Thumbnail returns a small sized copy of gtsImage{}, limited to 512x512 if not small enough.
func (m *gtsImage) Thumbnail() *gtsImage {
	const (
		// max thumb
		// dimensions.
		maxWidth  = 512
		maxHeight = 512
	)

	// Check the receiving image is within max thumnail bounds.
	if m.Width() <= maxWidth && m.Height() <= maxHeight {
		return &gtsImage{image: imaging.Clone(m.image)}
	}

	// Image is too large, needs to be resized to thumbnail max.
	img := imaging.Fit(m.image, maxWidth, maxHeight, imaging.Linear)
	return &gtsImage{image: img}
}

// Blurhash calculates the blurhash for the receiving image data.
func (m *gtsImage) Blurhash() (string, error) {
	// for generating blurhashes, it's more cost effective to
	// lose detail since it's blurry, so make a tiny version.
	tiny := imaging.Resize(m.image, 32, 0, imaging.NearestNeighbor)

	// Encode blurhash from resized version
	return blurhash.Encode(4, 3, tiny)
}

// ToJPEG creates a new streaming JPEG encoder from receiving image, and a size ptr
// which stores the number of bytes written during the image encoding process.
func (m *gtsImage) ToJPEG(opts *jpeg.Options) io.Reader {
	return iotools.StreamWriteFunc(func(w io.Writer) error {
		// Get encoding buffer
		bw := getJPEGBuffer(w)

		// Encode JPEG to buffered writer.
		err := jpeg.Encode(bw, m.image, opts)

		// Replace buffer.
		//
		// NOTE: jpeg.Encode() already
		// performs a bufio.Writer.Flush().
		putJPEGBuffer(bw)

		return err
	})
}

// ToPNG creates a new streaming PNG encoder from receiving image, and a size ptr
// which stores the number of bytes written during the image encoding process.
func (m *gtsImage) ToPNG() io.Reader {
	return iotools.StreamWriteFunc(func(w io.Writer) error {
		return pngEncoder.Encode(w, m.image)
	})
}

// getJPEGBuffer fetches a reset JPEG encoding buffer from global JPEG buffer pool.
func getJPEGBuffer(w io.Writer) *bufio.Writer {
	buf, _ := jpegBufferPool.Get().(*bufio.Writer)
	buf.Reset(w)
	return buf
}

// putJPEGBuffer resets the given bufio writer and places in global JPEG buffer pool.
func putJPEGBuffer(buf *bufio.Writer) {
	buf.Reset(nil)
	jpegBufferPool.Put(buf)
}

// pngEncoderBufferPool implements png.EncoderBufferPool.
type pngEncoderBufferPool sync.Pool

func (p *pngEncoderBufferPool) Get() *png.EncoderBuffer {
	buf, _ := (*sync.Pool)(p).Get().(*png.EncoderBuffer)
	return buf
}

func (p *pngEncoderBufferPool) Put(buf *png.EncoderBuffer) {
	(*sync.Pool)(p).Put(buf)
}

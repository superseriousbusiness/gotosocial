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
	"image"
	"image/color"
	"math"
)

// NOTE:
// the following code is borrowed from
// github.com/disintegration/imaging
// and collapses in some places for our
// particular usecases and with parallel()
// function (spans work across goroutines)
// removed, instead working synchronously.
//
// at gotosocial we take particular
// care about where we spawn goroutines
// to ensure we're in control of the
// amount of concurrency in relation
// to the amount configured by user.

// resizeDownLinear resizes image to given width x height using linear resampling.
// This is specifically optimized for resizing down (i.e. smaller), else is noop.
func resizeDownLinear(img image.Image, width, height int) image.Image {
	srcW, srcH := img.Bounds().Dx(), img.Bounds().Dy()
	if srcW <= 0 || srcH <= 0 ||
		width < 0 || height < 0 {
		return &image.NRGBA{}
	}

	if width == 0 {
		// If no width is given, use aspect preserving width.
		tmp := float64(height) * float64(srcW) / float64(srcH)
		width = int(math.Max(1.0, math.Floor(tmp+0.5)))
	}

	if height == 0 {
		// If no height is given, use aspect preserving height.
		tmp := float64(width) * float64(srcH) / float64(srcW)
		height = int(math.Max(1.0, math.Floor(tmp+0.5)))
	}

	if width < srcW {
		// Width is smaller, resize horizontally.
		img = resizeHorizontalLinear(img, width)
	}

	if height < srcH {
		// Height is smaller, resize vertically.
		img = resizeVerticalLinear(img, height)
	}

	return img
}

// flipH flips the image horizontally (left to right).
func flipH(img image.Image) image.Image {
	src := newScanner(img)
	dstW := src.w
	dstH := src.h
	rowSize := dstW * 4
	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		i := y * dst.Stride
		srcY := y
		src.scan(0, srcY, src.w, srcY+1, dst.Pix[i:i+rowSize])
		reverse(dst.Pix[i : i+rowSize])
	}
	return dst
}

// flipV flips the image vertically (from top to bottom).
func flipV(img image.Image) image.Image {
	src := newScanner(img)
	dstW := src.w
	dstH := src.h
	rowSize := dstW * 4
	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		i := y * dst.Stride
		srcY := dstH - y - 1
		src.scan(0, srcY, src.w, srcY+1, dst.Pix[i:i+rowSize])
	}
	return dst
}

// rotate90 rotates the image 90 counter-clockwise.
func rotate90(img image.Image) image.Image {
	src := newScanner(img)
	dstW := src.h
	dstH := src.w
	rowSize := dstW * 4
	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		i := y * dst.Stride
		srcX := dstH - y - 1
		src.scan(srcX, 0, srcX+1, src.h, dst.Pix[i:i+rowSize])
	}
	return dst
}

// rotate180 rotates the image 180 counter-clockwise.
func rotate180(img image.Image) image.Image {
	src := newScanner(img)
	dstW := src.w
	dstH := src.h
	rowSize := dstW * 4
	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		i := y * dst.Stride
		srcY := dstH - y - 1
		src.scan(0, srcY, src.w, srcY+1, dst.Pix[i:i+rowSize])
		reverse(dst.Pix[i : i+rowSize])
	}
	return dst
}

// rotate270 rotates the image 270 counter-clockwise.
func rotate270(img image.Image) image.Image {
	src := newScanner(img)
	dstW := src.h
	dstH := src.w
	rowSize := dstW * 4
	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		i := y * dst.Stride
		srcX := y
		src.scan(srcX, 0, srcX+1, src.h, dst.Pix[i:i+rowSize])
		reverse(dst.Pix[i : i+rowSize])
	}
	return dst
}

// transpose flips the image horizontally and rotates 90 counter-clockwise.
func transpose(img image.Image) image.Image {
	src := newScanner(img)
	dstW := src.h
	dstH := src.w
	rowSize := dstW * 4
	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		i := y * dst.Stride
		srcX := y
		src.scan(srcX, 0, srcX+1, src.h, dst.Pix[i:i+rowSize])
	}
	return dst
}

// transverse flips the image vertically and rotates 90 counter-clockwise.
func transverse(img image.Image) image.Image {
	src := newScanner(img)
	dstW := src.h
	dstH := src.w
	rowSize := dstW * 4
	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		i := y * dst.Stride
		srcX := dstH - y - 1
		src.scan(srcX, 0, srcX+1, src.h, dst.Pix[i:i+rowSize])
		reverse(dst.Pix[i : i+rowSize])
	}
	return dst
}

// resizeHorizontalLinear resizes image to given width using linear resampling.
func resizeHorizontalLinear(img image.Image, dstWidth int) image.Image {
	src := newScanner(img)
	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, src.h))
	weights := precomputeWeightsLinear(dstWidth, src.w)
	scanLine := make([]uint8, src.w*4)
	for y := 0; y < src.h; y++ {
		src.scan(0, y, src.w, y+1, scanLine)
		j0 := y * dst.Stride
		for x := range weights {
			var r, g, b, a float64
			for _, w := range weights[x] {
				i := w.index * 4
				s := scanLine[i : i+4 : i+4]
				aw := float64(s[3]) * w.weight
				r += float64(s[0]) * aw
				g += float64(s[1]) * aw
				b += float64(s[2]) * aw
				a += aw
			}
			if a != 0 {
				aInv := 1 / a
				j := j0 + x*4
				d := dst.Pix[j : j+4 : j+4]
				d[0] = clampFloat(r * aInv)
				d[1] = clampFloat(g * aInv)
				d[2] = clampFloat(b * aInv)
				d[3] = clampFloat(a)
			}
		}
	}
	return dst
}

// resizeVerticalLinear resizes image to given height using linear resampling.
func resizeVerticalLinear(img image.Image, height int) image.Image {
	src := newScanner(img)
	dst := image.NewNRGBA(image.Rect(0, 0, src.w, height))
	weights := precomputeWeightsLinear(height, src.h)
	scanLine := make([]uint8, src.h*4)
	for x := 0; x < src.w; x++ {
		src.scan(x, 0, x+1, src.h, scanLine)
		for y := range weights {
			var r, g, b, a float64
			for _, w := range weights[y] {
				i := w.index * 4
				s := scanLine[i : i+4 : i+4]
				aw := float64(s[3]) * w.weight
				r += float64(s[0]) * aw
				g += float64(s[1]) * aw
				b += float64(s[2]) * aw
				a += aw
			}
			if a != 0 {
				aInv := 1 / a
				j := y*dst.Stride + x*4
				d := dst.Pix[j : j+4 : j+4]
				d[0] = clampFloat(r * aInv)
				d[1] = clampFloat(g * aInv)
				d[2] = clampFloat(b * aInv)
				d[3] = clampFloat(a)
			}
		}
	}
	return dst
}

type indexWeight struct {
	index  int
	weight float64
}

func precomputeWeightsLinear(dstSize, srcSize int) [][]indexWeight {
	du := float64(srcSize) / float64(dstSize)
	scale := du
	if scale < 1.0 {
		scale = 1.0
	}

	ru := math.Ceil(scale)
	out := make([][]indexWeight, dstSize)
	tmp := make([]indexWeight, 0, dstSize*int(ru+2)*2)

	for v := 0; v < dstSize; v++ {
		fu := (float64(v)+0.5)*du - 0.5

		begin := int(math.Ceil(fu - ru))
		if begin < 0 {
			begin = 0
		}
		end := int(math.Floor(fu + ru))
		if end > srcSize-1 {
			end = srcSize - 1
		}

		var sum float64
		for u := begin; u <= end; u++ {
			w := resampleLinear((float64(u) - fu) / scale)
			if w != 0 {
				sum += w
				tmp = append(tmp, indexWeight{index: u, weight: w})
			}
		}
		if sum != 0 {
			for i := range tmp {
				tmp[i].weight /= sum
			}
		}

		out[v] = tmp
		tmp = tmp[len(tmp):]
	}

	return out
}

// resampleLinear is the resample kernel func for linear filtering.
func resampleLinear(x float64) float64 {
	x = math.Abs(x)
	if x < 1.0 {
		return 1.0 - x
	}
	return 0
}

// scanner wraps an image.Image for
// easier size access and image type
// agnostic access to data at coords.
type scanner struct {
	image   image.Image
	w, h    int
	palette []color.NRGBA
}

// newScanner wraps an image.Image in scanner{} type.
func newScanner(img image.Image) *scanner {
	b := img.Bounds()
	s := &scanner{
		image: img,

		w: b.Dx(),
		h: b.Dy(),
	}
	if img, ok := img.(*image.Paletted); ok {
		s.palette = make([]color.NRGBA, len(img.Palette))
		for i := 0; i < len(img.Palette); i++ {
			s.palette[i] = color.NRGBAModel.Convert(img.Palette[i]).(color.NRGBA)
		}
	}
	return s
}

// scan scans the given rectangular region of the image into dst.
func (s *scanner) scan(x1, y1, x2, y2 int, dst []uint8) {
	switch img := s.image.(type) {
	case *image.NRGBA:
		size := (x2 - x1) * 4
		j := 0
		i := y1*img.Stride + x1*4
		if size == 4 {
			for y := y1; y < y2; y++ {
				d := dst[j : j+4 : j+4]
				s := img.Pix[i : i+4 : i+4]
				d[0] = s[0]
				d[1] = s[1]
				d[2] = s[2]
				d[3] = s[3]
				j += size
				i += img.Stride
			}
		} else {
			for y := y1; y < y2; y++ {
				copy(dst[j:j+size], img.Pix[i:i+size])
				j += size
				i += img.Stride
			}
		}

	case *image.NRGBA64:
		j := 0
		for y := y1; y < y2; y++ {
			i := y*img.Stride + x1*8
			for x := x1; x < x2; x++ {
				s := img.Pix[i : i+8 : i+8]
				d := dst[j : j+4 : j+4]
				d[0] = s[0]
				d[1] = s[2]
				d[2] = s[4]
				d[3] = s[6]
				j += 4
				i += 8
			}
		}

	case *image.RGBA:
		j := 0
		for y := y1; y < y2; y++ {
			i := y*img.Stride + x1*4
			for x := x1; x < x2; x++ {
				d := dst[j : j+4 : j+4]
				a := img.Pix[i+3]
				switch a {
				case 0:
					d[0] = 0
					d[1] = 0
					d[2] = 0
					d[3] = a
				case 0xff:
					s := img.Pix[i : i+4 : i+4]
					d[0] = s[0]
					d[1] = s[1]
					d[2] = s[2]
					d[3] = a
				default:
					s := img.Pix[i : i+4 : i+4]
					r16 := uint16(s[0])
					g16 := uint16(s[1])
					b16 := uint16(s[2])
					a16 := uint16(a)
					d[0] = uint8(r16 * 0xff / a16) // #nosec G115 -- Overflow desired.
					d[1] = uint8(g16 * 0xff / a16) // #nosec G115 -- Overflow desired.
					d[2] = uint8(b16 * 0xff / a16) // #nosec G115 -- Overflow desired.
					d[3] = a
				}
				j += 4
				i += 4
			}
		}

	case *image.RGBA64:
		j := 0
		for y := y1; y < y2; y++ {
			i := y*img.Stride + x1*8
			for x := x1; x < x2; x++ {
				s := img.Pix[i : i+8 : i+8]
				d := dst[j : j+4 : j+4]
				a := s[6]
				switch a {
				case 0:
					d[0] = 0
					d[1] = 0
					d[2] = 0
				case 0xff:
					d[0] = s[0]
					d[1] = s[2]
					d[2] = s[4]
				default:
					r32 := uint32(s[0])<<8 | uint32(s[1])
					g32 := uint32(s[2])<<8 | uint32(s[3])
					b32 := uint32(s[4])<<8 | uint32(s[5])
					a32 := uint32(s[6])<<8 | uint32(s[7])
					d[0] = uint8((r32 * 0xffff / a32) >> 8) // #nosec G115 -- Overflow desired.
					d[1] = uint8((g32 * 0xffff / a32) >> 8) // #nosec G115 -- Overflow desired.
					d[2] = uint8((b32 * 0xffff / a32) >> 8) // #nosec G115 -- Overflow desired.
				}
				d[3] = a
				j += 4
				i += 8
			}
		}

	case *image.Gray:
		j := 0
		for y := y1; y < y2; y++ {
			i := y*img.Stride + x1
			for x := x1; x < x2; x++ {
				c := img.Pix[i]
				d := dst[j : j+4 : j+4]
				d[0] = c
				d[1] = c
				d[2] = c
				d[3] = 0xff
				j += 4
				i++
			}
		}

	case *image.Gray16:
		j := 0
		for y := y1; y < y2; y++ {
			i := y*img.Stride + x1*2
			for x := x1; x < x2; x++ {
				c := img.Pix[i]
				d := dst[j : j+4 : j+4]
				d[0] = c
				d[1] = c
				d[2] = c
				d[3] = 0xff
				j += 4
				i += 2
			}
		}

	case *image.YCbCr:
		j := 0
		x1 += img.Rect.Min.X
		x2 += img.Rect.Min.X
		y1 += img.Rect.Min.Y
		y2 += img.Rect.Min.Y

		hy := img.Rect.Min.Y / 2
		hx := img.Rect.Min.X / 2
		for y := y1; y < y2; y++ {
			iy := (y-img.Rect.Min.Y)*img.YStride + (x1 - img.Rect.Min.X)

			var yBase int
			switch img.SubsampleRatio {
			case image.YCbCrSubsampleRatio444, image.YCbCrSubsampleRatio422:
				yBase = (y - img.Rect.Min.Y) * img.CStride
			case image.YCbCrSubsampleRatio420, image.YCbCrSubsampleRatio440:
				yBase = (y/2 - hy) * img.CStride
			}

			for x := x1; x < x2; x++ {
				var ic int
				switch img.SubsampleRatio {
				case image.YCbCrSubsampleRatio444, image.YCbCrSubsampleRatio440:
					ic = yBase + (x - img.Rect.Min.X)
				case image.YCbCrSubsampleRatio422, image.YCbCrSubsampleRatio420:
					ic = yBase + (x/2 - hx)
				default:
					ic = img.COffset(x, y)
				}

				yy1 := int32(img.Y[iy]) * 0x10101
				cb1 := int32(img.Cb[ic]) - 128
				cr1 := int32(img.Cr[ic]) - 128

				r := yy1 + 91881*cr1
				if uint32(r)&0xff000000 == 0 { //nolint:gosec
					r >>= 16
				} else {
					r = ^(r >> 31)
				}

				g := yy1 - 22554*cb1 - 46802*cr1
				if uint32(g)&0xff000000 == 0 { //nolint:gosec
					g >>= 16
				} else {
					g = ^(g >> 31)
				}

				b := yy1 + 116130*cb1
				if uint32(b)&0xff000000 == 0 { //nolint:gosec
					b >>= 16
				} else {
					b = ^(b >> 31)
				}

				d := dst[j : j+4 : j+4]
				d[0] = uint8(r) // #nosec G115 -- Overflow desired.
				d[1] = uint8(g) // #nosec G115 -- Overflow desired.
				d[2] = uint8(b) // #nosec G115 -- Overflow desired.
				d[3] = 0xff

				iy++
				j += 4
			}
		}

	case *image.Paletted:
		j := 0
		for y := y1; y < y2; y++ {
			i := y*img.Stride + x1
			for x := x1; x < x2; x++ {
				c := s.palette[img.Pix[i]]
				d := dst[j : j+4 : j+4]
				d[0] = c.R
				d[1] = c.G
				d[2] = c.B
				d[3] = c.A
				j += 4
				i++
			}
		}

	default:
		j := 0
		b := s.image.Bounds()
		x1 += b.Min.X
		x2 += b.Min.X
		y1 += b.Min.Y
		y2 += b.Min.Y
		for y := y1; y < y2; y++ {
			for x := x1; x < x2; x++ {
				r16, g16, b16, a16 := s.image.At(x, y).RGBA()
				d := dst[j : j+4 : j+4]
				switch a16 {
				case 0xffff:
					d[0] = uint8(r16 >> 8) // #nosec G115 -- Overflow desired.
					d[1] = uint8(g16 >> 8) // #nosec G115 -- Overflow desired.
					d[2] = uint8(b16 >> 8) // #nosec G115 -- Overflow desired.
					d[3] = 0xff
				case 0:
					d[0] = 0
					d[1] = 0
					d[2] = 0
					d[3] = 0
				default:
					d[0] = uint8(((r16 * 0xffff) / a16) >> 8) // #nosec G115 -- Overflow desired.
					d[1] = uint8(((g16 * 0xffff) / a16) >> 8) // #nosec G115 -- Overflow desired.
					d[2] = uint8(((b16 * 0xffff) / a16) >> 8) // #nosec G115 -- Overflow desired.
					d[3] = uint8(a16 >> 8)                    // #nosec G115 -- Overflow desired.
				}
				j += 4
			}
		}
	}
}

// reverse reverses the data
// in contained pixel slice.
func reverse(pix []uint8) {
	if len(pix) <= 4 {
		return
	}
	i := 0
	j := len(pix) - 4
	for i < j {
		pi := pix[i : i+4 : i+4]
		pj := pix[j : j+4 : j+4]
		pi[0], pj[0] = pj[0], pi[0]
		pi[1], pj[1] = pj[1], pi[1]
		pi[2], pj[2] = pj[2], pi[2]
		pi[3], pj[3] = pj[3], pi[3]
		i += 4
		j -= 4
	}
}

// clampFloat rounds and clamps float64 value to fit into uint8.
func clampFloat(x float64) uint8 {
	v := int64(x + 0.5)
	if v > 255 {
		return 255
	}
	if v > 0 {
		return uint8(v) // #nosec G115 -- Just checked.
	}
	return 0
}

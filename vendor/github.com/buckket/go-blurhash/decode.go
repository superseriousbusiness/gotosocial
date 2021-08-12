package blurhash

import (
	"fmt"
	"github.com/buckket/go-blurhash/base83"
	"image"
	"image/color"
	"math"
)

// An InvalidHashError occurs when the given hash is either too short or the length does not match its size flag.
type InvalidHashError string

func (e InvalidHashError) Error() string {
	return fmt.Sprintf("blurhash: %s", string(e))
}

// Components decodes and returns the number of x and y components in the given BlurHash.
func Components(hash string) (xComponents, yComponents int, err error) {
	if len(hash) < 6 {
		return 0, 0, InvalidHashError("hash is invalid (too short)")
	}

	sizeFlag, err := base83.Decode(string(hash[0]))
	if err != nil {
		return 0, 0, err
	}

	yComponents = (sizeFlag / 9) + 1
	xComponents = (sizeFlag % 9) + 1

	if len(hash) != 4+2*xComponents*yComponents {
		return 0, 0, InvalidHashError("hash is invalid (length mismatch)")
	}

	return xComponents, yComponents, nil
}

// Decode generates an image of the given BlurHash with a size of width and height.
// Punch is a multiplier that adjusts the contrast of the resulting image.
func Decode(hash string, width, height, punch int) (image.Image, error) {
	xComp, yComp, err := Components(hash)
	if err != nil {
		return nil, err
	}

	quantisedMaximumValue, err := base83.Decode(string(hash[1]))
	if err != nil {
		return nil, err
	}
	maximumValue := (float64(quantisedMaximumValue) + 1) / 166

	if punch == 0 {
		punch = 1
	}

	colors := make([][3]float64, xComp*yComp)

	for i := range colors {
		if i == 0 {
			value, err := base83.Decode(hash[2:6])
			if err != nil {
				return nil, err
			}
			colors[i] = decodeDC(value)
		} else {
			value, err := base83.Decode(hash[4+i*2 : 6+i*2])
			if err != nil {
				return nil, err
			}
			colors[i] = decodeAC(value, maximumValue*float64(punch))
		}
	}

	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var r, g, b float64
			for j := 0; j < yComp; j++ {
				for i := 0; i < xComp; i++ {
					basis := math.Cos(math.Pi*float64(x)*float64(i)/float64(width)) *
						math.Cos(math.Pi*float64(y)*float64(j)/float64(height))
					pcolor := colors[i+j*xComp]
					r += pcolor[0] * basis
					g += pcolor[1] * basis
					b += pcolor[2] * basis
				}
			}
			img.SetNRGBA(x, y, color.NRGBA{R: uint8(linearTosRGB(r)), G: uint8(linearTosRGB(g)), B: uint8(linearTosRGB(b)), A: 255})
		}
	}

	return img, nil
}

func decodeDC(value int) [3]float64 {
	return [3]float64{sRGBToLinear(value >> 16), sRGBToLinear(value >> 8 & 255), sRGBToLinear(value & 255)}
}

func decodeAC(value int, maximumValue float64) [3]float64 {
	quantR := math.Floor(float64(value) / (19 * 19))
	quantG := math.Mod(math.Floor(float64(value)/19), 19)
	quantB := math.Mod(float64(value), 19)
	sp := func(quant float64) float64 {
		return signPow((quant-9)/9, 2.0) * maximumValue
	}
	return [3]float64{sp(quantR), sp(quantG), sp(quantB)}
}

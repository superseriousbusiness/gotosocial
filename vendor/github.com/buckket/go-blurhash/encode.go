package blurhash

import (
	"fmt"
	"github.com/buckket/go-blurhash/base83"
	"image"
	"math"
	"strings"
)

func init() {
	initLinearTable(channelToLinear[:])
}

var channelToLinear [256]float64

func initLinearTable(table []float64) {
	for i := range table {
		channelToLinear[i] = sRGBToLinear(i)
	}
}

// An InvalidParameterError occurs when an invalid argument is passed to either the Decode or Encode function.
type InvalidParameterError struct {
	Value     int
	Parameter string
}

func (e InvalidParameterError) Error() string {
	return fmt.Sprintf("blurhash: %sComponents (%d) must be element of [1-9]", e.Parameter, e.Value)
}

// An EncodingError represents an error that occurred during the encoding of the given value.
// This most likely means that your input image is invalid and can not be processed.
type EncodingError string

func (e EncodingError) Error() string {
	return fmt.Sprintf("blurhash: %s", string(e))
}

// Encode calculates the Blurhash for an image using the given x and y component counts.
// The x and y components have to be between 1 and 9 respectively.
// The image must be of image.Image type.
func Encode(xComponents int, yComponents int, rgba image.Image) (string, error) {
	if xComponents < 1 || xComponents > 9 {
		return "", InvalidParameterError{xComponents, "x"}
	}
	if yComponents < 1 || yComponents > 9 {
		return "", InvalidParameterError{yComponents, "y"}
	}

	var blurhash strings.Builder
	blurhash.Grow(4 + 2*xComponents*yComponents)

	// Size Flag
	str, err := base83.Encode((xComponents-1)+(yComponents-1)*9, 1)
	if err != nil {
		return "", EncodingError("could not encode size flag")
	}
	blurhash.WriteString(str)

	factors := make([]float64, yComponents*xComponents*3)
	multiplyBasisFunction(rgba, factors, xComponents, yComponents)

	var maximumValue float64
	var quantisedMaximumValue int
	var acCount = xComponents*yComponents - 1
	if acCount > 0 {
		var actualMaximumValue float64
		for i := 0; i < acCount*3; i++ {
			actualMaximumValue = math.Max(math.Abs(factors[i+3]), actualMaximumValue)
		}
		quantisedMaximumValue = int(math.Max(0, math.Min(82, math.Floor(actualMaximumValue*166-0.5))))
		maximumValue = (float64(quantisedMaximumValue) + 1) / 166
	} else {
		maximumValue = 1
	}

	// Quantised max AC component
	str, err = base83.Encode(quantisedMaximumValue, 1)
	if err != nil {
		return "", EncodingError("could not encode quantised max AC component")
	}
	blurhash.WriteString(str)

	// DC value
	str, err = base83.Encode(encodeDC(factors[0], factors[1], factors[2]), 4)
	if err != nil {
		return "", EncodingError("could not encode DC value")
	}
	blurhash.WriteString(str)

	// AC values
	for i := 0; i < acCount; i++ {
		str, err = base83.Encode(encodeAC(factors[3+(i*3+0)], factors[3+(i*3+1)], factors[3+(i*3+2)], maximumValue), 2)
		if err != nil {
			return "", EncodingError("could not encode AC value")
		}
		blurhash.WriteString(str)
	}

	if blurhash.Len() != 4+2*xComponents*yComponents {
		return "", EncodingError("hash does not match expected size")
	}

	return blurhash.String(), nil
}

func multiplyBasisFunction(rgba image.Image, factors []float64, xComponents int, yComponents int) {
	height := rgba.Bounds().Max.Y
	width := rgba.Bounds().Max.X

	xvalues := make([][]float64, xComponents)
	for xComponent := 0; xComponent < xComponents; xComponent++ {
		xvalues[xComponent] = make([]float64, width)
		for x := 0; x < width; x++ {
			xvalues[xComponent][x] = math.Cos(math.Pi * float64(xComponent) * float64(x) / float64(width))
		}
	}

	yvalues := make([][]float64, yComponents)
	for yComponent := 0; yComponent < yComponents; yComponent++ {
		yvalues[yComponent] = make([]float64, height)
		for y := 0; y < height; y++ {
			yvalues[yComponent][y] = math.Cos(math.Pi * float64(yComponent) * float64(y) / float64(height))
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rt, gt, bt, _ := rgba.At(x, y).RGBA()
			lr := channelToLinear[rt>>8]
			lg := channelToLinear[gt>>8]
			lb := channelToLinear[bt>>8]

			for yc := 0; yc < yComponents; yc++ {
				for xc := 0; xc < xComponents; xc++ {

					scale := 1 / float64(width*height)

					if xc != 0 || yc != 0 {
						scale = 2 / float64(width*height)
					}

					basis := xvalues[xc][x] * yvalues[yc][y]
					factors[0+xc*3+yc*3*xComponents] += lr * basis * scale
					factors[1+xc*3+yc*3*xComponents] += lg * basis * scale
					factors[2+xc*3+yc*3*xComponents] += lb * basis * scale
				}
			}
		}
	}
}

func encodeDC(r, g, b float64) int {
	return (linearTosRGB(r) << 16) + (linearTosRGB(g) << 8) + linearTosRGB(b)
}

func encodeAC(r, g, b, maximumValue float64) int {
	quant := func(f float64) int {
		return int(math.Max(0, math.Min(18, math.Floor(signPow(f/maximumValue, 0.5)*9+9.5))))
	}
	return quant(r)*19*19 + quant(g)*19 + quant(b)
}

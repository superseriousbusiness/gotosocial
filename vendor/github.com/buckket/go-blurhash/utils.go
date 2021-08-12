package blurhash

import "math"

func linearTosRGB(value float64) int {
	v := math.Max(0, math.Min(1, value))
	if v <= 0.0031308 {
		return int(v*12.92*255 + 0.5)
	}
	return int((1.055*math.Pow(v, 1/2.4)-0.055)*255 + 0.5)
}

func sRGBToLinear(value int) float64 {
	v := float64(value) / 255
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func signPow(value, exp float64) float64 {
	return math.Copysign(math.Pow(math.Abs(value), exp), value)
}

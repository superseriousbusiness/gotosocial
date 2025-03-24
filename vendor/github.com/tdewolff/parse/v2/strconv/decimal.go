package strconv

import (
	"math"
)

// ParseDecimal parses number of the format 1.2
func ParseDecimal(b []byte) (float64, int) {
	// float64 has up to 17 significant decimal digits and an exponent in [-1022,1023]
	i := 0
	sign := 1.0
	if 0 < len(b) && b[0] == '-' {
		sign = -1.0
		i++
	}

	start := -1
	dot := -1
	n := uint64(0)
	for ; i < len(b); i++ {
		// parse up to 18 significant digits (with dot will be 17) ignoring zeros before/after
		c := b[i]
		if '0' <= c && c <= '9' {
			if start == -1 {
				if '1' <= c && c <= '9' {
					n = uint64(c - '0')
					start = i
				}
			} else if i-start < 18 {
				n *= 10
				n += uint64(c - '0')
			}
		} else if c == '.' {
			if dot != -1 {
				break
			}
			dot = i
		} else {
			break
		}
	}
	if i == 1 && dot == 0 {
		return 0.0, 0 // only dot
	} else if start == -1 {
		return 0.0, i // only zeros and dot
	} else if dot == -1 {
		dot = i
	}

	exp := (dot - start) - LenUint(n)
	if dot < start {
		exp++
	}
	if 1023 < exp {
		if sign == 1.0 {
			return math.Inf(1), i
		} else {
			return math.Inf(-1), i
		}
	} else if exp < -1022 {
		return 0.0, i
	}

	f := sign * float64(n)
	if 0 <= exp && exp < 23 {
		return f * float64pow10[exp], i
	} else if 23 < exp && exp < 0 {
		return f / float64pow10[exp], i
	}
	return f * math.Pow10(exp), i
}

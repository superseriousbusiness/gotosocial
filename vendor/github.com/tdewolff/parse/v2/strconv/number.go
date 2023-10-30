package strconv

import (
	"math"
	"unicode/utf8"
)

// ParseNumber parses a byte-slice and returns the number it represents and the amount of decimals.
// If an invalid character is encountered, it will stop there.
func ParseNumber(b []byte, groupSym rune, decSym rune) (int64, int, int) {
	n, dec := 0, 0
	sign := int64(1)
	price := int64(0)
	hasDecimals := false
	if 0 < len(b) && b[0] == '-' {
		sign = -1
		n++
	}
	for n < len(b) {
		if '0' <= b[n] && b[n] <= '9' {
			digit := sign * int64(b[n]-'0')
			if sign == 1 && (math.MaxInt64/10 < price || math.MaxInt64-digit < price*10) {
				break
			} else if sign == -1 && (price < math.MinInt64/10 || price*10 < math.MinInt64-digit) {
				break
			}
			price *= 10
			price += digit
			if hasDecimals {
				dec++
			}
			n++
		} else if r, size := utf8.DecodeRune(b[n:]); !hasDecimals && (r == groupSym || r == decSym) {
			if r == decSym {
				hasDecimals = true
			}
			n += size
		} else {
			break
		}
	}
	return price, dec, n
}

// AppendNumber will append an int64 formatted as a number with the given number of decimal digits.
func AppendNumber(b []byte, price int64, dec int, groupSize int, groupSym rune, decSym rune) []byte {
	if dec < 0 {
		dec = 0
	}
	if utf8.RuneLen(groupSym) == -1 {
		groupSym = '.'
	}
	if utf8.RuneLen(decSym) == -1 {
		decSym = ','
	}

	sign := int64(1)
	if price < 0 {
		sign = -1
	}

	// calculate size
	n := LenInt(price)
	if dec < n && 0 < groupSize && groupSym != 0 {
		n += utf8.RuneLen(groupSym) * (n - dec - 1) / groupSize
	}
	if 0 < dec {
		if n <= dec {
			n = 1 + dec // zero and decimals
		}
		n += utf8.RuneLen(decSym)
	}
	if sign == -1 {
		n++
	}

	// resize byte slice
	i := len(b)
	if cap(b) < i+n {
		b = append(b, make([]byte, n)...)
	} else {
		b = b[:i+n]
	}

	// print fractional-part
	i += n - 1
	if 0 < dec {
		for 0 < dec {
			c := byte(sign*(price%10)) + '0'
			price /= 10
			b[i] = c
			dec--
			i--
		}
		i -= utf8.RuneLen(decSym)
		utf8.EncodeRune(b[i+1:], decSym)
	}

	// print integer-part
	if price == 0 {
		b[i] = '0'
		if sign == -1 {
			b[i-1] = '-'
		}
		return b
	}
	j := 0
	for price != 0 {
		if 0 < groupSize && groupSym != 0 && 0 < j && j%groupSize == 0 {
			i -= utf8.RuneLen(groupSym)
			utf8.EncodeRune(b[i+1:], groupSym)
		}

		c := byte(sign*(price%10)) + '0'
		price /= 10
		b[i] = c
		i--
		j++
	}

	if sign == -1 {
		b[i] = '-'
	}
	return b
}

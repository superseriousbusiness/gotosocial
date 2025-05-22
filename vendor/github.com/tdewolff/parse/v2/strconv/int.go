package strconv

import (
	"math"
)

// ParseInt parses a byte-slice and returns the integer it represents.
// If an invalid character is encountered, it will stop there.
func ParseInt(b []byte) (int64, int) {
	i := 0
	neg := false
	if len(b) > 0 && (b[0] == '+' || b[0] == '-') {
		neg = b[0] == '-'
		i++
	}
	start := i
	n := uint64(0)
	for i < len(b) {
		c := b[i]
		if '0' <= c && c <= '9' {
			if uint64(-math.MinInt64)/10 < n || uint64(-math.MinInt64)-uint64(c-'0') < n*10 {
				return 0, 0
			}
			n *= 10
			n += uint64(c - '0')
		} else {
			break
		}
		i++
	}
	if i == start {
		return 0, 0
	}
	if !neg && uint64(math.MaxInt64) < n {
		return 0, 0
	} else if neg {
		return -int64(n), i
	}
	return int64(n), i
}

// ParseUint parses a byte-slice and returns the integer it represents.
// If an invalid character is encountered, it will stop there.
func ParseUint(b []byte) (uint64, int) {
	i := 0
	n := uint64(0)
	for i < len(b) {
		c := b[i]
		if '0' <= c && c <= '9' {
			if math.MaxUint64/10 < n || math.MaxUint64-uint64(c-'0') < n*10 {
				return 0, 0
			}
			n *= 10
			n += uint64(c - '0')
		} else {
			break
		}
		i++
	}
	return n, i
}

// AppendInt will append an int64.
func AppendInt(b []byte, num int64) []byte {
	if num == 0 {
		return append(b, '0')
	} else if num == -9223372036854775808 {
		return append(b, "-9223372036854775808"...)
	}

	// resize byte slice
	i, n := len(b), LenInt(num)
	if cap(b) < i+n {
		b = append(b, make([]byte, n)...)
	} else {
		b = b[:i+n]
	}

	// print sign
	if num < 0 {
		num = -num
		b[i] = '-'
	}
	i += n - 1

	// print number
	for num != 0 {
		b[i] = byte(num%10) + '0'
		num /= 10
		i--
	}
	return b
}

// LenInt returns the written length of an integer.
func LenInt(i int64) int {
	if i < 0 {
		if i == -9223372036854775808 {
			return 20
		}
		return 1 + LenUint(uint64(-i))
	}
	return LenUint(uint64(i))
}

func LenUint(i uint64) int {
	switch {
	case i < 10:
		return 1
	case i < 100:
		return 2
	case i < 1000:
		return 3
	case i < 10000:
		return 4
	case i < 100000:
		return 5
	case i < 1000000:
		return 6
	case i < 10000000:
		return 7
	case i < 100000000:
		return 8
	case i < 1000000000:
		return 9
	case i < 10000000000:
		return 10
	case i < 100000000000:
		return 11
	case i < 1000000000000:
		return 12
	case i < 10000000000000:
		return 13
	case i < 100000000000000:
		return 14
	case i < 1000000000000000:
		return 15
	case i < 10000000000000000:
		return 16
	case i < 100000000000000000:
		return 17
	case i < 1000000000000000000:
		return 18
	case i < 10000000000000000000:
		return 19
	}
	return 20
}

var int64pow10 = []int64{
	1, 10, 100, 1000, 10000, 100000, 1000000, 10000000, 100000000, 1000000000, 10000000000, 100000000000, 1000000000000, 10000000000000, 100000000000000, 1000000000000000, 10000000000000000, 100000000000000000, 1000000000000000000,
}

package format

import "strconv"

// can32bitInt returns whether it's possible for 's' to contain an int on 32bit platforms.
func can32bitInt(s string) bool {
	return strconv.IntSize == 32 && (0 < len(s) && len(s) < 10)
}

// can64bitInt returns whether it's possible for 's' to contain an int on 64bit platforms.
func can64bitInt(s string) bool {
	return strconv.IntSize == 64 && (0 < len(s) && len(s) < 19)
}

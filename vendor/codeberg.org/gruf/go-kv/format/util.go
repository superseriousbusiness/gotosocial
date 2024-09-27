package format

import (
	"strings"
	"unsafe"
)

const (
	// SingleTermLine: beyond a certain length of string, all of the
	// extra checks to handle quoting/not-quoting add a significant
	// amount of extra processing time. Quoting in this manner only really
	// effects readability on a single line, so a max string length that
	// encompasses the maximum number of columns on *most* terminals was
	// selected. This was chosen using the metric that 1080p is one of the
	// most common display resolutions, and that a relatively small font size
	// of 7 requires 223 columns. So 256 should be >= $COLUMNS (fullscreen)
	// in 99% of usecases (these figures all pulled out of my ass).
	SingleTermLine = 256
)

// IsSafeASCII checks whether string is printable (i.e. non-control char) ASCII text.
func IsSafeASCII(str string) bool {
	for _, r := range str {
		if (r < ' ' && r != '\t') ||
			r >= 0x7f {
			return false
		}
	}
	return true
}

// ContainsSpaceOrTab checks if "s" contains space or tabs. EXPECTS ASCII.
func ContainsSpaceOrTab(s string) bool {
	if i := strings.IndexByte(s, ' '); i >= 0 {
		return true // note using indexbyte as it is ASM.
	} else if i := strings.IndexByte(s, '\t'); i >= 0 {
		return true
	}
	return false
}

// ContainsDoubleQuote checks if "s" contains a double quote. EXPECTS ASCII.
func ContainsDoubleQuote(s string) bool {
	return (strings.IndexByte(s, '"') >= 0)
}

// AppendEscape will append 's' to 'dst' and escape any double quotes. EXPECTS ASCII.
func AppendEscape(dst []byte, str string) []byte {
	for i := range str {
		switch str[i] {
		case '\\':
			// Append delimited '\'
			dst = append(dst, '\\', '\\')

		case '"':
			// Append delimited '"'
			dst = append(dst, '\\', '"')
		default:
			// Append char as-is
			dst = append(dst, str[i])
		}
	}
	return dst
}

// Byte2Str returns 'c' as a string, escaping if necessary.
func Byte2Str(c byte) string {
	switch c {
	case '\a':
		return `\a`
	case '\b':
		return `\b`
	case '\f':
		return `\f`
	case '\n':
		return `\n`
	case '\r':
		return `\r`
	case '\t':
		return `\t`
	case '\v':
		return `\v`
	case '\\':
		return `\\`
	default:
		if c < ' ' {
			const hex = "0123456789abcdef"
			return `\x` +
				string(hex[c>>4]) +
				string(hex[c&0xF])
		}
		return string(c)
	}
}

// isNil will safely check if 'v' is nil without dealing with weird Go interface nil bullshit.
func isNil(i interface{}) bool {
	type eface struct{ _type, data unsafe.Pointer }    //nolint
	return (*(*eface)(unsafe.Pointer(&i))).data == nil //nolint
}

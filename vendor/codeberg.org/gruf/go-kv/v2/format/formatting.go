package format

import (
	"strings"
	"unicode"
	"unicode/utf8"
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

// AppendEscape will append 's' to 'buf' and escape any double quotes. EXPECTS ASCII.
func AppendEscape(buf []byte, str string) []byte {
	for i := range str {
		switch str[i] {
		case '\\':
			// Append delimited '\'
			buf = append(buf, '\\', '\\')

		case '"':
			// Append delimited '"'
			buf = append(buf, '\\', '"')
		default:
			// Append char as-is
			buf = append(buf, str[i])
		}
	}
	return buf
}

const hex = "0123456789abcdef"

// AppendEscapeByte ...
func AppendEscapeByte(buf []byte, c byte) []byte {
	switch c {
	case '\a':
		return append(buf, `\a`...)
	case '\b':
		return append(buf, `\b`...)
	case '\f':
		return append(buf, `\f`...)
	case '\n':
		return append(buf, `\n`...)
	case '\r':
		return append(buf, `\r`...)
	case '\t':
		return append(buf, `\t`...)
	case '\v':
		return append(buf, `\v`...)
	case '\\':
		return append(buf, `\\`...)
	default:
		if c < ' ' {
			return append(buf, '\\', 'x', hex[c>>4], hex[c&0xF])
		}
		return append(buf, c)
	}
}

// AppendQuoteByte ...
func AppendQuoteByte(buf []byte, c byte) []byte {
	if c == '\'' {
		return append(buf, `'\''`...)
	}
	buf = append(buf, '\'')
	buf = AppendEscapeByte(buf, c)
	buf = append(buf, '\'')
	return buf
}

// AppendEscapeRune ...
func AppendEscapeRune(buf []byte, r rune) []byte {
	if unicode.IsPrint(r) {
		return utf8.AppendRune(buf, r)
	}
	switch r {
	case '\a':
		return append(buf, `\a`...)
	case '\b':
		return append(buf, `\b`...)
	case '\f':
		return append(buf, `\f`...)
	case '\n':
		return append(buf, `\n`...)
	case '\r':
		return append(buf, `\r`...)
	case '\t':
		return append(buf, `\t`...)
	case '\v':
		return append(buf, `\v`...)
	case '\\':
		return append(buf, `\\`...)
	default:
		switch {
		case r < ' ' || r == 0x7f:
			buf = append(buf, `\x`...)
			buf = append(buf, hex[byte(r)>>4])
			buf = append(buf, hex[byte(r)&0xF])
		case !utf8.ValidRune(r):
			r = 0xFFFD
			fallthrough
		case r < 0x10000:
			buf = append(buf, `\u`...)
			buf = append(buf,
				hex[r>>uint(12)&0xF],
				hex[r>>uint(8)&0xF],
				hex[r>>uint(4)&0xF],
				hex[r>>uint(0)&0xF],
			)
		default:
			buf = append(buf, `\U`...)
			buf = append(buf,
				hex[r>>uint(28)&0xF],
				hex[r>>uint(24)&0xF],
				hex[r>>uint(20)&0xF],
				hex[r>>uint(16)&0xF],
				hex[r>>uint(12)&0xF],
				hex[r>>uint(8)&0xF],
				hex[r>>uint(4)&0xF],
				hex[r>>uint(0)&0xF],
			)
		}
	}
	return buf
}

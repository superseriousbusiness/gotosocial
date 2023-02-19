package kv

import (
	"strconv"
	"strings"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-kv/format"
)

// AppendQuote will append (and escape/quote where necessary) a formatted field string.
func AppendQuote(buf *byteutil.Buffer, str string) {
	switch {
	case len(str) == 0:
		// Append empty quotes.
		buf.B = append(buf.B, `""`...)
		return

	case len(str) == 1:
		// Append quote single byte.
		appendQuoteByte(buf, str[0])
		return

	case len(str) > format.SingleTermLine || !format.IsSafeASCII(str):
		// Long line or contains non-ascii chars.
		buf.B = strconv.AppendQuote(buf.B, str)
		return

	case !isQuoted(str):
		// Single/double quoted already.

		// Get space / tab indices (if any).
		s := strings.IndexByte(str, ' ')
		t := strings.IndexByte(str, '\t')

		// Find first whitespace.
		sp0 := smallest(s, t)
		if sp0 < 0 {
			break
		}

		// Check if str is enclosed by braces.
		// (but without any key-value separator).
		if (enclosedBy(str, sp0, '{', '}') ||
			enclosedBy(str, sp0, '[', ']') ||
			enclosedBy(str, sp0, '(', ')')) &&
			strings.IndexByte(str, '=') < 0 {
			break
		}

		if format.ContainsDoubleQuote(str) {
			// Contains double quote, double quote
			// and append escaped existing.
			buf.B = append(buf.B, '"')
			buf.B = format.AppendEscape(buf.B, str)
			buf.B = append(buf.B, '"')
			return
		}

		// Quote un-enclosed spaces.
		buf.B = append(buf.B, '"')
		buf.B = append(buf.B, str...)
		buf.B = append(buf.B, '"')
		return
	}

	// Double quoted, enclosed in braces, or
	// literally anything else: append as-is.
	buf.B = append(buf.B, str...)
	return
}

// appendEscapeByte will append byte to buffer, quoting and escaping where necessary.
func appendQuoteByte(buf *byteutil.Buffer, c byte) {
	switch c {
	// Double quote space.
	case ' ':
		buf.B = append(buf.B, '"', c, '"')

	// Escape + double quote.
	case '\a':
		buf.B = append(buf.B, '"', '\\', 'a', '"')
	case '\b':
		buf.B = append(buf.B, '"', '\\', 'b', '"')
	case '\f':
		buf.B = append(buf.B, '"', '\\', 'f', '"')
	case '\n':
		buf.B = append(buf.B, '"', '\\', 'n', '"')
	case '\r':
		buf.B = append(buf.B, '"', '\\', 'r', '"')
	case '\t':
		buf.B = append(buf.B, '"', '\\', 't', '"')
	case '\v':
		buf.B = append(buf.B, '"', '\\', 'v', '"')

	// Append as-is.
	default:
		buf.B = append(buf.B, c)
	}
}

// isQuoted checks if string is single or double quoted.
func isQuoted(str string) bool {
	return (str[0] == '"' && str[len(str)-1] == '"') ||
		(str[0] == '\'' && str[len(str)-1] == '\'')
}

// smallest attempts to return the smallest positive value of those given.
func smallest(i1, i2 int) int {
	if i1 >= 0 && (i2 < 0 || i1 < i2) {
		return i1
	}
	return i2
}

// enclosedBy will check if given string is enclosed by end, and at least non-whitespace up to start.
func enclosedBy(str string, sp int, start, end byte) bool {
	// Check for ending char in string.
	if str[len(str)-1] != end {
		return false
	}

	// Check for starting char in string.
	i := strings.IndexByte(str, start)
	if i < 0 {
		return false
	}

	// Check before space.
	return i < sp
}

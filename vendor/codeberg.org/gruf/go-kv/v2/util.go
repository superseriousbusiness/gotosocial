package kv

import (
	"strconv"
	"strings"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-kv/v2/format"
)

// AppendQuoteString will append (and escape/quote where necessary) a field string.
func AppendQuoteString(buf *byteutil.Buffer, str string) {
	switch {
	case len(str) == 0:
		// Append empty quotes.
		buf.B = append(buf.B, `""`...)
		return

	case len(str) == 1:
		// Append escaped single byte.
		buf.B = format.AppendEscapeByte(buf.B, str[0])
		return

	case len(str) > format.SingleTermLine || !format.IsSafeASCII(str):
		// Long line or contains non-ascii chars.
		buf.B = strconv.AppendQuote(buf.B, str)
		return

	case !isQuoted(str):
		// Not single/double quoted already.

		if format.ContainsSpaceOrTab(str) {
			// Quote un-enclosed spaces.
			buf.B = append(buf.B, '"')
			buf.B = append(buf.B, str...)
			buf.B = append(buf.B, '"')
			return
		}

		if format.ContainsDoubleQuote(str) {
			// Contains double quote, double quote
			// and append escaped existing.
			buf.B = append(buf.B, '"')
			buf.B = format.AppendEscape(buf.B, str)
			buf.B = append(buf.B, '"')
			return
		}
	}

	// Double quoted, enclosed in braces, or
	// literally anything else: append as-is.
	buf.B = append(buf.B, str...)
	return
}

// AppendQuoteValue will append (and escape/quote where necessary) a formatted value string.
func AppendQuoteValue(buf *byteutil.Buffer, str string) {
	switch {
	case len(str) == 0:
		// Append empty quotes.
		buf.B = append(buf.B, `""`...)
		return

	case len(str) == 1:
		// Append quoted single byte.
		buf.B = format.AppendQuoteByte(buf.B, str[0])
		return

	case len(str) > format.SingleTermLine || !format.IsSafeASCII(str):
		// Long line or contains non-ascii chars.
		buf.B = strconv.AppendQuote(buf.B, str)
		return

	case !isQuoted(str):
		// Not single/double quoted already.

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

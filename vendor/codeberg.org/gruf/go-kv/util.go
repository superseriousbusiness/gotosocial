package kv

import (
	"strconv"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-kv/format"
)

// AppendQuoteKey will append and escape/quote a formatted key string.
func AppendQuoteKey(buf *byteutil.Buffer, str string) {
	switch {
	case len(str) > format.SingleTermLine || !strconv.CanBackquote(str):
		// Append quoted and escaped string
		buf.B = strconv.AppendQuote(buf.B, str)
	case format.ContainsDoubleQuote(str):
		// Double quote and escape string
		buf.B = append(buf.B, '"')
		buf.B = format.AppendEscape(buf.B, str)
		buf.B = append(buf.B, '"')
	case len(str) < 1 || format.ContainsSpaceOrTab(str):
		// Double quote this string as-is
		buf.WriteString(`"` + str + `"`)
	default:
		// Append string as-is
		buf.WriteString(str)
	}
}

// AppendQuoteValue will append and escape/quote a formatted value string.
func AppendQuoteValue(buf *byteutil.Buffer, str string) {
	switch {
	case len(str) > format.SingleTermLine || !strconv.CanBackquote(str):
		// Append quoted and escaped string
		buf.B = strconv.AppendQuote(buf.B, str)
		return
	case !doubleQuoted(str):
		if format.ContainsDoubleQuote(str) {
			// Double quote and escape string
			buf.B = append(buf.B, '"')
			buf.B = format.AppendEscape(buf.B, str)
			buf.B = append(buf.B, '"')
			return
		} else if format.ContainsSpaceOrTab(str) {
			// Double quote this string as-is
			buf.WriteString(`"` + str + `"`)
			return
		}
	}

	// Append string as-is
	buf.WriteString(str)
}

// doubleQuoted will return whether 'str' is double quoted.
func doubleQuoted(str string) bool {
	if len(str) < 2 ||
		str[0] != '"' || str[len(str)-1] != '"' {
		return false
	}
	var delim bool
	for i := len(str) - 2; i >= 0; i-- {
		switch str[i] {
		case '\\':
			delim = !delim
		default:
			return !delim
		}
	}
	return !delim
}

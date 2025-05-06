package split

import (
	"strconv"
	"strings"
)

// singleTermLine: beyond a certain length of string, all of the
// extra checks to handle quoting/not-quoting add a significant
// amount of extra processing time. Quoting in this manner only really
// effects readability on a single line, so a max string length that
// encompasses the maximum number of columns on *most* terminals was
// selected. This was chosen using the metric that 1080p is one of the
// most common display resolutions, and that a relatively small font size
// of 7 requires ~ 223 columns. So 256 should be >= $COLUMNS (fullscreen)
// in 99% of usecases (these figures all pulled out of my ass).
const singleTermLine = 256

// appendQuote will append 'str' to 'buf', double quoting and escaping if needed.
func appendQuote(buf []byte, str string) []byte {
	switch {
	case len(str) > singleTermLine || !strconv.CanBackquote(str):
		// Append quoted and escaped string
		return strconv.AppendQuote(buf, str)

	case (strings.IndexByte(str, '"') != -1):
		// Double quote and escape string
		buf = append(buf, '"')
		buf = appendEscape(buf, str)
		buf = append(buf, '"')
		return buf

	case (strings.IndexByte(str, ',') != -1):
		// Double quote this string as-is
		buf = append(buf, '"')
		buf = append(buf, str...)
		buf = append(buf, '"')
		return buf

	default:
		// Append string as-is
		return append(buf, str...)
	}
}

// appendEscape will append 'str' to 'buf' and escape any double quotes.
func appendEscape(buf []byte, str string) []byte {
	var delim bool
	for i := range str {
		switch {
		case str[i] == '\\' && !delim:
			// Set delim flag
			delim = true

		case str[i] == '"' && !delim:
			// Append escaped double quote
			buf = append(buf, `\"`...)

		case delim:
			// Append skipped slash
			buf = append(buf, `\`...)
			delim = false
			fallthrough

		default:
			// Append char as-is
			buf = append(buf, str[i])
		}
	}
	return buf
}

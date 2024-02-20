package codegen

import (
	"strings"
)

const (
	max_width         = 80
	tab_assumed_width = 8
	replacement       = "\n// "
	httpsScheme       = "https://"
	httpScheme        = "http://"
)

// FormatPackageDocumentation is used to format package-level comments.
func FormatPackageDocumentation(s string) string {
	return insertNewlines(s)
}

// insertNewlines is used to trade a space character for a newline character
// in order to keep a string's visual width under a certain amount.
func insertNewlines(s string) string {
	s = strings.Replace(s, "\n", replacement, -1)
	return insertNewlinesEvery(s, max_width)
}

// insertNewlinesIndented is used to trade a space character for a newline
// character in order to keep a string's visual width under a certain amount. It
// assumes that the string will be indented once, and accounts for it in the
// final result.
func insertNewlinesIndented(s string) string {
	return insertNewlinesEvery(s, max_width-tab_assumed_width)
}

// insertNewlinesEvery inserts a newline every n characters maximum, unless
// there is a very long run-on word.
func insertNewlinesEvery(s string, n int) string {
	since := 0
	found := -1
	diff := len(replacement) - 1
	i := 0
	for i < len(s) {
		if s[i] == ' ' && (since < n || found < 0) {
			found = i
		} else if s[i] == '\n' {
			// Reset, found a newline
			since = 0
			found = -1
		} else if i > len(httpScheme) && s[i-len(httpScheme)+1:i+1] == httpScheme {
			// Reset, let the link just extend annoyingly.
			found = -1
		} else if i > len(httpsScheme) && s[i-len(httpsScheme)+1:i+1] == httpsScheme {
			// Reset, let the link just extend annoyingly.
			found = -1
		}
		if since >= n && found >= 0 {
			// Replace character
			s = s[:found] + replacement + s[found+1:]
			i += diff
			since = i - found
			found = -1
		} else {
			i++
			since++
		}
	}
	return "// " + s
}

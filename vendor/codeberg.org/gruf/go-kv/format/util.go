package format

import (
	"strings"
	"unsafe"
)

// ContainsSpaceOrTab checks if "s" contains space or tabs.
func ContainsSpaceOrTab(s string) bool {
	if i := strings.IndexByte(s, ' '); i != -1 {
		return true // note using indexbyte as it is ASM.
	} else if i := strings.IndexByte(s, '\t'); i != -1 {
		return true
	}
	return false
}

// ContainsDoubleQuote checks if "s" contains a double quote.
func ContainsDoubleQuote(s string) bool {
	return (strings.IndexByte(s, '"') != -1)
}

// AppendEscape will append 's' to 'dst' and escape any double quotes.
func AppendEscape(dst []byte, str string) []byte {
	var delim bool
	for i := range str {
		if str[i] == '\\' && !delim {
			// Set delim flag
			delim = true
			continue
		} else if str[i] == '"' && !delim {
			// Append escaped double quote
			dst = append(dst, `\"`...)
			continue
		} else if delim {
			// Append skipped slash
			dst = append(dst, `\`...)
			delim = false
		}

		// Append char as-is
		dst = append(dst, str[i])
	}
	return dst
}

// isNil will safely check if 'v' is nil without dealing with weird Go interface nil bullshit.
func isNil(i interface{}) bool {
	type eface struct{ _type, data unsafe.Pointer }    //nolint
	return (*(*eface)(unsafe.Pointer(&i))).data == nil //nolint
}

// b2s converts a byteslice to string without allocation.
func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

package split

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Splitter holds onto a byte buffer for use in minimising allocations during SplitFunc().
type Splitter struct{ B []byte }

// SplitFunc will split input string on commas, taking into account string quoting and
// stripping extra whitespace, passing each split to the given function hook.
func (s *Splitter) SplitFunc(str string, fn func(string) error) error {
	for {
		// Reset buffer
		s.B = s.B[0:0]

		// Trim leading space
		str = trimLeadingSpace(str)

		if len(str) < 1 {
			// Reached end
			return nil
		}

		switch {
		// Single / double quoted
		case str[0] == '\'', str[0] == '"':
			// Calculate next string elem
			i := 1 + s.next(str[1:], str[0])
			if i == 0 /* i.e. if .next() returned -1 */ {
				return errors.New("missing end quote")
			}

			// Pass next element to callback func
			if err := fn(string(s.B)); err != nil {
				return err
			}

			// Reslice + trim leading space
			str = trimLeadingSpace(str[i+1:])

			if len(str) < 1 {
				// reached end
				return nil
			}

			if str[0] != ',' {
				// malformed element without comma after quote
				return errors.New("missing comma separator")
			}

			// Skip comma
			str = str[1:]

		// Empty segment
		case str[0] == ',':
			str = str[1:]

		// No quoting
		default:
			// Calculate next string elem
			i := s.next(str, ',')

			switch i {
			// Reached end
			case -1:
				// we know len > 0

				// Pass to callback
				return fn(string(s.B))

			// Empty elem
			case 0:
				str = str[1:]

			// Non-zero elem
			default:
				// Pass next element to callback
				if err := fn(string(s.B)); err != nil {
					return err
				}

				// Skip past eleme
				str = str[i+1:]
			}
		}
	}
}

// next will build the next string element in s.B up to non-delimited instance of c,
// returning number of characters iterated, or -1 if the end of the string was reached.
func (s *Splitter) next(str string, c byte) int {
	var delims int

	// Guarantee buf large enough
	if len(str) > cap(s.B)-len(s.B) {
		nb := make([]byte, 2*cap(s.B)+len(str))
		_ = copy(nb, s.B)
		s.B = nb[:len(s.B)]
	}

	for i := 0; i < len(str); i++ {
		// Increment delims
		if str[i] == '\\' {
			delims++
			continue
		}

		if str[i] == c {
			var count int

			if count = delims / 2; count > 0 {
				// Add backslashes to buffer
				slashes := backslashes(count)
				s.B = append(s.B, slashes...)
			}

			// Reached delim'd char
			if delims-count == 0 {
				return i
			}
		} else if delims > 0 {
			// Add backslashes to buffer
			slashes := backslashes(delims)
			s.B = append(s.B, slashes...)
		}

		// Write byte to buffer
		s.B = append(s.B, str[i])

		// Reset count
		delims = 0
	}

	return -1
}

// asciiSpace is a lookup table of ascii space chars (see: strings.asciiSet).
var asciiSpace = func() (as [8]uint32) {
	as['\t'/32] |= 1 << ('\t' % 32)
	as['\n'/32] |= 1 << ('\n' % 32)
	as['\v'/32] |= 1 << ('\v' % 32)
	as['\f'/32] |= 1 << ('\f' % 32)
	as['\r'/32] |= 1 << ('\r' % 32)
	as[' '/32] |= 1 << (' ' % 32)
	return
}()

// trimLeadingSpace trims the leading space from a string.
func trimLeadingSpace(str string) string {
	var start int

	for ; start < len(str); start++ {
		// If beyond ascii range, trim using slower rune check.
		if str[start] >= utf8.RuneSelf {
			return trimLeadingSpaceSlow(str[start:])
		}

		// Ascii character
		char := str[start]

		// This is first non-space ASCII, trim up to here
		if (asciiSpace[char/32] & (1 << (char % 32))) == 0 {
			break
		}
	}

	return str[start:]
}

// trimLeadingSpaceSlow trims leading space using the slower unicode.IsSpace check.
func trimLeadingSpaceSlow(str string) string {
	for i, r := range str {
		if !unicode.IsSpace(r) {
			return str[i:]
		}
	}
	return str
}

// backslashes will return a string of backslashes of given length.
func backslashes(count int) string {
	const backslashes = `\\\\\\\\\\\\\\\\\\\\`

	// Fast-path, use string const
	if count < len(backslashes) {
		return backslashes[:count]
	}

	// Slow-path, build custom string
	return backslashSlow(count)
}

// backslashSlow will build a string of backslashes of custom length.
func backslashSlow(count int) string {
	var buf strings.Builder
	for i := 0; i < count; i++ {
		buf.WriteByte('\\')
	}
	return buf.String()
}

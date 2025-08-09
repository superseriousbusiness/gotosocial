// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package format

import (
	"encoding/json"
	"time"
	"unicode/utf8"

	"code.superseriousbusiness.org/gotosocial/internal/log/level"
	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-caller"
	"codeberg.org/gruf/go-kv/v2"
)

type JSON struct{ Base }

func (fmt *JSON) Format(buf *byteutil.Buffer, stamp time.Time, pc uintptr, lvl level.LEVEL, kvs []kv.Field, msg string) {
	// Prepend opening JSON brace.
	buf.B = append(buf.B, `{`...)

	if fmt.TimeFormat != "" {
		// Append JSON formatted timestamp string.
		buf.B = append(buf.B, `"timestamp":"`...)
		fmt.AppendFormatStamp(buf, stamp)
		buf.B = append(buf.B, `", `...)
	}

	// Append JSON formatted caller func.
	buf.B = append(buf.B, `"func":"`...)
	buf.B = append(buf.B, caller.Get(pc)...)
	buf.B = append(buf.B, `", `...)

	if lvl != level.UNSET {
		// Append JSON formatted level string.
		buf.B = append(buf.B, `"level":"`...)
		buf.B = append(buf.B, lvl.String()...)
		buf.B = append(buf.B, `", `...)
	}

	// Append JSON formatted fields.
	for _, field := range kvs {
		appendStringJSON(buf, field.K)
		buf.B = append(buf.B, `:`...)
		b, _ := json.Marshal(field.V)
		buf.B = append(buf.B, b...)
		buf.B = append(buf.B, `, `...)
	}

	if msg != "" {
		// Append JSON formatted msg string.
		buf.B = append(buf.B, `"msg":`...)
		appendStringJSON(buf, msg)
	} else if string(buf.B[len(buf.B)-2:]) == ", " {
		// Drop the trailing ", ".
		buf.B = buf.B[:len(buf.B)-2]
	}

	// Append closing JSON brace.
	buf.B = append(buf.B, `}`...)
}

// appendStringJSON is modified from the encoding/json.appendString()
// function, copied in here such that we can use it for key appending.
func appendStringJSON(buf *byteutil.Buffer, src string) {
	const hex = "0123456789abcdef"
	buf.B = append(buf.B, '"')
	start := 0
	for i := 0; i < len(src); {
		if b := src[i]; b < utf8.RuneSelf {
			if jsonSafeSet[b] {
				i++
				continue
			}
			buf.B = append(buf.B, src[start:i]...)
			switch b {
			case '\\', '"':
				buf.B = append(buf.B, '\\', b)
			case '\b':
				buf.B = append(buf.B, '\\', 'b')
			case '\f':
				buf.B = append(buf.B, '\\', 'f')
			case '\n':
				buf.B = append(buf.B, '\\', 'n')
			case '\r':
				buf.B = append(buf.B, '\\', 'r')
			case '\t':
				buf.B = append(buf.B, '\\', 't')
			default:
				// This encodes bytes < 0x20 except for \b, \f, \n, \r and \t.
				buf.B = append(buf.B, '\\', 'u', '0', '0', hex[b>>4], hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		n := len(src) - i
		if n > utf8.UTFMax {
			n = utf8.UTFMax
		}
		c, size := utf8.DecodeRuneInString(src[i : i+n])
		if c == utf8.RuneError && size == 1 {
			buf.B = append(buf.B, src[start:i]...)
			buf.B = append(buf.B, `\ufffd`...)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See https://en.wikipedia.org/wiki/JSON#Safety.
		if c == '\u2028' || c == '\u2029' {
			buf.B = append(buf.B, src[start:i]...)
			buf.B = append(buf.B, '\\', 'u', '2', '0', '2', hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	buf.B = append(buf.B, src[start:]...)
	buf.B = append(buf.B, '"')
}

var jsonSafeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      true,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      true,
	'=':      true,
	'>':      true,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}

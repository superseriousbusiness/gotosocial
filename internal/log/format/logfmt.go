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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/log/level"
	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-caller"
	"codeberg.org/gruf/go-kv/v2"
	"codeberg.org/gruf/go-kv/v2/format"
)

var args = format.DefaultArgs()

type Logfmt struct{ Base }

func (fmt *Logfmt) Format(buf *byteutil.Buffer, stamp time.Time, pc uintptr, lvl level.LEVEL, kvs []kv.Field, msg string) {
	if fmt.TimeFormat != "" {
		// Append formatted timestamp string.
		buf.B = append(buf.B, `timestamp="`...)
		fmt.AppendFormatStamp(buf, stamp)
		buf.B = append(buf.B, `" `...)
	}

	// Append formatted calling func.
	buf.B = append(buf.B, `func=`...)
	buf.B = append(buf.B, caller.Get(pc)...)
	buf.B = append(buf.B, ' ')

	if lvl != level.UNSET {
		// Append formatted level string.
		buf.B = append(buf.B, `level=`...)
		buf.B = append(buf.B, lvl.String()...)
		buf.B = append(buf.B, ' ')
	}

	// Append formatted fields.
	for _, field := range kvs {
		kv.AppendQuoteString(buf, field.K)
		buf.B = append(buf.B, '=')
		buf.B = format.Global.Append(buf.B, field.V, args)
		buf.B = append(buf.B, ' ')
	}

	if msg != "" {
		// Append formatted msg string.
		buf.B = append(buf.B, `msg=`...)
		kv.AppendQuoteString(buf, msg)
	}
}

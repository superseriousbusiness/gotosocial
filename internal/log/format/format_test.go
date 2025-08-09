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

package format_test

import (
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/log/format"
	"codeberg.org/gruf/go-byteutil"
)

func BenchmarkStampCache(b *testing.B) {
	var base format.Base
	base.TimeFormat = `02/01/2006 15:04:05.000`

	b.RunParallel(func(pb *testing.PB) {
		var buf byteutil.Buffer
		buf.B = make([]byte, 0, 1024)

		for pb.Next() {
			base.AppendFormatStamp(&buf, time.Now())
			buf.B = buf.B[:0]
		}

		buf.B = buf.B[:0]
	})
}

func BenchmarkNoStampCache(b *testing.B) {
	var base format.Base
	base.TimeFormat = `02/01/2006 15:04:05.000`

	b.RunParallel(func(pb *testing.PB) {
		var buf byteutil.Buffer
		buf.B = make([]byte, 0, 1024)

		for pb.Next() {
			buf.B = time.Now().AppendFormat(buf.B, base.TimeFormat)
			buf.B = buf.B[:0]
		}

		buf.B = buf.B[:0]
	})
}

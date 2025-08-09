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
	"sync/atomic"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/log/level"
	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-kv/v2"
)

var (
	// ensure func signature conformance.
	_ FormatFunc = (*Logfmt)(nil).Format
	_ FormatFunc = (*JSON)(nil).Format
)

// FormatFunc defines a function capable of formatting a log entry (args = 1+) to a given buffer (args = 0).
type FormatFunc func(buf *byteutil.Buffer, stamp time.Time, pc uintptr, lvl level.LEVEL, kvs []kv.Field, msg string) //nolint:revive

type Base struct {
	// TimeFormat defines time.Format() layout to
	// use when appending a timestamp to log entry.
	TimeFormat string

	// stampCache caches recently formatted stamps.
	//
	// see the following benchmark:
	// goos: linux
	// goarch: amd64
	// pkg: code.superseriousbusiness.org/gotosocial/internal/log/format
	// cpu: AMD Ryzen 7 7840U w/ Radeon  780M Graphics
	// BenchmarkStampCache
	// BenchmarkStampCache-16          272199975                4.447 ns/op           0 B/op          0 allocs/op
	// BenchmarkNoStampCache
	// BenchmarkNoStampCache-16        76041058                15.94 ns/op            0 B/op          0 allocs/op
	stampCache atomic.Pointer[struct {
		stamp  time.Time
		format string
	}]
}

// AppendFormatStamp will append given timestamp according to TimeFormat,
// caching recently formatted stamp strings to reduce number of Format() calls.
func (b *Base) AppendFormatStamp(buf *byteutil.Buffer, stamp time.Time) {
	const precision = time.Millisecond

	// Load cached stamp value.
	last := b.stampCache.Load()

	// Round stamp to min precision.
	stamp = stamp.Round(precision)

	// If a cached entry exists use this string.
	if last != nil && stamp.Equal(last.stamp) {
		buf.B = append(buf.B, last.format...)
		return
	}

	// Else format new and store ASAP,
	// i.e. ignoring any CAS result.
	format := stamp.Format(b.TimeFormat)
	b.stampCache.CompareAndSwap(last, &struct {
		stamp  time.Time
		format string
	}{
		stamp:  stamp,
		format: format,
	})

	// Finally, append new timestamp.
	buf.B = append(buf.B, format...)
}

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

package gtserror

import (
	"io"

	"codeberg.org/gruf/go-byteutil"
)

// drainBody will produce a truncated output of the content
// of given io.ReadCloser body, useful for logs / errors.
func drainBody(body io.ReadCloser, trunc int) string {
	// Limit response to 'trunc' bytes.
	buf := make([]byte, trunc)

	// Read body into err buffer.
	n, _ := io.ReadFull(body, buf)

	if n == 0 {
		// No error body, return
		// reasonable error str.
		return "<empty>"
	}

	return byteutil.B2S(buf[:n])
}

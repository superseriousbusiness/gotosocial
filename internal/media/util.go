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

package media

// newHdrBuf returns a buffer of suitable size to
// read bytes from a file header or magic number.
//
// File header is *USUALLY* 261 bytes at the start
// of a file; magic number can be much less than
// that (just a few bytes).
//
// To cover both cases, this function returns a buffer
// suitable for whichever is smallest: the first 261
// bytes of the file, or the whole file.
//
// See:
//
//   - https://en.wikipedia.org/wiki/File_format#File_header
//   - https://github.com/h2non/filetype.
func newHdrBuf(fileSize int) []byte {
	bufSize := 261
	if fileSize > 0 && fileSize < bufSize {
		bufSize = fileSize
	}
	return make([]byte, bufSize)
}

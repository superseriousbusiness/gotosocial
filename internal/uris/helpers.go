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

package uris

// helper functions for string building that are designed
// to be easily inlineable, and much faster than fmt.Sprintf().
//
// you can check this by ensuring these funcs are not output in:
// go build -gcflags='-m=2' ./internal/uris/ 2>&1 | grep 'cannot inline'

func buildURL1(proto, host, path1 string) string {
	return proto + "://" + host + "/" + path1
}

func buildURL2(proto, host, path1, path2 string) string {
	return proto + "://" + host + "/" + path1 + "/" + path2
}

func buildURL4(proto, host, path1, path2, path3, path4 string) string {
	return proto + "://" + host + "/" + path1 + "/" + path2 + "/" + path3 + "/" + path4
}

func buildURL5(proto, host, path1, path2, path3, path4, path5 string) string {
	return proto + "://" + host + "/" + path1 + "/" + path2 + "/" + path3 + "/" + path4 + "/" + path5
}

func buildPath4(path1, path2, path3, path4 string) string {
	return path1 + "/" + path2 + "/" + path3 + "/" + path4
}

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

package text

import (
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

// m is the global minify instance.
var m = func() *minify.M {
	m := minify.New()
	m.Add("text/html", &html.Minifier{
		KeepEndTags: true,
		KeepQuotes:  true,
	})
	return m
}()

// MinifyHTML minifies the given string
// under the assumption that it's HTML.
//
// If input is not HTML encoded, this
// function will try to do minimization
// anyway, but this may produce unexpected
// results.
//
// If an error occurs during minimization,
// it will be logged and the original string
// returned unmodified.
func MinifyHTML(in string) string {
	out, err := m.String("text/html", in)
	if err != nil {
		log.Error(nil, err)
	}

	return out
}

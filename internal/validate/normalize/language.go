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

package normalize

import (
	"golang.org/x/text/language"
)

// Language converts a previously validated but possibly non-canonical BCP 47 tag to its canonical form.
// See: https://pkg.go.dev/golang.org/x/text/language
// See: [internal/validate.Language]
func Language(lang string) string {
	canonical, err := language.Parse(lang)
	if err != nil {
		// Should not happen: input should have been previously validated.
		return ""
	}
	return canonical.String()
}

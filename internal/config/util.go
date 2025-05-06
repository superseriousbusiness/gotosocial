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

package config

import (
	"fmt"

	"codeberg.org/gruf/go-split"
	"github.com/spf13/cast"
)

func toStringSlice(a any) ([]string, error) {
	switch a := a.(type) {
	case []string:
		return a, nil
	case string:
		return split.SplitStrings[string](a)
	case []any:
		ss := make([]string, len(a))
		for i, a := range a {
			var err error
			ss[i], err = cast.ToStringE(a)
			if err != nil {
				return nil, err
			}
		}
		return ss, nil
	default:
		return nil, fmt.Errorf("cannot cast %T to []string", a)
	}
}

func mapGet(m map[string]any, keys ...string) (any, bool) {
	for len(keys) > 0 {
		key := keys[0]
		keys = keys[1:]

		// Check for key.
		v, ok := m[key]
		if !ok {
			return nil, false
		}

		if len(keys) == 0 {
			// Has to be value.
			return v, true
		}

		// Else, it needs to have
		// nesting to keep searching.
		switch t := v.(type) {
		case map[string]any:
			m = t
		default:
			return nil, false
		}
	}
	return nil, false
}

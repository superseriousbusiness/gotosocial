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

package log

import (
	"testing"

	"codeberg.org/gruf/go-kv"
	"github.com/stretchr/testify/assert"
)

func TestToFields(t *testing.T) {
	tests := []struct {
		name           string
		keysAndValues  []any
		expectedFields []kv.Field
	}{
		{
			name:          "Even number of elements",
			keysAndValues: []any{"count", 2, "total_dropped", 0},
			expectedFields: []kv.Field{
				{K: "count", V: 2},
				{K: "total_dropped", V: 0},
			},
		},
		{
			name:          "Odd number of elements",
			keysAndValues: []any{"count", 2, "whatever"},
			expectedFields: []kv.Field{
				{K: "count", V: 2},
				{K: "whatever", V: nil},
			},
		},
		{
			name:           "Empty input",
			keysAndValues:  []any{},
			expectedFields: []kv.Field{},
		},
		{
			name:          "Single element",
			keysAndValues: []any{"single"},
			expectedFields: []kv.Field{
				{K: "single", V: nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualFields := toFields(tt.keysAndValues...)
			assert.Equal(t, tt.expectedFields, actualFields)
		})
	}
}

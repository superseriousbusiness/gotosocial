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

package util_test

import (
	"testing"

	"github.com/superseriousbusiness/gotosocial/internal/api/util"
)

func TestScopes(t *testing.T) {
	for _, test := range []struct {
		HasScope   util.Scope
		WantsScope util.Scope
		Expect     bool
	}{
		{
			HasScope:   util.ScopeRead,
			WantsScope: util.ScopeRead,
			Expect:     true,
		},
		{
			HasScope:   util.ScopeRead,
			WantsScope: util.ScopeWrite,
			Expect:     false,
		},
		{
			HasScope:   util.ScopeWrite,
			WantsScope: util.ScopeWrite,
			Expect:     true,
		},
		{
			HasScope:   util.ScopeWrite,
			WantsScope: util.ScopeRead,
			Expect:     false,
		},
		{
			HasScope:   util.ScopePush,
			WantsScope: util.ScopePush,
			Expect:     true,
		},
		{
			HasScope:   util.ScopeAdmin,
			WantsScope: util.ScopeAdmin,
			Expect:     true,
		},
		{
			HasScope:   util.ScopeProfile,
			WantsScope: util.ScopeProfile,
			Expect:     true,
		},
		{
			HasScope:   util.ScopeReadAccounts,
			WantsScope: util.ScopeWriteAccounts,
			Expect:     false,
		},
		{
			HasScope:   util.ScopeWriteAccounts,
			WantsScope: util.ScopeWriteAccounts,
			Expect:     true,
		},
		{
			HasScope:   util.ScopeWrite,
			WantsScope: util.ScopeWriteAccounts,
			Expect:     true,
		},
		{
			HasScope:   util.ScopeRead,
			WantsScope: util.ScopeWriteAccounts,
			Expect:     false,
		},
		{
			HasScope:   util.ScopeWriteAccounts,
			WantsScope: util.ScopeWrite,
			Expect:     false,
		},
		{
			HasScope:   util.ScopeProfile,
			WantsScope: util.ScopePush,
			Expect:     false,
		},
		{
			HasScope:   util.Scope("p"),
			WantsScope: util.ScopePush,
			Expect:     false,
		},
	} {
		res := test.HasScope.Permits(test.WantsScope)
		if res != test.Expect {
			t.Errorf(
				"did not get expected result %v for input: has %s, wants %s",
				test.Expect, test.HasScope, test.WantsScope,
			)
		}
	}
}

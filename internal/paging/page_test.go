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

package paging_test

import (
	"testing"

	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"golang.org/x/exp/slices"
)

type Case struct {
	// Name is the test case name.
	Name string

	// Input contains test case input ID slice.
	Input []string

	// Expect contains expected test case output.
	Expect []string

	// Page contains the paging function to use.
	Page func([]string) []string
}

var cases = []Case{
	{
		Name: "min_id and max_id set",
		Input: []string{
			"064Q5D7VG6TPPQ46T09MHJ96FW",
			"064Q5D7VGPTC4NK5T070VYSSF8",
			"064Q5D7VH5F0JXG6W5NCQ3JCWW",
			"064Q5D7VHMSW9DF3GCS088VAZC",
			"064Q5D7VJ073XG9ZTWHA2KHN10",
			"064Q5D7VJADJTPA3GW8WAX10TW",
			"064Q5D7VJMWXZD3S1KT7RD51N8",
			"064Q5D7VJYFBYSAH86KDBKZ6AC",
			"064Q5D7VK8H7WMJS399SHEPCB0",
			"064Q5D7VKG5EQ43TYP71B4K6K0",
		},
		Expect: []string{
			"064Q5D7VGPTC4NK5T070VYSSF8",
			"064Q5D7VH5F0JXG6W5NCQ3JCWW",
			"064Q5D7VHMSW9DF3GCS088VAZC",
			"064Q5D7VJ073XG9ZTWHA2KHN10",
			"064Q5D7VJADJTPA3GW8WAX10TW",
			"064Q5D7VJMWXZD3S1KT7RD51N8",
			"064Q5D7VJYFBYSAH86KDBKZ6AC",
			"064Q5D7VK8H7WMJS399SHEPCB0",
		},
		Page: (&paging.Page[string]{
			Min: paging.MinID("064Q5D7VG6TPPQ46T09MHJ96FW", ""),
			Max: paging.MaxID("064Q5D7VKG5EQ43TYP71B4K6K0"),
		}).PageAsc,
	},
	{
		Name: "min_id, max_id and limit set",
		Input: []string{
			"064Q5D7VG6TPPQ46T09MHJ96FW",
			"064Q5D7VGPTC4NK5T070VYSSF8",
			"064Q5D7VH5F0JXG6W5NCQ3JCWW",
			"064Q5D7VHMSW9DF3GCS088VAZC",
			"064Q5D7VJ073XG9ZTWHA2KHN10",
			"064Q5D7VJADJTPA3GW8WAX10TW",
			"064Q5D7VJMWXZD3S1KT7RD51N8",
			"064Q5D7VJYFBYSAH86KDBKZ6AC",
			"064Q5D7VK8H7WMJS399SHEPCB0",
			"064Q5D7VKG5EQ43TYP71B4K6K0",
		},
		Expect: []string{
			"064Q5D7VGPTC4NK5T070VYSSF8",
			"064Q5D7VH5F0JXG6W5NCQ3JCWW",
			"064Q5D7VHMSW9DF3GCS088VAZC",
			"064Q5D7VJ073XG9ZTWHA2KHN10",
			"064Q5D7VJADJTPA3GW8WAX10TW",
		},
		Page: (&paging.Page[string]{
			Min:   paging.MinID("064Q5D7VG6TPPQ46T09MHJ96FW", ""),
			Max:   paging.MaxID("064Q5D7VKG5EQ43TYP71B4K6K0"),
			Limit: 5,
		}).PageAsc,
	},
	{
		Name: "min_id, max_id and too-large limit set",
		Input: []string{
			"064Q5D7VG6TPPQ46T09MHJ96FW",
			"064Q5D7VGPTC4NK5T070VYSSF8",
			"064Q5D7VH5F0JXG6W5NCQ3JCWW",
			"064Q5D7VHMSW9DF3GCS088VAZC",
			"064Q5D7VJ073XG9ZTWHA2KHN10",
			"064Q5D7VJADJTPA3GW8WAX10TW",
			"064Q5D7VJMWXZD3S1KT7RD51N8",
			"064Q5D7VJYFBYSAH86KDBKZ6AC",
			"064Q5D7VK8H7WMJS399SHEPCB0",
			"064Q5D7VKG5EQ43TYP71B4K6K0",
		},
		Expect: []string{
			"064Q5D7VGPTC4NK5T070VYSSF8",
			"064Q5D7VH5F0JXG6W5NCQ3JCWW",
			"064Q5D7VHMSW9DF3GCS088VAZC",
			"064Q5D7VJ073XG9ZTWHA2KHN10",
			"064Q5D7VJADJTPA3GW8WAX10TW",
			"064Q5D7VJMWXZD3S1KT7RD51N8",
			"064Q5D7VJYFBYSAH86KDBKZ6AC",
			"064Q5D7VK8H7WMJS399SHEPCB0",
		},
		Page: (&paging.Page[string]{
			Min:   paging.MinID("064Q5D7VG6TPPQ46T09MHJ96FW", ""),
			Max:   paging.MaxID("064Q5D7VKG5EQ43TYP71B4K6K0"),
			Limit: 100,
		}).PageAsc,
	},
	{
		Name: "since_id and max_id set",
		Input: []string{
			"064Q5D7VG6TPPQ46T09MHJ96FW",
			"064Q5D7VGPTC4NK5T070VYSSF8",
			"064Q5D7VH5F0JXG6W5NCQ3JCWW",
			"064Q5D7VHMSW9DF3GCS088VAZC",
			"064Q5D7VJ073XG9ZTWHA2KHN10",
			"064Q5D7VJADJTPA3GW8WAX10TW",
			"064Q5D7VJMWXZD3S1KT7RD51N8",
			"064Q5D7VJYFBYSAH86KDBKZ6AC",
			"064Q5D7VK8H7WMJS399SHEPCB0",
			"064Q5D7VKG5EQ43TYP71B4K6K0",
		},
		Expect: []string{
			"064Q5D7VK8H7WMJS399SHEPCB0",
			"064Q5D7VJYFBYSAH86KDBKZ6AC",
			"064Q5D7VJMWXZD3S1KT7RD51N8",
			"064Q5D7VJADJTPA3GW8WAX10TW",
			"064Q5D7VJ073XG9ZTWHA2KHN10",
			"064Q5D7VHMSW9DF3GCS088VAZC",
			"064Q5D7VH5F0JXG6W5NCQ3JCWW",
			"064Q5D7VGPTC4NK5T070VYSSF8",
		},
		Page: (&paging.Page[string]{
			Min: paging.MinID("", "064Q5D7VG6TPPQ46T09MHJ96FW"),
			Max: paging.MaxID("064Q5D7VKG5EQ43TYP71B4K6K0"),
		}).PageAsc,
	},
}

func TestPage(t *testing.T) {
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// Page the input slice.
			out := c.Page(c.Input)

			// Check paged output is as expected.
			if !slices.Equal(out, c.Expect) {
				t.Errorf("\nreceived=%v\nexpect%v\n", out, c.Expect)
			}
		})
	}
}

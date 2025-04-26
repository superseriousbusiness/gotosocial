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

package xslices_test

import (
	"math/rand"
	"net/url"
	"slices"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
	"github.com/stretchr/testify/assert"
)

func TestGrowJust(t *testing.T) {
	for _, l := range []int{0, 2, 4, 8, 16, 32, 64} {
		for _, x := range []int{0, 2, 4, 8, 16, 32, 64} {
			s := make([]int, l, l+x)
			for _, g := range []int{0, 2, 4, 8, 16, 32, 64} {
				s2 := xslices.GrowJust(s, g)

				// Slice length should not be different.
				assert.Equal(t, len(s), len(s2))

				switch {
				// If slice already has capacity for
				// 'g' then it should not be changed.
				case cap(s) >= len(s)+g:
					assert.Equal(t, cap(s), cap(s2))

				// Else, returned slice should only
				// have capacity for original length
				// plus extra elements, NOTHING MORE.
				default:
					assert.Equal(t, cap(s2), len(s)+g)
				}
			}
		}
	}
}

func TestAppendJust(t *testing.T) {
	for _, l := range []int{0, 2, 4, 8, 16, 32, 64} {
		for _, x := range []int{0, 2, 4, 8, 16, 32, 64} {
			s := make([]int, l, l+x)

			// Randomize slice.
			for i := range s {
				s[i] = rand.Int()
			}

			for _, a := range []int{0, 2, 4, 8, 16, 32, 64} {
				toAppend := make([]int, a)

				// Randomize appended vals.
				for i := range toAppend {
					toAppend[i] = rand.Int()
				}

				s2 := xslices.AppendJust(s, toAppend...)

				// Slice length should be as expected.
				assert.Equal(t, len(s)+a, len(s2))

				// Slice contents should be as expected.
				assert.Equal(t, append(s, toAppend...), s2)

				switch {
				// If slice already has capacity for
				// 'toAppend' then it should not change.
				case cap(s) >= len(s)+a:
					assert.Equal(t, cap(s), cap(s2))

				// Else, returned slice should only
				// have capacity for original length
				// plus extra elements, NOTHING MORE.
				default:
					assert.Equal(t, len(s)+a, cap(s2))
				}
			}
		}
	}
}

func TestGather(t *testing.T) {
	out := xslices.Gather(nil, []*url.URL{
		{Scheme: "https", Host: "google.com", Path: "/some-search"},
		{Scheme: "http", Host: "example.com", Path: "/robots.txt"},
	}, (*url.URL).String)
	if !slices.Equal(out, []string{
		"https://google.com/some-search",
		"http://example.com/robots.txt",
	}) {
		t.Fatal("unexpected gather output")
	}

	out = xslices.Gather([]string{
		"starting input string",
		"another starting input",
	}, []*url.URL{
		{Scheme: "https", Host: "google.com", Path: "/some-search"},
		{Scheme: "http", Host: "example.com", Path: "/robots.txt"},
	}, (*url.URL).String)
	if !slices.Equal(out, []string{
		"starting input string",
		"another starting input",
		"https://google.com/some-search",
		"http://example.com/robots.txt",
	}) {
		t.Fatal("unexpected gather output")
	}
}

func TestGatherIf(t *testing.T) {
	out := xslices.GatherIf(nil, []string{
		"hello world",
		"not hello world",
		"hello world",
	}, func(s string) (string, bool) {
		return s, s == "hello world"
	})
	if !slices.Equal(out, []string{
		"hello world",
		"hello world",
	}) {
		t.Fatal("unexpected gatherif output")
	}

	out = xslices.GatherIf([]string{
		"starting input string",
		"another starting input",
	}, []string{
		"hello world",
		"not hello world",
		"hello world",
	}, func(s string) (string, bool) {
		return s, s == "hello world"
	})
	if !slices.Equal(out, []string{
		"starting input string",
		"another starting input",
		"hello world",
		"hello world",
	}) {
		t.Fatal("unexpected gatherif output")
	}
}

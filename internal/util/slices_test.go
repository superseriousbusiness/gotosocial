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
	"net/url"
	"slices"
	"testing"

	"github.com/superseriousbusiness/gotosocial/internal/util"
)

var (
	testURLSlice = []*url.URL{}
)

func TestGather(t *testing.T) {
	out := util.Gather(nil, []*url.URL{
		{Scheme: "https", Host: "google.com", Path: "/some-search"},
		{Scheme: "http", Host: "example.com", Path: "/robots.txt"},
	}, (*url.URL).String)
	if !slices.Equal(out, []string{
		"https://google.com/some-search",
		"http://example.com/robots.txt",
	}) {
		t.Fatal("unexpected gather output")
	}

	out = util.Gather([]string{
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
	out := util.GatherIf(nil, []string{
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

	out = util.GatherIf([]string{
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

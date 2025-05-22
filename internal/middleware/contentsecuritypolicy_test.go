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

package middleware_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/middleware"
)

func TestBuildContentSecurityPolicy(t *testing.T) {
	type cspTest struct {
		extraURLs []string
		expected  string
	}

	for _, test := range []cspTest{
		{
			extraURLs: nil,
			expected:  "default-src 'self'; connect-src 'self' https://api.listenbrainz.org/1/user/; object-src 'none'; img-src 'self' blob:; media-src 'self'",
		},
		{
			extraURLs: []string{
				"https://some-bucket-provider.com",
			},
			expected: "default-src 'self'; connect-src 'self' https://api.listenbrainz.org/1/user/; object-src 'none'; img-src 'self' blob: https://some-bucket-provider.com; media-src 'self' https://some-bucket-provider.com",
		},
		{
			extraURLs: []string{
				"https://some-bucket-provider.com:6969",
			},
			expected: "default-src 'self'; connect-src 'self' https://api.listenbrainz.org/1/user/; object-src 'none'; img-src 'self' blob: https://some-bucket-provider.com:6969; media-src 'self' https://some-bucket-provider.com:6969",
		},
		{
			extraURLs: []string{
				"http://some-bucket-provider.com:6969",
			},
			expected: "default-src 'self'; connect-src 'self' https://api.listenbrainz.org/1/user/; object-src 'none'; img-src 'self' blob: http://some-bucket-provider.com:6969; media-src 'self' http://some-bucket-provider.com:6969",
		},
		{
			extraURLs: []string{
				"https://s3.nl-ams.scw.cloud",
			},
			expected: "default-src 'self'; connect-src 'self' https://api.listenbrainz.org/1/user/; object-src 'none'; img-src 'self' blob: https://s3.nl-ams.scw.cloud; media-src 'self' https://s3.nl-ams.scw.cloud",
		},
		{
			extraURLs: []string{
				"https://s3.nl-ams.scw.cloud",
				"https://s3.somewhere.else.example.org",
			},
			expected: "default-src 'self'; connect-src 'self' https://api.listenbrainz.org/1/user/; object-src 'none'; img-src 'self' blob: https://s3.nl-ams.scw.cloud https://s3.somewhere.else.example.org; media-src 'self' https://s3.nl-ams.scw.cloud https://s3.somewhere.else.example.org",
		},
	} {
		csp := middleware.BuildContentSecurityPolicy(test.extraURLs...)
		if csp != test.expected {
			t.Logf("expected '%s', got '%s'", test.expected, csp)
			t.Fail()
		}
	}
}

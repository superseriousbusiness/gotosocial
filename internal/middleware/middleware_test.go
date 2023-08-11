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

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
)

func TestBuildContentSecurityPolicy(t *testing.T) {
	type cspTest struct {
		s3Endpoint string
		s3Proxy    bool
		expected   string
		actual     string
	}

	for _, test := range []cspTest{
		{
			s3Endpoint: "",
			s3Proxy:    false,
			expected:   "default-src 'self'",
		},
		{
			s3Endpoint: "https://some-bucket-provider.com",
			s3Proxy:    false,
			expected:   "default-src 'self'; image-src https://some-bucket-provider.com; media-src https://some-bucket-provider.com",
		},
		{
			s3Endpoint: "https://some-bucket-provider.com:6969",
			s3Proxy:    false,
			expected:   "default-src 'self'; image-src https://some-bucket-provider.com:6969; media-src https://some-bucket-provider.com:6969",
		},
		{
			s3Endpoint: "some-bucket-provider.com:6969",
			s3Proxy:    false,
			expected:   "default-src 'self'; image-src some-bucket-provider.com:6969; media-src some-bucket-provider.com:6969",
		},
		{
			s3Endpoint: "s3.nl-ams.scw.cloud",
			s3Proxy:    false,
			expected:   "default-src 'self'; image-src s3.nl-ams.scw.cloud; media-src s3.nl-ams.scw.cloud",
		},
		{
			s3Endpoint: "https://some-bucket-provider.com",
			s3Proxy:    true,
			expected:   "default-src 'self'",
		},
		{
			s3Endpoint: "https://some-bucket-provider.com:6969",
			s3Proxy:    true,
			expected:   "default-src 'self'",
		},
		{
			s3Endpoint: "some-bucket-provider.com:6969",
			s3Proxy:    true,
			expected:   "default-src 'self'",
		},
		{
			s3Endpoint: "s3.nl-ams.scw.cloud",
			s3Proxy:    true,
			expected:   "default-src 'self'",
		},
	} {
		config.SetStorageS3Endpoint(test.s3Endpoint)
		config.SetStorageS3Proxy(test.s3Proxy)

		csp := middleware.BuildContentSecurityPolicy()
		if csp != test.expected {
			t.Logf("expected '%s', got '%s'", test.expected, csp)
			t.Fail()
		}
	}
}

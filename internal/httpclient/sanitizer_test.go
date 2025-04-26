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

package httpclient_test

import (
	"errors"
	"net/netip"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/httpclient"
)

func TestSafeIP(t *testing.T) {
	tests := []struct {
		name string
		ip   netip.Addr
	}{
		// IPv4 tests
		{
			name: "IPv4 this host on this network",
			ip:   netip.MustParseAddr("0.0.0.0"),
		},
		{
			name: "IPv4 dummy address",
			ip:   netip.MustParseAddr("192.0.0.8"),
		},
		{
			name: "IPv4 Port Control Protocol Anycast",
			ip:   netip.MustParseAddr("192.0.0.9"),
		},
		{
			name: "IPv4 Traversal Using Relays around NAT Anycast",
			ip:   netip.MustParseAddr("192.0.0.10"),
		},
		{
			name: "IPv4 NAT64/DNS64 Discovery 1",
			ip:   netip.MustParseAddr("192.0.0.17"),
		},
		{
			name: "IPv4 NAT64/DNS64 Discovery 2",
			ip:   netip.MustParseAddr("192.0.0.171"),
		},
		// IPv6 tests
		{
			name: "IPv4-mapped address",
			ip:   netip.MustParseAddr("::ffff:169.254.169.254"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if safe := httpclient.SafeIP(tc.ip); safe {
				t.Fatalf("Expected IP %s to not safe (%t), got: %t", tc.ip, false, safe)
			}
		})
	}
}

func TestSanitizer(t *testing.T) {
	s := httpclient.Sanitizer{
		Allow: []netip.Prefix{
			netip.MustParsePrefix("192.0.0.8/32"),
			netip.MustParsePrefix("::ffff:169.254.169.254/128"),
		},
		Block: []netip.Prefix{
			netip.MustParsePrefix("93.184.216.34/32"), // example.org
		},
	}

	tests := []struct {
		name     string
		ntwrk    string
		addr     string
		expected error
	}{
		// IPv4 tests
		{
			name:     "IPv4 this host on this network",
			ntwrk:    "tcp4",
			addr:     "0.0.0.0:80",
			expected: httpclient.ErrReservedAddr,
		},
		{
			name:     "IPv4 dummy address",
			ntwrk:    "tcp4",
			addr:     "192.0.0.8:80",
			expected: nil, // We allowed this explicitly.
		},
		{
			name:     "IPv4 Port Control Protocol Anycast",
			ntwrk:    "tcp4",
			addr:     "192.0.0.9:80",
			expected: httpclient.ErrReservedAddr,
		},
		{
			name:     "IPv4 Traversal Using Relays around NAT Anycast",
			ntwrk:    "tcp4",
			addr:     "192.0.0.10:80",
			expected: httpclient.ErrReservedAddr,
		},
		{
			name:     "IPv4 NAT64/DNS64 Discovery 1",
			ntwrk:    "tcp4",
			addr:     "192.0.0.17:80",
			expected: httpclient.ErrReservedAddr,
		},
		{
			name:     "IPv4 NAT64/DNS64 Discovery 2",
			ntwrk:    "tcp4",
			addr:     "192.0.0.171:80",
			expected: httpclient.ErrReservedAddr,
		},
		{
			name:     "example.org",
			ntwrk:    "tcp4",
			addr:     "93.184.216.34:80",
			expected: httpclient.ErrReservedAddr, // We blocked this explicitly.
		},
		// IPv6 tests
		{
			name:     "IPv4-mapped address",
			ntwrk:    "tcp6",
			addr:     "[::ffff:169.254.169.254]:80",
			expected: nil, // We allowed this explicitly.
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := s.Sanitize(tc.ntwrk, tc.addr, nil); !errors.Is(err, tc.expected) {
				t.Fatalf("Expected error %q for addr %s, got: %q", tc.expected, tc.addr, err)
			}
		})
	}
}

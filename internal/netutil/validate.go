/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package netutil

import (
	"net/netip"
)

var (
	// IPv6GlobalUnicast is the global IPv6 unicast IP prefix.
	IPv6GlobalUnicast = netip.MustParsePrefix("ff00::/8")

	// IPvReserved contains IPv4 reserved IP prefixes.
	IPv4Reserved = [...]netip.Prefix{
		netip.MustParsePrefix("0.0.0.0/8"),       // Current network
		netip.MustParsePrefix("10.0.0.0/8"),      // Private
		netip.MustParsePrefix("100.64.0.0/10"),   // RFC6598
		netip.MustParsePrefix("127.0.0.0/8"),     // Loopback
		netip.MustParsePrefix("169.254.0.0/16"),  // Link-local
		netip.MustParsePrefix("172.16.0.0/12"),   // Private
		netip.MustParsePrefix("192.0.0.0/24"),    // RFC6890
		netip.MustParsePrefix("192.0.2.0/24"),    // Test, doc, examples
		netip.MustParsePrefix("192.88.99.0/24"),  // IPv6 to IPv4 relay
		netip.MustParsePrefix("192.168.0.0/16"),  // Private
		netip.MustParsePrefix("198.18.0.0/15"),   // Benchmarking tests
		netip.MustParsePrefix("198.51.100.0/24"), // Test, doc, examples
		netip.MustParsePrefix("203.0.113.0/24"),  // Test, doc, examples
		netip.MustParsePrefix("224.0.0.0/4"),     // Multicast
		netip.MustParsePrefix("240.0.0.0/4"),     // Reserved (includes broadcast / 255.255.255.255)
	}
)

// ValidateAddr will parse a netip.AddrPort from string, and return the result of ValidateIP() on addr.
func ValidateAddr(s string) bool {
	ipport, err := netip.ParseAddrPort(s)
	if err != nil {
		return false
	}
	return ValidateIP(ipport.Addr())
}

// ValidateIP returns whether IP is an IPv4/6 address in non-reserved, public ranges.
func ValidateIP(ip netip.Addr) bool {
	switch {
	// IPv4: check if IPv4 in reserved nets
	case ip.Is4():
		for _, reserved := range IPv4Reserved {
			if reserved.Contains(ip) {
				return false
			}
		}
		return true

	// IPv6: check if in global unicast (public internet)
	case ip.Is6():
		return IPv6GlobalUnicast.Contains(ip)

	// Assume malicious by default
	default:
		return false
	}
}

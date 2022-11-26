package netutil

import (
	"net/netip"
	"testing"
)

func TestValidateIP(t *testing.T) {
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
			if valid := ValidateIP(tc.ip); valid != false {
				t.Fatalf("Expected IP %s to be: %t, got: %t", tc.ip, false, valid)
			}
		})
	}
}

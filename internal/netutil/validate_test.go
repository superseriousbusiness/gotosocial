package netutil

import (
	"net/netip"
	"testing"
)

func TestValidateIP(t *testing.T) {
	tests := []struct {
		name  string
		ip    netip.Addr
		valid bool
	}{
		{
			name:  "IPv4-mapped address",
			ip:    netip.MustParseAddr("::ffff:169.254.169.254"),
			valid: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if valid := ValidateIP(tc.ip); valid != tc.valid {
				t.Fatalf("Expected IP %s to be: %t, got: %t", tc.ip, tc.valid, valid)
			}
		})
	}
}

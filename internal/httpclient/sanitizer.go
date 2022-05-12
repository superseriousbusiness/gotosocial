package httpclient

import (
	"net/netip"
	"syscall"

	"github.com/superseriousbusiness/gotosocial/internal/netutil"
)

type sanitizer struct {
	allow []netip.Prefix
	block []netip.Prefix
}

// Sanitize implements the required net.Dialer.Control function signature.
func (s *sanitizer) Sanitize(ntwrk, addr string, _ syscall.RawConn) error {
	// Parse IP+port from addr
	ipport, err := netip.ParseAddrPort(addr)
	if err != nil {
		return err
	}

	// Seperate the IP
	ip := ipport.Addr()

	// Check if this is explicitly allowed
	for i := 0; i < len(s.allow); i++ {
		if s.allow[i].Contains(ip) {
			return nil
		}
	}

	// Now check if explicity blocked
	for i := 0; i < len(s.block); i++ {
		if s.block[i].Contains(ip) {
			return ErrReservedAddr
		}
	}

	// Validate this is a safe IP
	if !netutil.ValidateIP(ip) {
		return ErrReservedAddr
	}

	return nil
}

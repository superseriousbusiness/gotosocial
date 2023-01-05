/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

	if !(ntwrk == "tcp4" || ntwrk == "tcp6") {
		return ErrInvalidNetwork
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

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

package util

import (
	"net/url"
	"strings"

	"golang.org/x/net/idna"
)

var (
	// IDNA (Internationalized Domain Names for Applications)
	// profiles for fast punycode conv and full verification.
	punifyProfile = *idna.Punycode
	verifyProfile = *idna.Lookup
)

// PunifySafely validates the provided domain name,
// and converts unicode chars to ASCII, i.e. punified form.
func PunifySafely(domain string) (string, error) {
	if i := strings.LastIndexByte(domain, ':'); i >= 0 {

		// If there is a port included in domain, we
		// strip it as colon is invalid in a hostname.
		domain, port := domain[:i], domain[i:]
		domain, err := verifyProfile.ToASCII(domain)
		if err != nil {
			return "", err
		}

		// Then rebuild with port after.
		domain = strings.ToLower(domain)
		return domain + port, nil
	} else { //nolint:revive

		// Otherwise we just punify domain as-is.
		domain, err := verifyProfile.ToASCII(domain)
		return strings.ToLower(domain), err
	}
}

// Punify is a faster form of PunifySafely() without validation.
func Punify(domain string) (string, error) {
	domain, err := punifyProfile.ToASCII(domain)
	return strings.ToLower(domain), err
}

// DePunify converts any punycode-encoded unicode characters
// in domain name back to their origin unicode. Please note
// that this performs minimal validation of domain name.
func DePunify(domain string) (string, error) {
	domain = strings.ToLower(domain)
	return punifyProfile.ToUnicode(domain)
}

// URIMatches returns true if the expected URI matches
// any of the given URIs, taking account of punycode.
func URIMatches(expect *url.URL, uris ...*url.URL) (ok bool, err error) {

	// Create new URL to hold
	// punified URI information.
	punyURI := new(url.URL)
	*punyURI = *expect

	// Set punified expected URL host.
	punyURI.Host, err = Punify(expect.Host)
	if err != nil {
		return false, err
	}

	// Calculate expected URI string.
	expectStr := punyURI.String()

	// Use punyURI to iteratively
	// store each punified URI info
	// and generate punified URI
	// strings to check against.
	for _, uri := range uris {
		*punyURI = *uri
		punyURI.Host, err = Punify(uri.Host)
		if err != nil {
			return false, err
		}

		// Check for a match against expect.
		if expectStr == punyURI.String() {
			return true, nil
		}
	}

	// Didn't match.
	return false, nil
}

// PunifyURI returns a new copy of URI with the 'host'
// part converted to punycode with PunifySafely().
// For simple comparisons prefer the faster URIMatches().
func PunifyURI(in *url.URL) (*url.URL, error) {
	punyHost, err := PunifySafely(in.Host)
	if err != nil {
		return nil, err
	}
	out := new(url.URL)
	*out = *in
	out.Host = punyHost
	return out, nil
}

// PunifyURIToStr returns given URI serialized with the
// 'host' part converted to punycode with PunifySafely().
// For simple comparisons prefer the faster URIMatches().
func PunifyURIToStr(in *url.URL) (string, error) {
	punyHost, err := PunifySafely(in.Host)
	if err != nil {
		return "", err
	}
	oldHost := in.Host
	in.Host = punyHost
	str := in.String()
	in.Host = oldHost
	return str, nil
}

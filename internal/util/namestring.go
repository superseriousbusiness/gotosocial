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
	"fmt"
	"net/url"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/regexes"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
)

// ExtractNamestringParts extracts the username test_user and
// the domain example.org from a string like @test_user@example.org.
//
// If nothing is matched, it will return an error.
func ExtractNamestringParts(namestring string) (username, host string, err error) {
	matches := regexes.MentionName.FindStringSubmatch(namestring)
	switch len(matches) {
	case 2:
		return matches[1], "", nil
	case 3:
		return matches[1], matches[2], nil
	default:
		return "", "", fmt.Errorf("couldn't match namestring %s", namestring)
	}
}

// ExtractWebfingerParts returns the username and domain from the "subject"
// part of a webfinger response: either an account namestring or an actor URI.
//
// All AP implementations in the wild perform webfinger account resource
// queries with the "acct" scheme and without a leading "@"" on the username.
// This is also the format the "subject" in a webfinger response adheres to.
//
// Despite this fact, we're permissive about a single leading @. This makes
// a query for "acct:user@domain.tld" and "acct:@user@domain.tld" equivalent.
//
// We also permit a resource of "user@domain.tld" or "@user@domain.tld", without
// a scheme. In that case it gets interpreted as if it was using "acct:".
//
// Will error if parsing fails, or if the extracted username or domain are empty.
func ExtractWebfingerParts(subject string) (
	string, // username
	string, // domain
	error,
) {
	u, err := url.ParseRequestURI(subject)
	if err != nil {
		// Most likely reason for failing to parse is if
		// the "acct" scheme was missing but a :port was
		// included. So try an extra time with the scheme.
		u, err = url.ParseRequestURI("acct:" + subject)
	}
	if err != nil {
		return "", "", fmt.Errorf("failed to parse %s: %w", subject, err)
	}

	switch u.Scheme {

	// Subject looks like
	// "https://example.org/users/whatever"
	// or "https://example.org/@whatever".
	case "http", "https":
		return partsFromURI(u)

	// Subject looks like
	// "acct:whatever@example.org"
	// or "acct:@whatever@example.org".
	case "acct":
		// Pass string without "acct:" prefix.
		return partsFromNamestring(u.Opaque)

	// Subject was probably a relative URL.
	// Fail since we need the domain.
	case "":
		return "", "", fmt.Errorf("no scheme for resource %s", subject)
	}

	return "", "", fmt.Errorf("unsupported scheme %s for resource %s", u.Scheme, subject)
}

// partsFromNamestring returns the username
// and host parts extracted from a passed-in actor
// namestring of the format "whatever@example.org".
//
// The function returns an error if username or
// host cannot be extracted.
func partsFromNamestring(namestring string) (
	string, // username
	string, // host
	error,
) {
	// Trim all leading "@" symbols,
	// and then inject just one "@".
	namestring = strings.TrimLeft(namestring, "@")
	namestring = "@" + namestring

	username, host, err := ExtractNamestringParts(namestring)
	if err != nil {
		return "", "", err
	}

	if username == "" {
		err := fmt.Errorf("failed to extract username from: %s", namestring)
		return "", "", err
	}

	if host == "" {
		err := fmt.Errorf("failed to extract domain from: %s", namestring)
		return "", "", err
	}

	return username, host, nil
}

// partsFromURI returns the username and host
// extracted from the passed in actor URI.
//
// The username will be extracted from one of
// the patterns "/@whatever" or "/users/whatever".
// These paths match the "aliases" and "links"
// we include in our own webfinger responses.
//
// This function tries to be permissive with
// regard to leading "@" symbols. Nevertheless,
// an error will be returned if username or host
// cannot be extracted.
func partsFromURI(uri *url.URL) (
	string, // username
	string, // host
	error,
) {
	host := uri.Host
	if host == "" {
		err := fmt.Errorf("failed to extract domain from: %s", uri)
		return "", "", err
	}

	// Copy the URL, taking
	// only the parts we need.
	short := &url.URL{
		Path: uri.Path,
	}

	// Try "/users/whatever".
	username, err := uris.ParseUserPath(short)
	if err == nil && username != "" {
		return username, host, nil
	}

	// Try "/@whatever"
	username, err = uris.ParseUserWebPath(short)
	if err == nil && username != "" {
		return username, host, nil
	}

	// Try some exotic fallbacks like
	// "/users/@whatever", "/@@whatever", etc.
	short.Path = strings.TrimLeft(short.Path, "/")
	segs := strings.Split(short.Path, "/")
	if segs[0] == "users" {
		username = segs[1]
	} else {
		username = segs[0]
	}

	username = strings.TrimLeft(username, "@")
	if username != "" {
		return username, host, nil
	}

	return "", "", fmt.Errorf("failed to extract username from: %s", uri)
}

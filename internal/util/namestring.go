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

	"github.com/superseriousbusiness/gotosocial/internal/regexes"
)

// ExtractNamestringParts extracts the username test_user and
// the domain example.org from a string like @test_user@example.org.
//
// If nothing is matched, it will return an error.
func ExtractNamestringParts(mention string) (username, host string, err error) {
	matches := regexes.MentionName.FindStringSubmatch(mention)
	switch len(matches) {
	case 2:
		return matches[1], "", nil
	case 3:
		return matches[1], matches[2], nil
	default:
		return "", "", fmt.Errorf("couldn't match mention %s", mention)
	}
}

// ExtractWebfingerParts returns the username and domain from either an
// account query or an actor URI.
//
// All implementations in the wild generate webfinger account resource
// queries with the "acct" scheme and without a leading "@"" on the username.
// This is also the format the "subject" in a webfinger response adheres to.
//
// Despite this fact, we're being permissive about a single leading @. This
// makes a query for acct:user@domain.tld and acct:@user@domain.tld
// equivalent. But a query for acct:@@user@domain.tld will have its username
// returned with the @ prefix.
//
// We also permit a resource of user@domain.tld or @user@domain.tld, without
// a scheme. In that case it gets interpreted as if it was using the "acct"
// scheme.
//
// When parsing fails, an error is returned.
func ExtractWebfingerParts(webfinger string) (username, host string, err error) {
	orig := webfinger

	u, oerr := url.ParseRequestURI(webfinger)
	if oerr != nil {
		// Most likely reason for failing to parse is if the "acct" scheme was
		// missing but a :port was included. So try an extra time with the scheme.
		u, err = url.ParseRequestURI("acct:" + webfinger)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse %s with acct sheme: %w", orig, oerr)
		}
	}

	if u.Scheme == "http" || u.Scheme == "https" {
		return ExtractWebfingerPartsFromURI(u)
	}

	if u.Scheme != "acct" {
		return "", "", fmt.Errorf("unsupported scheme: %s for resource: %s", u.Scheme, orig)
	}

	userDomain := strings.Split(
		strings.TrimPrefix(
			u.Opaque,
			"@",
		),
		"@")
	if len(userDomain) != 2 {
		return "", "", fmt.Errorf("failed to extract user and domain from: %s", orig)
	}
	return userDomain[0], userDomain[1], nil
}

// ExtractWebfingerPartsFromURI returns the user and domain extracted from
// the passed in URI. The URI should be an actor URI.
//
// The domain returned is the hostname, and the user will be extracted
// from either /@test_user or /users/test_user. These two paths match the
// "aliasses" we include in our webfinger response and are also present in
// our "links".
//
// Like with ExtractWebfingerParts, we're being permissive about a single
// leading @.
//
// Errors are returned in case we end up with an empty domain or username.
func ExtractWebfingerPartsFromURI(uri *url.URL) (username, host string, err error) {
	host = uri.Host
	if host == "" {
		return "", "", fmt.Errorf("failed to extract domain from: %s", uri)
	}

	// strip any leading slashes
	path := strings.TrimLeft(uri.Path, "/")
	segs := strings.Split(path, "/")
	if segs[0] == "users" {
		username = segs[1]
	} else {
		username = segs[0]
	}

	username = strings.TrimPrefix(username, "@")
	if username == "" {
		return "", "", fmt.Errorf("failed to extract username from: %s", uri)
	}

	return
}

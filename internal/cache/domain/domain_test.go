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

package domain_test

import (
	"errors"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/cache/domain"
)

func TestCache(t *testing.T) {
	c := new(domain.Cache)

	cachedDomains := []string{
		"google.com",               //
		"mail.google.com",          // should be ignored since covered above
		"dev.mail.google.com",      // same again
		"google.co.uk",             //
		"mail.google.co.uk",        //
		"pleroma.bad.host",         //
		"pleroma.still.a.bad.host", //
	}

	loader := func() ([]string, error) {
		t.Log("load: returning cached domains")
		return cachedDomains, nil
	}

	// Check a list of known matching domains.
	for _, domain := range []string{
		"google.com",
		"mail.google.com",
		"dev.mail.google.com",
		"google.co.uk",
		"mail.google.co.uk",
		"pleroma.bad.host",
		"dev.pleroma.bad.host",
		"pleroma.still.a.bad.host",
		"dev.pleroma.still.a.bad.host",
	} {
		t.Logf("checking domain matches: %s", domain)
		if b, _ := c.Matches(domain, loader); !b {
			t.Fatalf("domain should be matched: %s", domain)
		}
	}

	// Check a list of known unmatched domains.
	for _, domain := range []string{
		"askjeeves.com",
		"ask-kim.co.uk",
		"google.ie",
		"mail.google.ie",
		"gts.bad.host",
		"mastodon.bad.host",
		"akkoma.still.a.bad.host",
	} {
		t.Logf("checking domain isn't matched: %s", domain)
		if b, _ := c.Matches(domain, loader); b {
			t.Fatalf("domain should not be matched: %s", domain)
		}
	}

	// Clear the cache
	t.Logf("%+v\n", c)
	c.Clear()
	t.Logf("%+v\n", c)

	knownErr := errors.New("known error")

	// Check that reload is actually performed and returns our error
	if _, err := c.Matches("", func() ([]string, error) {
		t.Log("load: returning known error")
		return nil, knownErr
	}); !errors.Is(err, knownErr) {
		t.Fatalf("matches did not return expected error: %v", err)
	}
}

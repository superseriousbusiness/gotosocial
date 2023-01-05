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

package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/cache/domain"
)

func TestBlockCache(t *testing.T) {
	c := domain.New(100, time.Second)

	blocks := []string{
		"google.com",
		"google.co.uk",
		"pleroma.bad.host",
	}

	loader := func() ([]string, error) {
		t.Log("load: returning blocked domains")
		return blocks, nil
	}

	// Check a list of known blocked domains.
	for _, domain := range []string{
		"google.com",
		"mail.google.com",
		"google.co.uk",
		"mail.google.co.uk",
		"pleroma.bad.host",
		"dev.pleroma.bad.host",
	} {
		t.Logf("checking domain is blocked: %s", domain)
		if b, _ := c.IsBlocked(domain, loader); !b {
			t.Errorf("domain should be blocked: %s", domain)
		}
	}

	// Check a list of known unblocked domains.
	for _, domain := range []string{
		"askjeeves.com",
		"ask-kim.co.uk",
		"google.ie",
		"mail.google.ie",
		"gts.bad.host",
		"mastodon.bad.host",
	} {
		t.Logf("checking domain isn't blocked: %s", domain)
		if b, _ := c.IsBlocked(domain, loader); b {
			t.Errorf("domain should not be blocked: %s", domain)
		}
	}

	// Clear the cache
	c.Clear()

	knownErr := errors.New("known error")

	// Check that reload is actually performed and returns our error
	if _, err := c.IsBlocked("", func() ([]string, error) {
		t.Log("load: returning known error")
		return nil, knownErr
	}); !errors.Is(err, knownErr) {
		t.Errorf("is blocked did not return expected error: %v", err)
	}
}

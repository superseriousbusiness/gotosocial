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

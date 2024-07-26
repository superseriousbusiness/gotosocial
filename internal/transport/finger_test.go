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

package transport_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type FingerTestSuite struct {
	TransportTestSuite
}

func (suite *FingerTestSuite) TestFinger() {
	wc := suite.state.Caches.Webfinger
	suite.Equal(0, wc.Len(), "expect webfinger cache to be empty")

	_, err := suite.transport.Finger(context.TODO(), "brand_new_person", "unknown-instance.com")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(0, wc.Len(), "expect webfinger cache to be empty for normal webfinger request")
}

func (suite *FingerTestSuite) TestFingerPunycode() {
	wc := suite.state.Caches.Webfinger
	suite.Equal(0, wc.Len(), "expect webfinger cache to be empty")

	_, err := suite.transport.Finger(context.TODO(), "brand_new_person", "pünycöde.example.org")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(0, wc.Len(), "expect webfinger cache to be empty for normal webfinger request")
}

func (suite *FingerTestSuite) TestFingerWithHostMeta() {
	wc := suite.state.Caches.Webfinger
	suite.Equal(0, wc.Len(), "expect webfinger cache to be empty")

	_, err := suite.transport.Finger(context.TODO(), "someone", "misconfigured-instance.com")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(1, wc.Len(), "expect webfinger cache to hold one entry")
	suite.True(wc.Has("misconfigured-instance.com"), "expect webfinger cache to have entry for misconfigured-instance.com")
}

func (suite *FingerTestSuite) TestFingerWithHostMetaCacheStrategy() {
	if os.Getenv("CI") == "true" {
		suite.T().Skip("this test is flaky on CI for as of yet unknown reasons")
	}

	wc := suite.state.Caches.Webfinger

	// Reset the sweep frequency so nothing interferes with the test
	wc.Stop()
	wc.SetTTL(1*time.Hour, false)
	wc.Start(1 * time.Hour)

	suite.Equal(0, wc.Len(), "expect webfinger cache to be empty")

	_, err := suite.transport.Finger(context.TODO(), "someone", "misconfigured-instance.com")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(1, wc.Len(), "expect webfinger cache to hold one entry")
	wc.Lock()
	suite.True(wc.Cache.Has("misconfigured-instance.com"), "expect webfinger cache to have entry for misconfigured-instance.com")
	ent, _ := wc.Cache.Get("misconfigured-instance.com")
	wc.Unlock()

	initialTime := ent.Expiry

	// finger them again
	_, err = suite.transport.Finger(context.TODO(), "someone", "misconfigured-instance.com")
	if err != nil {
		suite.FailNow(err.Error())
	}

	// there should still only be 1 cache entry
	suite.Equal(1, wc.Len(), "expect webfinger cache to hold one entry")
	wc.Lock()
	suite.True(wc.Cache.Has("misconfigured-instance.com"), "expect webfinger cache to have entry for misconfigured-instance.com")
	rep, _ := wc.Cache.Get("misconfigured-instance.com")
	wc.Unlock()

	repeatTime := rep.Expiry

	// the TTL of the entry should have extended because we did a second
	// successful finger
	if repeatTime == initialTime {
		suite.FailNowf("expected webfinger cache entry to have different expiry times", "initial: '%s', repeat: '%s'", initialTime, repeatTime)
	} else if repeatTime < initialTime {
		suite.FailNowf("expected webfinger cache entry to not be a time traveller", "initial: '%s', repeat: '%s'", initialTime, repeatTime)
	}

	// finger a non-existing user on that same instance which will return an error
	_, err = suite.transport.Finger(context.TODO(), "invalid", "misconfigured-instance.com")
	if err == nil {
		suite.FailNow("expected request for invalid user to fail")
	}

	// there should still only be 1 cache entry, because we don't evict from cache on failure
	suite.Equal(1, wc.Len(), "expect webfinger cache to hold one entry")
	wc.Lock()
	suite.True(wc.Cache.Has("misconfigured-instance.com"), "expect webfinger cache to have entry for misconfigured-instance.com")
	last, _ := wc.Cache.Get("misconfigured-instance.com")
	wc.Unlock()

	lastTime := last.Expiry

	// The TTL of the previous and new entry should be the same since
	// a failed request must not extend the entry TTL
	suite.Equal(repeatTime, lastTime)
}

func TestFingerTestSuite(t *testing.T) {
	suite.Run(t, &FingerTestSuite{})
}

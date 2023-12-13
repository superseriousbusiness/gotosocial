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

package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

func TestHeaderFilter(t *testing.T) {
	testrig.InitTestLog()
	testrig.InitTestConfig()

	for _, test := range []struct {
		mode   string
		allow  []headerfilter
		block  []headerfilter
		input  http.Header
		expect bool
	}{
		{
			mode: config.RequestHeaderFilterModeAllow,
			allow: []headerfilter{
				{"User-Agent", ".*Firefox.*"},
			},
			block: []headerfilter{},
			input: http.Header{
				"User-Agent": []string{"Firefox v169.42; Extra Tracking Info"},
			},
			expect: true,
		},
		{
			mode: config.RequestHeaderFilterModeAllow,
			allow: []headerfilter{
				{"User-Agent", ".*Firefox.*"},
			},
			block: []headerfilter{},
			input: http.Header{
				"User-Agent": []string{"Chromium v169.42; Extra Tracking Info"},
			},
			expect: false,
		},
		{
			mode:  config.RequestHeaderFilterModeBlock,
			allow: []headerfilter{},
			block: []headerfilter{
				{"User-Agent", ".*Firefox.*"},
			},
			input: http.Header{
				"User-Agent": []string{"Firefox v169.42; Extra Tracking Info"},
			},
			expect: false,
		},
		{
			mode:  config.RequestHeaderFilterModeBlock,
			allow: []headerfilter{},
			block: []headerfilter{
				{"User-Agent", ".*Firefox.*"},
			},
			input: http.Header{
				"User-Agent": []string{"Chromium v169.42; Extra Tracking Info"},
			},
			expect: true,
		},
		{
			mode: config.RequestHeaderFilterModeDisabled,
			allow: []headerfilter{
				{"Key1", "only-this"},
				{"Key2", "only-this"},
				{"Key3", "only-this"},
			},
			block: []headerfilter{
				{"Key1", "Value"},
				{"Key2", "Value"},
				{"Key3", "Value"},
			},
			input: http.Header{
				"Key1": []string{"Value"},
				"Key2": []string{"Value"},
				"Key3": []string{"Value"},
			},
			expect: true,
		},
	} {
		// Generate a unique name for this test case.
		name := fmt.Sprintf("%s allow=%v block=%v => expect=%v",
			test.mode,
			test.allow,
			test.block,
			test.expect,
		)

		// Update header filter mode to test case.
		config.SetRequestHeaderFilterMode(test.mode)

		// Run this particular test case.
		ok := t.Run(name, func(t *testing.T) {
			testHeaderFilter(t,
				test.allow,
				test.block,
				test.input,
				test.expect,
			)
		})

		if !ok {
			return
		}
	}
}

func testHeaderFilter(t *testing.T, allow, block []headerfilter, input http.Header, expect bool) {
	var err error

	// Create test context with cancel.
	ctx := context.Background()
	ctx, cncl := context.WithCancel(ctx)
	defer cncl()

	// Initialize caches.
	var state state.State
	state.Caches.Init()

	// Create new database instance with test config.
	state.DB, err = bundb.NewBunDBService(ctx, &state)
	if err != nil {
		t.Fatalf("error opening database: %v", err)
	}

	// Insert all allow filters into DB.
	for _, filter := range allow {
		filter := &gtsmodel.HeaderFilter{
			ID:       id.NewULID(),
			Header:   filter.header,
			Regex:    filter.regex,
			AuthorID: "admin-id",
			Author:   nil,
		}

		if err := state.DB.PutAllowHeaderFilter(ctx, filter); err != nil {
			t.Fatalf("error inserting allow filter into database: %v", err)
		}
	}

	// Insert all block filters into DB.
	for _, filter := range block {
		filter := &gtsmodel.HeaderFilter{
			ID:       id.NewULID(),
			Header:   filter.header,
			Regex:    filter.regex,
			AuthorID: "admin-id",
			Author:   nil,
		}

		if err := state.DB.PutBlockHeaderFilter(ctx, filter); err != nil {
			t.Fatalf("error inserting block filter into database: %v", err)
		}
	}

	// Gin test http engine
	// (used for ctx init).
	e := gin.New()

	// Create new filter middleware to test against.
	middleware := middleware.HeaderFilter(&state)
	e.Use(middleware)

	// Set the empty gin handler (always returns okay).
	e.Handle("GET", "/", func(ctx *gin.Context) { ctx.Status(200) })

	// Prepare a gin test context.
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	// Set input headers.
	r.Header = input

	// Pass req through
	// engine handler.
	e.ServeHTTP(rw, r)

	// Get http result.
	res := rw.Result()

	switch {
	case expect && res.StatusCode != http.StatusOK:
		t.Errorf("unexpected response (should allow): %s", res.Status)

	case !expect && res.StatusCode != http.StatusForbidden:
		t.Errorf("unexpected response (should block): %s", res.Status)
	}
}

type headerfilter struct {
	header string
	regex  string
}

func (hf *headerfilter) String() string {
	return fmt.Sprintf("%s=%q", hf.header, hf.regex)
}

func blockingHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		<-ctx.Done()
		ctx.Status(201) // specifically not 200
	}
}

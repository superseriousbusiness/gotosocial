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
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/middleware"
	"code.superseriousbusiness.org/gotosocial/internal/router"
	"codeberg.org/gruf/go-byteutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNoLLaMasMiddleware(t *testing.T) {
	// Gin test engine.
	e := gin.New()

	// Setup necessary configuration variables.
	config.SetAdvancedScraperDeterrenceEnabled(true)
	config.SetWebTemplateBaseDir("../../web/template")

	// Load templates into engine.
	err := router.LoadTemplates(e)
	assert.NoError(t, err)

	// Add middleware to the gin engine handler stack.
	middleware := middleware.NoLLaMas(apiutil.CookiePolicy{}, getInstanceV1)
	e.Use(middleware)

	// Set test handler we can
	// easily check if was used.
	e.Handle("GET", "/", testHandler)

	// Test with differing user-agents.
	for _, userAgent := range []string{
		"CURL",
		"Mozilla FireSox",
		"Google Gnome",
	} {
		testNoLLaMasMiddleware(t, e, userAgent)
	}
}

func testNoLLaMasMiddleware(t *testing.T, e *gin.Engine, userAgent string) {
	// Prepare a test request for gin engine.
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("User-Agent", userAgent)
	rw := httptest.NewRecorder()

	// Pass req through
	// engine handler.
	e.ServeHTTP(rw, r)

	// Get http result.
	res := rw.Result()

	// It should have been stopped
	// by middleware and NOT used
	// the expected test handler.
	ok := usedTestHandler(res)
	assert.False(t, ok)

	// Read entire response body.
	b, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var seed string
	var challenge string

	// Parse output body and find the challenge / difficulty.
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "data-nollamas-seed=\""):
			line = line[20:]
			line = line[:len(line)-1]
			seed = line
		case strings.HasPrefix(line, "data-nollamas-challenge=\""):
			line = line[25:]
			line = line[:len(line)-1]
			challenge = line
		}
	}

	// Ensure valid posed challenge.
	assert.NotEmpty(t, challenge)
	assert.NotEmpty(t, seed)

	// Prepare a test request for gin engine.
	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Set("User-Agent", userAgent)
	rw = httptest.NewRecorder()

	t.Logf("seed=%s", seed)
	t.Logf("challenge=%s", challenge)

	// Now compute and set solution query paramater.
	solution := computeSolution(seed, challenge)
	r.URL.RawQuery = "nollamas_solution=" + solution
	t.Logf("solution=%s", solution)

	// Pass req through
	// engine handler.
	e.ServeHTTP(rw, r)

	// Get http result.
	res = rw.Result()

	// Should have received redirect.
	uri, err := res.Location()
	assert.NoError(t, err)
	assert.Equal(t, uri.String(), "/")

	// Ensure our expected solution cookie (to bypass challenge) was set.
	ok = slices.ContainsFunc(res.Cookies(), func(c *http.Cookie) bool {
		return c.Name == "gts-nollamas"
	})
	assert.True(t, ok)
}

// computeSolution does the functional equivalent of our nollamas workerTask.js.
func computeSolution(seed, challenge string) string {
	for i := 0; ; i++ {
		solution := strconv.Itoa(i)
		combined := seed + solution
		hash := sha256.Sum256(byteutil.S2B(combined))
		encoded := hex.EncodeToString(hash[:])
		if encoded != challenge {
			continue
		}
		return solution
	}
}

// usedTestHandler returns whether testHandler() was used.
func usedTestHandler(res *http.Response) bool {
	return res.Header.Get("test-handler") == "ok"
}

func testHandler(c *gin.Context) {
	c.Writer.Header().Set("test-handler", "ok")
	c.Writer.WriteHeader(http.StatusOK)
}

func getInstanceV1(context.Context) (*model.InstanceV1, gtserror.WithCode) {
	return &model.InstanceV1{}, nil
}

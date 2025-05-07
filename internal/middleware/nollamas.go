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

package middleware

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"hash"
	"io"
	"net/http"
	"time"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"codeberg.org/gruf/go-byteutil"
	"github.com/gin-gonic/gin"
)

// NoLLaMas returns a piece of HTTP middleware that provides a deterrence
// on routes it is applied to, against bots and scrapers. It generates a
// unique but deterministic challenge for each HTTP client within an hour
// TTL that requires a proof-of-work solution to pass onto the next handler.
// On successful solution, the client is provided a cookie that allows them
// to bypass this check within that hour TTL. The outcome of this is that it
// should make scraping of these endpoints economically unfeasible, when enabled,
// and with an absurdly minimal performance impact. The downside is that it
// requires javascript to be enabled on the client to pass the middleware check.
//
// Heavily inspired by: https://github.com/TecharoHQ/anubis
func NoLLaMas(
	cookiePolicy apiutil.CookiePolicy,
	getInstanceV1 func(context.Context) (*apimodel.InstanceV1, gtserror.WithCode),
) gin.HandlerFunc {

	if !config.GetAdvancedScraperDeterrenceEnabled() {
		// NoLLaMas middleware disabled.
		return func(*gin.Context) {}
	}

	seed := make([]byte, 32)

	// Read random data for the token seed.
	_, err := io.ReadFull(rand.Reader, seed)
	if err != nil {
		panic(err)
	}

	// Configure nollamas.
	var nollamas nollamas
	nollamas.seed = seed
	nollamas.ttl = time.Hour
	nollamas.diff = config.GetAdvancedScraperDeterrenceDifficulty()
	nollamas.getInstanceV1 = getInstanceV1
	nollamas.policy = cookiePolicy
	return nollamas.Serve
}

// hashWithBufs encompasses a hash along
// with the necessary buffers to generate
// a hashsum and then encode that sum.
type hashWithBufs struct {
	hash hash.Hash
	hbuf []byte
	ebuf []byte
}

type nollamas struct {
	// our instance cookie policy.
	policy apiutil.CookiePolicy

	// unique token seed
	// to prevent hashes
	// being guessable
	seed []byte

	// success cookie TTL
	ttl time.Duration

	// algorithm difficulty knobs.
	// diff determines the number
	// of leading zeroes required.
	diff uint8

	// extra fields required for
	// our template rendering.
	getInstanceV1 func(ctx context.Context) (*apimodel.InstanceV1, gtserror.WithCode)
}

func (m *nollamas) Serve(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		// Only interested in protecting
		// crawlable 'GET' endpoints.
		c.Next()
		return
	}

	// Extract request context.
	ctx := c.Request.Context()

	if ctx.Value(oauth.SessionAuthorizedToken) != nil {
		// Don't guard against requests
		// providing valid OAuth tokens.
		c.Next()
		return
	}

	if gtscontext.HTTPSignature(ctx) != "" {
		// Don't guard against requests
		// providing HTTP signatures.
		c.Next()
		return
	}

	// i.e. outputted hash slice length.
	const hashLen = sha256.Size

	// i.e. hex.EncodedLen(hashLen).
	const encodedHashLen = 2 * hashLen

	// Prepare hash + buffers.
	hash := hashWithBufs{
		hash: sha256.New(),
		hbuf: make([]byte, 0, hashLen),
		ebuf: make([]byte, encodedHashLen),
	}

	// Extract client fingerprint data.
	userAgent := c.GetHeader("User-Agent")
	clientIP := c.ClientIP()

	// Generate a unique token for this request,
	// only valid for a period of now +- m.ttl.
	token := m.token(&hash, userAgent, clientIP)

	// For unique challenge string just use a
	// single portion of their 'success' token.
	// SHA256 is not yet cracked, this is not an
	// application of a hash requiring serious
	// cryptographic security and it rotates on
	// a TTL basis, so it should be fine.
	challenge := token[:len(token)/4]

	// Check for a provided success token.
	cookie, _ := c.Cookie("gts-nollamas")

	// Check whether passed cookie
	// is the expected success token.
	if subtle.ConstantTimeCompare(
		byteutil.S2B(token),
		byteutil.S2B(cookie),
	) == 1 {

		// They passed us a valid, expected
		// token. They already passed checks.
		c.Next()
		return
	}

	// From here-on out, all
	// possibilities are handled
	// by us. Prevent further http
	// handlers from being called.
	c.Abort()

	// Prepare new log entry.
	l := log.WithContext(ctx).
		WithField("userAgent", userAgent).
		WithField("challenge", challenge)

	// Extract and parse query.
	query := c.Request.URL.Query()

	// Check query to see if an in-progress
	// challenge solution has been provided.
	nonce := query.Get("nollamas_solution")
	if nonce == "" || len(nonce) > 20 {

		// noting that here, 20 is
		// max integer string len.
		//
		// An invalid solution string, just
		// present them with new challenge.
		l.Info("posing new challenge")
		m.renderChallenge(c, challenge)
		return
	}

	// Reset the hash.
	hash.hash.Reset()

	// Check challenge+nonce as possible solution.
	if !m.checkChallenge(&hash, challenge, nonce) {

		// They failed challenge,
		// re-present challenge page.
		l.Info("invalid solution provided")
		m.renderChallenge(c, challenge)
		return
	}

	l.Infof("challenge passed: %s", nonce)

	// Drop solution query and encode.
	query.Del("nollamas_solution")
	c.Request.URL.RawQuery = query.Encode()

	// They passed the challenge! Set success token
	// cookie and allow them to continue to next handlers.
	m.policy.SetCookie(c, "gts-nollamas", token, int(m.ttl/time.Second), "/")
	c.Redirect(http.StatusTemporaryRedirect, c.Request.URL.RequestURI())
}

func (m *nollamas) renderChallenge(c *gin.Context, challenge string) {
	// Fetch current instance information for templating vars.
	instance, errWithCode := m.getInstanceV1(c.Request.Context())
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.getInstanceV1)
		return
	}

	// Write templated challenge response to client.
	apiutil.TemplateWebPage(c, apiutil.WebPage{
		Template: "nollamas.tmpl",
		Instance: instance,
		Stylesheets: []string{
			"/assets/dist/nollamas.css",
			// Include fork-awesome stylesheet
			// to get nice loading spinner.
			"/assets/Fork-Awesome/css/fork-awesome.min.css",
		},
		Extra: map[string]any{
			"challenge":  challenge,
			"difficulty": m.diff,
		},
		Javascript: []apiutil.JavascriptEntry{
			{
				Src:   "/assets/dist/nollamas.js",
				Defer: true,
			},
		},
	})
}

func (m *nollamas) token(hash *hashWithBufs, userAgent, clientIP string) string {
	// Use our unique seed to seed hash,
	// to ensure we have cryptographically
	// unique, yet deterministic, tokens
	// generated for a given http client.
	hash.hash.Write(m.seed)

	// Include difficulty level in
	// hash input data so if config
	// changes then token invalidates.
	hash.hash.Write([]byte{m.diff})

	// Also seed the generated input with
	// current time rounded to TTL, so our
	// single comparison handles expiries.
	now := time.Now().Round(m.ttl).Unix()
	hash.hash.Write([]byte{
		byte(now >> 56),
		byte(now >> 48),
		byte(now >> 40),
		byte(now >> 32),
		byte(now >> 24),
		byte(now >> 16),
		byte(now >> 8),
		byte(now),
	})

	// Finally, append unique client request data.
	hash.hash.Write(byteutil.S2B(userAgent))
	hash.hash.Write(byteutil.S2B(clientIP))

	// Return hex encoded hash output.
	hash.hbuf = hash.hash.Sum(hash.hbuf[:0])
	hex.Encode(hash.ebuf, hash.hbuf)
	return string(hash.ebuf)
}

func (m *nollamas) checkChallenge(hash *hashWithBufs, challenge, nonce string) bool {
	// Hash and encode input challenge with
	// proposed nonce as a possible solution.
	hash.hash.Write(byteutil.S2B(challenge))
	hash.hash.Write(byteutil.S2B(nonce))
	hash.hbuf = hash.hash.Sum(hash.hbuf[:0])
	hex.Encode(hash.ebuf, hash.hbuf)
	solution := hash.ebuf

	// Compiler bound-check hint.
	if len(solution) < int(m.diff) {
		panic(gtserror.New("BCE"))
	}

	// Check that the first 'diff'
	// many chars are indeed zeroes.
	for i := range m.diff {
		if solution[i] != '0' {
			return false
		}
	}

	return true
}

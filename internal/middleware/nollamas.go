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
	"strconv"
	"time"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"codeberg.org/gruf/go-bitutil"
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

	var seed [32]byte

	// Read random data for the token seed.
	_, err := io.ReadFull(rand.Reader, seed[:])
	if err != nil {
		panic(err)
	}

	// Configure nollamas.
	var nollamas nollamas
	nollamas.entropy = seed
	nollamas.ttl = time.Hour
	nollamas.rounds = config.GetAdvancedScraperDeterrenceDifficulty()
	nollamas.getInstanceV1 = getInstanceV1
	nollamas.policy = cookiePolicy
	return nollamas.Serve
}

// i.e. hash slice length.
const hashLen = sha256.Size

// i.e. hex.EncodedLen(hashLen).
const encodedHashLen = 2 * hashLen

// hashWithBufs encompasses a hash along
// with the necessary buffers to generate
// a hashsum and then encode that sum.
type hashWithBufs struct {
	hash hash.Hash
	hbuf [hashLen]byte
	ebuf [encodedHashLen]byte
}

// write is a passthrough to hash.Hash{}.Write().
func (h *hashWithBufs) write(b []byte) {
	_, _ = h.hash.Write(b)
}

// writeString is a passthrough to hash.Hash{}.Write([]byte(s)).
func (h *hashWithBufs) writeString(s string) {
	_, _ = h.hash.Write(byteutil.S2B(s))
}

// EncodedSum returns the hex encoded sum of hash.Sum().
func (h *hashWithBufs) EncodedSum() string {
	_ = h.hash.Sum(h.hbuf[:0])
	hex.Encode(h.ebuf[:], h.hbuf[:])
	return string(h.ebuf[:])
}

// Reset will reset hash and buffers.
func (h *hashWithBufs) Reset() {
	h.ebuf = [encodedHashLen]byte{}
	h.hbuf = [hashLen]byte{}
	h.hash.Reset()
}

type nollamas struct {
	// our instance cookie policy.
	policy apiutil.CookiePolicy

	// unique entropy
	// to prevent hashes
	// being guessable
	entropy [32]byte

	// success cookie TTL
	ttl time.Duration

	// rounds determines roughly how
	// many hash-encode rounds each
	// client is required to complete.
	rounds uint32

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

	// Prepare new hash with buffers.
	hash := hashWithBufs{hash: sha256.New()}

	// Extract client fingerprint data.
	userAgent := c.GetHeader("User-Agent")
	clientIP := c.ClientIP()

	// Generate a unique token for this request,
	// only valid for a period of now +- m.ttl.
	token := m.getToken(&hash, userAgent, clientIP)

	// Check for a provided success token.
	cookie, _ := c.Cookie("gts-nollamas")

	// Check whether passed cookie
	// is the expected success token.
	if subtle.ConstantTimeCompare(
		byteutil.S2B(cookie),
		byteutil.S2B(token),
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

	// Generate challenge for this unique (yet deterministic) token,
	// returning seed, wanted 'challenge' result and expected solution.
	seed, challenge, solution := m.getChallenge(&hash, token)

	// Prepare new log entry.
	l := log.WithContext(ctx).
		WithField("userAgent", userAgent).
		WithField("seed", seed).
		WithField("rounds", solution)

	// Extract and parse query.
	query := c.Request.URL.Query()

	// Check query to see if an in-progress
	// challenge solution has been provided.
	nonce := query.Get("nollamas_solution")
	if nonce == "" {

		// No solution given, likely new client!
		// Simply present them with challenge.
		m.renderChallenge(c, seed, challenge)
		l.Info("posing new challenge")
		return
	}

	// Check nonce matches expected.
	if subtle.ConstantTimeCompare(
		byteutil.S2B(solution),
		byteutil.S2B(nonce),
	) != 1 {

		// Their nonce failed, re-challenge them.
		m.renderChallenge(c, challenge, solution)
		l.Infof("invalid solution provided: %s", nonce)
		return
	}

	l.Info("challenge passed")

	// Drop solution query and encode.
	query.Del("nollamas_solution")
	c.Request.URL.RawQuery = query.Encode()

	// They passed the challenge! Set success token
	// cookie and allow them to continue to next handlers.
	m.policy.SetCookie(c, "gts-nollamas", token, int(m.ttl/time.Second), "/")
	c.Redirect(http.StatusTemporaryRedirect, c.Request.URL.RequestURI())
}

func (m *nollamas) renderChallenge(c *gin.Context, seed, challenge string) {
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
			"seed":      seed,
			"challenge": challenge,
		},
		Javascript: []apiutil.JavascriptEntry{
			{
				Src:   "/assets/dist/nollamas.js",
				Defer: true,
			},
		},
	})
}

// getToken generates a unique yet deterministic token for given HTTP request
// details, seeded by runtime generated entropy data and ttl rounded timestamp.
func (m *nollamas) getToken(hash *hashWithBufs, userAgent, clientIP string) string {

	// Reset before
	// using hash.
	hash.Reset()

	// Use our unique entropy to seed hash,
	// to ensure we have cryptographically
	// unique, yet deterministic, tokens
	// generated for a given http client.
	hash.write(m.entropy[:])

	// Also seed the generated input with
	// current time rounded to TTL, so our
	// single comparison handles expiries.
	now := time.Now().Round(m.ttl).Unix()
	hash.write([]byte{
		byte(now >> 56),
		byte(now >> 48),
		byte(now >> 40),
		byte(now >> 32),
		byte(now >> 24),
		byte(now >> 16),
		byte(now >> 8),
		byte(now),
	})

	// Append client request data.
	hash.writeString(userAgent)
	hash.writeString(clientIP)

	// Return hex encoded hash.
	return hash.EncodedSum()
}

// getChallenge prepares a new challenge given the deterministic input token for this request.
// it will return an input seed string, a challenge string which is the end result the client
// should be looking for, and the solution for this such that challenge = hex(sha256(seed + solution)).
// the solution will always be a string-encoded 64bit integer calculated from m.rounds + random jitter.
func (m *nollamas) getChallenge(hash *hashWithBufs, token string) (seed, challenge, solution string) {

	// For their unique seed string just use a
	// single portion of their 'success' token.
	// SHA256 is not yet cracked, this is not an
	// application of a hash requiring serious
	// cryptographic security and it rotates on
	// a TTL basis, so it should be fine.
	seed = token[:len(token)/4]

	// BEFORE resetting the hash, get the last
	// two bytes of NON-hex-encoded data from
	// token generation to use for random jitter.
	// This is taken from the end of the hash as
	// this is the "unseen" end part of token.
	//
	// (if we used hex-encoded data it would
	// only ever be '0-9' or 'a-z' ASCII chars).
	//
	// Security-wise, same applies as-above.
	jitter := int16(hash.hbuf[len(hash.hbuf)-2]) |
		int16(hash.hbuf[len(hash.hbuf)-1])<<8

	var rounds int64
	switch {
	// For some small percentage of
	// clients we purposely low-ball
	// their rounds required, to make
	// it so gaming it with a starting
	// nonce value may suddenly fail.
	case jitter%37 == 0:
		rounds = int64(m.rounds/10) + int64(jitter/10)
	case jitter%31 == 0:
		rounds = int64(m.rounds/5) + int64(jitter/5)
	case jitter%29 == 0:
		rounds = int64(m.rounds/3) + int64(jitter/3)
	case jitter%13 == 0:
		rounds = int64(m.rounds/2) + int64(jitter/2)

	// Determine an appropriate number of hash rounds
	// we want the client to perform on input seed. This
	// is determined as configured m.rounds +- jitter.
	// This will be the 'solution' to create 'challenge'.
	default:
		rounds = int64(m.rounds) + int64(jitter) //nolint:gosec
	}

	// Encode (positive) determined hash rounds as string.
	solution = strconv.FormatInt(bitutil.Abs64(rounds), 10)

	// Reset before
	// using hash.
	hash.Reset()

	// Calculate the expected result
	// of hex(sha256(seed + solution)),
	// i.e. the proposed 'challenge'.
	hash.writeString(seed)
	hash.writeString(solution)
	challenge = hash.EncodedSum()

	return
}

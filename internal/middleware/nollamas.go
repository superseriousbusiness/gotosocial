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
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"crypto/x509"
	"encoding/hex"
	"hash"
	"net/http"
	"time"

	"codeberg.org/gruf/go-byteutil"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func NoLLaMas(instanceAcc *gtsmodel.Account) gin.HandlerFunc {
	// Generate seed hash from
	// this instance private key.
	priv := instanceAcc.PrivateKey
	bpriv := x509.MarshalPKCS1PrivateKey(priv)
	seed := sha512.Sum512(bpriv)

	// Configure nollamas.
	var nollamas nollamas
	nollamas.seed = seed[:]
	nollamas.ttl = time.Hour
	nollamas.diff = 4
	return nollamas.Serve
}

// i.e. outputted hash slice length.
const hashLen = sha256.BlockSize

// i.e. hex.EncodedLen(hashLen).
const encodedHashLen = 2 * hashLen

func newHash() hash.Hash { return sha256.New() }

type nollamas struct {
	seed []byte // securely hashed instance private key
	ttl  time.Duration
	diff uint8
}

func (m *nollamas) Serve(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		// Only interested in protecting
		// crawlable 'GET' endpoints.
		c.Next()
		return
	}

	if _, ok := c.Get(oauth.SessionAuthorizedToken); ok {
		// Don't guard against requests
		// providing valid OAuth tokens.
		c.Next()
		return
	}

	// Get new hasher.
	hash := newHash()

	// Reset hash.
	hash.Reset()

	// Generate a unique token for
	// this request only valid for
	// a period of now +- m.ttl.
	token := m.token(c, hash)

	// For unique challenge string just use a
	// portion of their unique 'success' token.
	// SHA256 is not yet cracked, this is not an
	// application of a hash requiring serious
	// cryptographic security and it rotates on
	// a TTL basis, so it should be fine.
	challenge := token[:len(token)/2]

	// Check for a provided success token.
	cookie, _ := c.Cookie("gts-nollamas")

	if len(cookie) > encodedHashLen {
		// Clearly invalid cookie, just
		// present them with new challenge.
		m.renderChallenge(c, challenge)
		return
	}

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

	// Check query to see if an in-progress
	// challenge solution has been provided.
	query := c.Request.URL.Query()
	nonce := query.Get("nollamas_solution")
	if nonce == "" || len(nonce) > 20 {

		// No attempted solution, just
		// present them with new challenge.
		m.renderChallenge(c, challenge)
		return
	}

	// Reset hash.
	hash.Reset()

	// Hash and encode input challenge with
	// proposed nonce as a possible solution.
	_, _ = hash.Write(byteutil.S2B(challenge))
	_, _ = hash.Write(byteutil.S2B(nonce))
	solution := hex.AppendEncode(nil, hash.Sum(nil))

	// Check that the first 'diff'
	// many chars are indeed zeroes.
	for i := range m.diff {
		if solution[i] != '0' {

			// They failed challenge,
			// re-present challenge page.
			m.renderChallenge(c, challenge)
			return
		}
	}

	// Drop the solution from query.
	query.Del("nollamas_solution")
	c.Request.URL.RawQuery = query.Encode()

	// They passed the challenge! Set success
	// token cookie and allow them to continue.
	c.SetCookie("gts-nollamas", token, int(m.ttl/time.Second),
		"", "", false, false)
	c.Next()
}

func (m *nollamas) renderChallenge(c *gin.Context, challenge string) {
	// Don't pass to further
	// handlers, they only get
	// our challenge page.
	c.Abort()

	// Write the templated challenge HTML response to client.
	c.HTML(http.StatusOK, "nollamas.tmpl", map[string]any{
		"challenge":  challenge,
		"difficulty": m.diff,
	})
}

func (m *nollamas) token(c *gin.Context, hash hash.Hash) string {
	// Use our safe, unique input seed which
	// is already hashed, but will get rehashed.
	// This ensures we don't leak private keys,
	// but also we have cryptographically safe
	// deterministic tokens for comparisons.
	_, _ = hash.Write(m.seed)

	// Include difficulty level in
	// hash input data so if config
	// changes then token invalidates.
	_, _ = hash.Write([]byte{m.diff})

	// Also seed the generated input with
	// current time rounded to TTL, so with
	// our single comparison handles expiries.
	now := time.Now().Round(m.ttl).Unix()
	_, _ = hash.Write([]byte{
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
	userAgent := c.Request.Header.Get("User-Agent")
	_, _ = hash.Write(byteutil.S2B(userAgent))
	clientIP := c.ClientIP()
	_, _ = hash.Write(byteutil.S2B(clientIP))

	// Return hex encoded hash output.
	return hex.EncodeToString(hash.Sum(nil))
}

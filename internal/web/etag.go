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

package web

import (
	// nolint:gosec
	"crypto/sha1"
	"encoding/hex"
	"io"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/log"

	"codeberg.org/gruf/go-cache/v3"
)

func newETagCache() cache.Cache[string, eTagCacheEntry] {
	eTagCache := cache.New[string, eTagCacheEntry](0, 1000, 0)
	eTagCache.SetTTL(time.Hour, false)
	if !eTagCache.Start(time.Minute) {
		log.Panic(nil, "could not start eTagCache")
	}
	return eTagCache
}

type eTagCacheEntry struct {
	eTag         string
	lastModified time.Time
}

// generateEtag generates a strong (byte-for-byte) etag using
// the entirety of the provided reader.
func generateEtag(r io.Reader) (string, error) {
	// nolint:gosec
	hash := sha1.New()

	if _, err := io.Copy(hash, r); err != nil {
		return "", err
	}

	b := make([]byte, 0, sha1.Size)
	b = hash.Sum(b)

	return `"` + hex.EncodeToString(b) + `"`, nil
}

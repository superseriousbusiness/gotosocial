/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type eTagCacheEntry struct {
	eTag             string
	fileLastModified time.Time
}

// generateEtag generates a strong (byte-for-byte) etag using
// the entirety of the provided reader.
func generateEtag(r io.Reader) (string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	// nolint:gosec
	sum := sha1.Sum(b)

	return hex.EncodeToString(sum[:]), nil
}

// getAssetFileInfo tries to fetch the ETag for the given filePath from the module's
// assetsETagCache. If it can't be found there, it uses the provided http.FileSystem
// to generate a new ETag to go in the cache, which it then returns.
func (m *Module) getAssetETag(filePath string, fs http.FileSystem) (string, error) {
	file, err := fs.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening %s: %s", filePath, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("error statting %s: %s", filePath, err)
	}

	fileLastModified := fileInfo.ModTime()

	if cachedETag, ok := m.assetsETagCache.Get(filePath); ok && !fileLastModified.After(cachedETag.fileLastModified) {
		// only return our cached etag if the file wasn't
		// modified since last time, otherwise generate a
		// new one; eat fresh!
		return cachedETag.eTag, nil
	}

	eTag, err := generateEtag(file)
	if err != nil {
		return "", fmt.Errorf("error generating etag: %s", err)
	}

	// put new entry in cache before we return
	m.assetsETagCache.Set(filePath, eTagCacheEntry{
		eTag:             eTag,
		fileLastModified: fileLastModified,
	})

	return eTag, nil
}

// cacheControlMiddleware implements Cache-Control header setting, and checks for
// files inside the given http.FileSystem.
//
// The middleware checks if the file has been modified using If-None-Match etag,
// if present. If the file hasn't been modified, the middleware returns 304.
//
// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-None-Match
// and: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
func (m *Module) cacheControlMiddleware(fs http.FileSystem) gin.HandlerFunc {
	return func(c *gin.Context) {
		// no-cache prevents clients using default caching or heuristic caching,
		// and also ensures that clients will validate their cached version against
		// the version stored on the server to keep up to date.
		c.Header("Cache-Control", "no-cache")

		ifNoneMatch := c.Request.Header.Get("If-None-Match")

		// derive the path of the requested asset inside the provided filesystem
		upath := c.Request.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
		}
		assetFilePath := strings.TrimPrefix(path.Clean(upath), assetsPath)

		// either fetch etag from ttlcache or generate it
		eTag, err := m.getAssetETag(assetFilePath, fs)
		if err != nil {
			logrus.Errorf("error getting ETag for %s: %s", assetFilePath, err)
			return
		}

		// Regardless of what happens further down, set the etag header
		// so that the client has the up-to-date version.
		c.Header("Etag", eTag)

		// If client already has latest version of the asset, 304 + bail.
		if ifNoneMatch == eTag {
			c.AbortWithStatus(http.StatusNotModified)
			return
		}

		// else let the rest of the request be processed normally
	}
}

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
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// generateEtag generates a weak etag by concatenating the given filePath string
// with the unix timestamp of the given time, then doing a sha1 hash on the result.
func generateEtag(filePath string, lastModified time.Time) (string, error) {
	b := []byte(fmt.Sprintf("%s%d", filePath, lastModified.Unix()))

	// nolint:gosec
	hash := sha1.New()

	if _, err := hash.Write(b); err != nil {
		return "", err
	}

	return `/W"` + hex.EncodeToString(hash.Sum(nil)) + `"`, nil
}

// getAssetFileInfo tries to fetch info for the given filePath from the module's
// assetsFileInfoCache. If it can't be found there, it uses the provided http.FileSystem
// to generate a new assetFileInfo entry to go in the cache, which it then returns.
func (m *Module) getAssetFileInfo(filePath string, fs http.FileSystem) (assetFileInfo, error) {
	// return fileinfo from cache directly if we have it
	if cachedFileInfo, ok := m.assetsFileInfoCache.Get(filePath); ok {
		return cachedFileInfo, nil
	}

	// we don't have it, create a new one
	afi := assetFileInfo{}

	file, err := fs.Open(filePath)
	if err != nil {
		return afi, fmt.Errorf("error opening %s: %s", filePath, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return afi, fmt.Errorf("error statting %s: %s", filePath, err)
	}

	afi.lastModified = fileInfo.ModTime()

	etag, err := generateEtag(filePath, afi.lastModified)
	if err != nil {
		return afi, fmt.Errorf("error generating etag: %s", err)
	}
	afi.etag = etag

	// put new entry in cache before we return
	m.assetsFileInfoCache.Set(filePath, afi)
	return afi, nil
}

// cacheControlMiddleware implements Cache-Control header setting, and etag/last-modified
// checks for files inside the given http.FileSystem.
//
// First check if the file has been modified using If-None-Match etag, if present.
// If the file hasn't been modified, bail with a 304.
//
// Then, check if the file has been modified using If-Modified-Since, if present.
// If the file hasn't been modified since the given date, bail with a 304.
// We only do this second check if If-None-Match wasn't set.
//
// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since
// and: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-None-Match
// and: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
func (m *Module) cacheControlMiddleware(fs http.FileSystem) gin.HandlerFunc {
	return func(c *gin.Context) {
		// no-cache prevents clients using default caching or heuristic caching,
		// and also ensures that clients will validate their cached version against
		// the version stored on the server to keep up to date.
		c.Header("Cache-Control", "no-cache")

		// pull some variables out of the request
		ifNoneMatch := c.Request.Header.Get("If-None-Match")
		ifModifiedSinceString := c.Request.Header.Get("If-Modified-Since")

		// derive the path of the requested asset inside the provided filesystem
		upath := c.Request.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
		}
		assetFilePath := strings.TrimPrefix(path.Clean(upath), assetsPath)

		// generate etag/last modified or fetch from ttlcache
		assetFileInfo, err := m.getAssetFileInfo(assetFilePath, fs)
		if err != nil {
			logrus.Errorf("error getting file info for %s: %s", assetFilePath, err)
			return
		}

		// Regardless of what happens further down, set the etag header
		// so that the client has the up-to-date version.
		c.Header("Etag", assetFileInfo.etag)

		// If client already has latest version of the asset, 304 + bail.
		if ifNoneMatch == assetFileInfo.etag {
			c.AbortWithStatus(http.StatusNotModified)
			return
		}

		// only fall back to using If-Modifed-Since if If-None-Match wasn't set.
		if ifNoneMatch == "" && ifModifiedSinceString != "" {
			ifModifiedSince, err := http.ParseTime(ifModifiedSinceString)
			if err != nil {
				logrus.Debugf("If-Modifed-Since header could not be parsed: %s", err)
				return
			}

			// If client already has latest version of the asset, 304 + bail.
			if assetFileInfo.lastModified.Before(ifModifiedSince) {
				c.AbortWithStatus(http.StatusNotModified)
				return
			}
		}

		// if we reach this point, either the file has been modified, or we don't have
		// enough information from the caller to determine caching; either way, let the
		// request proceed as normal
	}
}

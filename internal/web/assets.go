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
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type fileSystem struct {
	fs http.FileSystem
}

// FileSystem server that only accepts directory listings when an index.html is available
// from https://gist.github.com/hauxe/f2ea1901216177ccf9550a1b8bd59178
func (fs fileSystem) Open(path string) (http.File, error) {
	f, err := fs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, _ := f.Stat()
	if s.IsDir() {
		index := strings.TrimSuffix(path, "/") + "/index.html"
		if _, err := fs.fs.Open(index); err != nil {
			return nil, err
		}
	}

	return f, nil
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

	if cachedETag, ok := m.eTagCache.Get(filePath); ok && !fileLastModified.After(cachedETag.lastModified) {
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
	m.eTagCache.Set(filePath, eTagCacheEntry{
		eTag:         eTag,
		lastModified: fileLastModified,
	})

	return eTag, nil
}

// assetsCacheControlMiddleware implements Cache-Control header setting, and checks
// for files inside the given http.FileSystem.
//
// The middleware checks if the file has been modified using If-None-Match etag,
// if present. If the file hasn't been modified, the middleware returns 304.
//
// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-None-Match
// and: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
//
// todo: move this middleware out of the 'web' package and into the 'middleware'
// package along with the other middlewares
func (m *Module) assetsCacheControlMiddleware(fs http.FileSystem) gin.HandlerFunc {
	return func(c *gin.Context) {
		// set this Cache-Control header to instruct clients to validate the response with us
		// before each reuse (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control)
		c.Header(cacheControlHeader, cacheControlNoCache)

		ifNoneMatch := c.Request.Header.Get(ifNoneMatchHeader)

		// derive the path of the requested asset inside the provided filesystem
		upath := c.Request.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
		}
		assetFilePath := strings.TrimPrefix(path.Clean(upath), assetsPathPrefix)

		// either fetch etag from ttlcache or generate it
		eTag, err := m.getAssetETag(assetFilePath, fs)
		if err != nil {
			log.Errorf("error getting ETag for %s: %s", assetFilePath, err)
			return
		}

		// Regardless of what happens further down, set the etag header
		// so that the client has the up-to-date version.
		c.Header(eTagHeader, eTag)

		// If client already has latest version of the asset, 304 + bail.
		if ifNoneMatch == eTag {
			c.AbortWithStatus(http.StatusNotModified)
			return
		}

		// else let the rest of the request be processed normally
	}
}

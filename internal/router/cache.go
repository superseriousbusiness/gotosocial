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

package router

import (
	// nolint:gosec
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// generateEtag generates a weak etag for the given byte slice.
func generateEtag(in []byte) string {
	// nolint:gosec
	sum := sha1.Sum(in)
	etag := fmt.Sprintf(`/W"%d-%x"`, len(in), sum)
	return etag
}

func cacheMiddleware(fs http.FileSystem) gin.HandlerFunc {
	return func(c *gin.Context) {
		// We set the cache-control header in this middleware to avoid clients using
		// default cacheing or heuristic caching.
		//
		// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
		c.Header("Cache-Control", "no-cache")

		// pull some variables out of the request
		upath := c.Request.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
		}
		filePath := path.Clean(upath)
		reqEtag := c.Request.Header.Get("If-None-Match")
		sinceString := c.Request.Header.Get("If-Modified-Since")

		// First check if the file has been modified using If-None-Match etag, if present.
		// If the file hasn't been modified, bail with a 304.
		//
		// Then, check if the file has been modified using If-Modified-Since, if present.
		// If the file hasn't been modified since the given date, bail with a 304.
		// We only do this second check if If-None-Match wasn't set.
		//
		// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since
		// and: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-None-Match

		file, err := fs.Open(strings.TrimPrefix(filePath, "/assets"))
		if err != nil {
			logrus.Errorf("error opening asset: %s", err)
			return
		}
		defer file.Close()

		b, err := io.ReadAll(file)
		if err != nil {
			logrus.Errorf("error reading file: %s", err)
			return
		}

		etag := generateEtag(b)

		// Regardless of what happens further down, set the etag header
		// so that the client has the up-to-date version.
		c.Header("Etag", etag)

		// If the client already has the latest version, we can bail early.
		if reqEtag == etag {
			c.AbortWithStatus(http.StatusNotModified)
			return
		}

		// only fall back to using If-Modifed-Since if If-None-Match wasn't set.
		if reqEtag == "" && sinceString != "" {
			ifModifiedSince, err := http.ParseTime(sinceString)
			if err != nil {
				logrus.Debugf("If-Modifed-Since header could not be parsed: %s", err)
				return
			}

			fileInfo, err := file.Stat()
			if err != nil {
				logrus.Errorf("error statting asset: %s", err)
				return
			}

			lastModified := fileInfo.ModTime()
			if !lastModified.After(ifModifiedSince) {
				c.AbortWithStatus(http.StatusNotModified)
				return
			}
		}

		// if we reach this point, either the file has been modified, or we don't have
		// enough information from the caller to determine caching; either way, let the
		// request proceed as normal
	}
}

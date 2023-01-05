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
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

const appRSSUTF8 = string(apiutil.AppRSSXML + "; charset=utf-8")

func (m *Module) GetRSSETag(urlPath string, lastModified time.Time, getRSSFeed func() (string, gtserror.WithCode)) (string, error) {
	if cachedETag, ok := m.eTagCache.Get(urlPath); ok && !lastModified.After(cachedETag.lastModified) {
		// only return our cached etag if the file wasn't
		// modified since last time, otherwise generate a
		// new one; eat fresh!
		return cachedETag.eTag, nil
	}

	rssFeed, errWithCode := getRSSFeed()
	if errWithCode != nil {
		return "", fmt.Errorf("error getting rss feed: %s", errWithCode)
	}

	eTag, err := generateEtag(bytes.NewReader([]byte(rssFeed)))
	if err != nil {
		return "", fmt.Errorf("error generating etag: %s", err)
	}

	// put new entry in cache before we return
	m.eTagCache.Set(urlPath, eTagCacheEntry{
		eTag:         eTag,
		lastModified: lastModified,
	})

	return eTag, nil
}

func extractIfModifiedSince(header string) time.Time {
	if header == "" {
		return time.Time{}
	}

	t, err := http.ParseTime(header)
	if err != nil {
		log.Errorf("couldn't parse if-modified-since %s: %s", header, err)
		return time.Time{}
	}

	return t
}

func (m *Module) rssFeedGETHandler(c *gin.Context) {
	// set this Cache-Control header to instruct clients to validate the response with us
	// before each reuse (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control)
	c.Header(cacheControlHeader, cacheControlNoCache)
	ctx := c.Request.Context()

	if _, err := apiutil.NegotiateAccept(c, apiutil.AppRSSXML); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	// usernames on our instance will always be lowercase
	username := strings.ToLower(c.Param(usernameKey))
	if username == "" {
		err := errors.New("no account username specified")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	ifNoneMatch := c.Request.Header.Get(ifNoneMatchHeader)
	ifModifiedSince := extractIfModifiedSince(c.Request.Header.Get(ifModifiedSinceHeader))

	getRssFeed, accountLastPostedPublic, errWithCode := m.processor.AccountGetRSSFeedForUsername(ctx, username)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	var rssFeed string
	cacheKey := c.Request.URL.Path
	cacheEntry, ok := m.eTagCache.Get(cacheKey)

	if !ok || cacheEntry.lastModified.Before(accountLastPostedPublic) {
		// we either have no cache entry for this, or we have an expired cache entry; generate a new one
		rssFeed, errWithCode = getRssFeed()
		if errWithCode != nil {
			apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
			return
		}

		eTag, err := generateEtag(bytes.NewBufferString(rssFeed))
		if err != nil {
			apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGet)
			return
		}

		cacheEntry.lastModified = accountLastPostedPublic
		cacheEntry.eTag = eTag
		m.eTagCache.Set(cacheKey, cacheEntry)
	}

	c.Header(eTagHeader, cacheEntry.eTag)
	c.Header(lastModifiedHeader, accountLastPostedPublic.Format(http.TimeFormat))

	if ifNoneMatch == cacheEntry.eTag {
		c.AbortWithStatus(http.StatusNotModified)
		return
	}

	lmUnix := cacheEntry.lastModified.Unix()
	imsUnix := ifModifiedSince.Unix()
	if lmUnix <= imsUnix {
		c.AbortWithStatus(http.StatusNotModified)
		return
	}

	if rssFeed == "" {
		// we had a cache entry already so we didn't call to get the rss feed yet
		rssFeed, errWithCode = getRssFeed()
		if errWithCode != nil {
			apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
			return
		}
	}

	c.Data(http.StatusOK, appRSSUTF8, []byte(rssFeed))
}

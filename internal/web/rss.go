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

package web

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/gin-gonic/gin"
)

const appRSSUTF8 = string(apiutil.AppRSSXML) + "; charset=utf-8"

func (m *Module) rssFeedGETHandler(c *gin.Context) {
	if _, err := apiutil.NegotiateAccept(c, apiutil.AppRSSXML); err != nil {
		apiutil.WebErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	// Fetch + normalize username from URL.
	username, errWithCode := apiutil.ParseUsername(c.Param(apiutil.UsernameKey))
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// Usernames on our instance will always be lowercase.
	//
	// todo: https://codeberg.org/superseriousbusiness/gotosocial/issues/1813
	username = strings.ToLower(username)

	// Retrieve the getRSSFeed function from the processor.
	// We'll only call the function if we need to, to save db calls.
	// lastPostAt may be a zero time if account has never posted.
	getRSSFeed, lastPostAt, errWithCode := m.processor.Account().GetRSSFeedForUsername(c.Request.Context(), username)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	var (
		rssFeed string // Stringified rss feed.

		cacheKey              = c.Request.URL.Path
		cacheEntry, wasCached = m.eTagCache.Get(cacheKey)
	)

	if !wasCached || unixAfter(lastPostAt, cacheEntry.lastModified) {
		// We either have no ETag cache entry for this account's feed,
		// or we have an expired cache entry (account has posted since
		// the cache entry was last generated).
		//
		// As such, we need to generate a new ETag, and for that we need
		// the string representation of the RSS feed.
		rssFeed, errWithCode = getRSSFeed()
		if errWithCode != nil {
			apiutil.WebErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
			return
		}

		eTag, err := generateEtag(bytes.NewBufferString(rssFeed))
		if err != nil {
			apiutil.WebErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGetV1)
			return
		}

		// We never want lastModified to be zero, so if account
		// has never actually posted anything, just use Now as
		// the lastModified time instead for cache control.
		var lastModified time.Time
		if lastPostAt.IsZero() {
			lastModified = time.Now()
		} else {
			lastModified = lastPostAt
		}

		// Store the new cache entry.
		cacheEntry = eTagCacheEntry{
			eTag:         eTag,
			lastModified: lastModified,
		}
		m.eTagCache.Set(cacheKey, cacheEntry)
	}

	// Set 'ETag' and 'Last-Modified' headers no matter what;
	// even if we return 304 in the next checks, caller may
	// want to cache these header values.
	c.Header(eTagHeader, cacheEntry.eTag)
	c.Header(lastModifiedHeader, cacheEntry.lastModified.Format(http.TimeFormat))

	// Instruct caller to validate the response with us before
	// each reuse, so that the 'ETag' and 'Last-Modified' headers
	// actually take effect.
	//
	// "The no-cache response directive indicates that the response
	// can be stored in caches, but the response must be validated
	// with the origin server before each reuse, even when the cache
	// is disconnected from the origin server."
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
	c.Header(cacheControlHeader, cacheControlNoCache)

	// Check if caller submitted an ETag via 'If-None-Match'.
	// If they did + it matches what we have, that means they've
	// already seen the latest version of this feed, so just bail.
	ifNoneMatch := c.Request.Header.Get(ifNoneMatchHeader)
	if ifNoneMatch == cacheEntry.eTag {
		c.AbortWithStatus(http.StatusNotModified)
		return
	}

	// Check if the caller submitted a time via 'If-Modified-Since'.
	// If they did, and our cached ETag entry is not newer than the
	// given time, this means the caller has already seen the latest
	// version of this feed, so just bail.
	ifModifiedSince := extractIfModifiedSince(c.Request)
	if !ifModifiedSince.IsZero() &&
		!unixAfter(cacheEntry.lastModified, ifModifiedSince) {
		c.AbortWithStatus(http.StatusNotModified)
		return
	}

	// At this point we know that the client wants the newest
	// representation of the RSS feed, either because they didn't
	// submit any 'If-None-Match' / 'If-Modified-Since' cache headers,
	// or because they did but the account has posted more recently
	// than the values of the submitted headers would suggest.
	//
	// If we had a cache hit earlier, we may not have called the
	// getRSSFeed function yet; if that's the case then do call it
	// now because we definitely need it.
	if rssFeed == "" {
		rssFeed, errWithCode = getRSSFeed()
		if errWithCode != nil {
			apiutil.WebErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
			return
		}
	}

	c.Data(http.StatusOK, appRSSUTF8, []byte(rssFeed))
}

// unixAfter returns true if the unix value of t1
// is greater than (ie., after) the unix value of t2.
func unixAfter(t1 time.Time, t2 time.Time) bool {
	if t1.IsZero() {
		// if t1 is zero then it cannot
		// possibly be greater than t2.
		return false
	}

	if t2.IsZero() {
		// t1 is not zero but t2 is,
		// so t1 is necessarily greater.
		return true
	}

	return t1.Unix() > t2.Unix()
}

// extractIfModifiedSince parses a time.Time from the
// 'If-Modified-Since' header of the given request.
//
// If no time was provided, or the provided time was
// not parseable, it will return a zero time.
func extractIfModifiedSince(r *http.Request) time.Time {
	imsStr := r.Header.Get(ifModifiedSinceHeader)
	if imsStr == "" {
		return time.Time{} // Nothing set.
	}

	ifModifiedSince, err := http.ParseTime(imsStr)
	if err != nil {
		log.Errorf(r.Context(), "couldn't parse %s value '%s' as time: %q", ifModifiedSinceHeader, imsStr, err)
		return time.Time{}
	}

	return ifModifiedSince
}

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

package fileserver

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"codeberg.org/gruf/go-fastcopy"
	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// ServeFile is for serving attachments, headers, and avatars to the requester from instance storage.
//
// Note: to mitigate scraping attempts, no information should be given out on a bad request except "404 page not found".
// Don't give away account ids or media ids or anything like that; callers shouldn't be able to infer anything.
func (m *Module) ServeFile(c *gin.Context) {
	authed, err := oauth.Authed(c, false, false, false, false)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(err), m.processor.InstanceGetV1)
		return
	}

	// We use request params to check what to pull out of the database/storage so check everything. A request URL should be formatted as follows:
	// "https://example.org/fileserver/[ACCOUNT_ID]/[MEDIA_TYPE]/[MEDIA_SIZE]/[FILE_NAME]"
	// "FILE_NAME" consists of two parts, the attachment's database id, a period, and the file extension.
	accountID := c.Param(AccountIDKey)
	if accountID == "" {
		err := fmt.Errorf("missing %s from request", AccountIDKey)
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(err), m.processor.InstanceGetV1)
		return
	}

	mediaType := c.Param(MediaTypeKey)
	if mediaType == "" {
		err := fmt.Errorf("missing %s from request", MediaTypeKey)
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(err), m.processor.InstanceGetV1)
		return
	}

	mediaSize := c.Param(MediaSizeKey)
	if mediaSize == "" {
		err := fmt.Errorf("missing %s from request", MediaSizeKey)
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(err), m.processor.InstanceGetV1)
		return
	}

	fileName := c.Param(FileNameKey)
	if fileName == "" {
		err := fmt.Errorf("missing %s from request", FileNameKey)
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(err), m.processor.InstanceGetV1)
		return
	}

	// Acquire context from gin request.
	ctx := c.Request.Context()

	content, errWithCode := m.processor.FileGet(ctx, authed, &apimodel.GetContentRequestForm{
		AccountID: accountID,
		MediaType: mediaType,
		MediaSize: mediaSize,
		FileName:  fileName,
	})
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if content.URL != nil {
		// This is a non-local, non-proxied S3 file we're redirecting to.
		// Derive the max-age value from how long the link has left until
		// it expires.
		maxAge := int(time.Until(content.URL.Expiry).Seconds())
		c.Header("Cache-Control", "private,max-age="+strconv.Itoa(maxAge))
		c.Redirect(http.StatusFound, content.URL.String())
		return
	}

	defer func() {
		// Close content when we're done, catch errors.
		if err := content.Content.Close(); err != nil {
			log.Errorf(ctx, "ServeFile: error closing readcloser: %s", err)
		}
	}()

	// TODO: if the requester only accepts text/html we should try to serve them *something*.
	// This is mostly needed because when sharing a link to a gts-hosted file on something like mastodon, the masto servers will
	// attempt to look up the content to provide a preview of the link, and they ask for text/html.
	format, err := apiutil.NegotiateAccept(c, apiutil.MIME(content.ContentType))
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	// if this is a head request, just return info + throw the reader away
	if c.Request.Method == http.MethodHead {
		c.Header("Content-Type", format)
		c.Header("Content-Length", strconv.FormatInt(content.ContentLength, 10))
		c.Status(http.StatusOK)
		return
	}

	// Look for a provided range header.
	rng := c.GetHeader("Range")
	if rng == "" {
		// This is a simple query for the whole file, so do a read from whole reader.
		c.DataFromReader(http.StatusOK, content.ContentLength, format, content.Content, nil)
		return
	}

	// Set known content-type and serve range.
	c.Header("Content-Type", format)
	serveFileRange(
		c.Writer,
		c.Request,
		content.Content,
		rng,
		content.ContentLength,
	)
}

// serveFileRange serves the range of a file from a given source reader, without the
// need for implementation of io.Seeker. Instead we read the first 'start' many bytes
// into a discard reader. Code is adapted from https://codeberg.org/gruf/simplehttp.
func serveFileRange(rw http.ResponseWriter, r *http.Request, src io.Reader, rng string, size int64) {
	var i int

	if i = strings.IndexByte(rng, '='); i < 0 {
		// Range must include a separating '=' to indicate start
		http.Error(rw, "Bad Range Header", http.StatusBadRequest)
		return
	}

	if rng[:i] != "bytes" {
		// We only support byte ranges in our implementation
		http.Error(rw, "Unsupported Range Unit", http.StatusBadRequest)
		return
	}

	// Reslice past '='
	rng = rng[i+1:]

	if i = strings.IndexByte(rng, '-'); i < 0 {
		// Range header must contain a beginning and end separated by '-'
		http.Error(rw, "Bad Range Header", http.StatusBadRequest)
		return
	}

	var (
		err error

		// default start + end ranges
		start, end = int64(0), size - 1

		// start + end range strings
		startRng, endRng string
	)

	if startRng = rng[:i]; len(startRng) > 0 {
		// Parse the start of this byte range
		start, err = strconv.ParseInt(startRng, 10, 64)
		if err != nil {
			http.Error(rw, "Bad Range Header", http.StatusBadRequest)
			return
		}

		if start < 0 {
			// This range starts *before* the file start, why did they send this lol
			rw.Header().Set("Content-Range", "bytes *"+strconv.FormatInt(size, 10))
			http.Error(rw, "Unsatisfiable Range", http.StatusRequestedRangeNotSatisfiable)
			return
		}
	} else {
		// No start supplied, implying file start
		startRng = "0"
	}

	if endRng = rng[i+1:]; len(endRng) > 0 {
		// Parse the end of this byte range
		end, err = strconv.ParseInt(endRng, 10, 64)
		if err != nil {
			http.Error(rw, "Bad Range Header", http.StatusBadRequest)
			return
		}

		if end > size {
			// This range exceeds length of the file, therefore unsatisfiable
			rw.Header().Set("Content-Range", "bytes *"+strconv.FormatInt(size, 10))
			http.Error(rw, "Unsatisfiable Range", http.StatusRequestedRangeNotSatisfiable)
			return
		}
	} else {
		// No end supplied, implying file end
		endRng = strconv.FormatInt(end, 10)
	}

	if start >= end {
		// This range starts _after_ their range end, unsatisfiable and nonsense!
		rw.Header().Set("Content-Range", "bytes *"+strconv.FormatInt(size, 10))
		http.Error(rw, "Unsatisfiable Range", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// Dump the first 'start' many bytes into the void...
	if _, err := fastcopy.CopyN(io.Discard, src, start); err != nil {
		log.Errorf(r.Context(), "error reading from source: %v", err)
		return
	}

	// Determine new content length
	// after slicing to given range.
	length := end - start + 1

	if end < size-1 {
		// Range end < file end, limit the reader
		src = io.LimitReader(src, length)
	}

	// Write the necessary length and range headers
	rw.Header().Set("Content-Range", "bytes "+startRng+"-"+endRng+"/"+strconv.FormatInt(size, 10))
	rw.Header().Set("Content-Length", strconv.FormatInt(length, 10))
	rw.WriteHeader(http.StatusPartialContent)

	// Read the "seeked" source reader into destination writer.
	if _, err := fastcopy.Copy(rw, src); err != nil {
		log.Errorf(r.Context(), "error reading from source: %v", err)
		return
	}
}

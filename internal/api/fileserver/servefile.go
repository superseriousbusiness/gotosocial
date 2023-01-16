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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/iotools"
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
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(err), m.processor.InstanceGet)
		return
	}

	// We use request params to check what to pull out of the database/storage so check everything. A request URL should be formatted as follows:
	// "https://example.org/fileserver/[ACCOUNT_ID]/[MEDIA_TYPE]/[MEDIA_SIZE]/[FILE_NAME]"
	// "FILE_NAME" consists of two parts, the attachment's database id, a period, and the file extension.
	accountID := c.Param(AccountIDKey)
	if accountID == "" {
		err := fmt.Errorf("missing %s from request", AccountIDKey)
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(err), m.processor.InstanceGet)
		return
	}

	mediaType := c.Param(MediaTypeKey)
	if mediaType == "" {
		err := fmt.Errorf("missing %s from request", MediaTypeKey)
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(err), m.processor.InstanceGet)
		return
	}

	mediaSize := c.Param(MediaSizeKey)
	if mediaSize == "" {
		err := fmt.Errorf("missing %s from request", MediaSizeKey)
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(err), m.processor.InstanceGet)
		return
	}

	fileName := c.Param(FileNameKey)
	if fileName == "" {
		err := fmt.Errorf("missing %s from request", FileNameKey)
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(err), m.processor.InstanceGet)
		return
	}

	content, errWithCode := m.processor.FileGet(c.Request.Context(), authed, &apimodel.GetContentRequestForm{
		AccountID: accountID,
		MediaType: mediaType,
		MediaSize: mediaSize,
		FileName:  fileName,
	})
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	defer func() {
		// close content when we're done
		if content.Content != nil {
			if err := content.Content.Close(); err != nil {
				log.Errorf("ServeFile: error closing readcloser: %s", err)
			}
		}
	}()

	if content.URL != nil {
		c.Redirect(http.StatusFound, content.URL.String())
		return
	}

	// TODO: if the requester only accepts text/html we should try to serve them *something*.
	// This is mostly needed because when sharing a link to a gts-hosted file on something like mastodon, the masto servers will
	// attempt to look up the content to provide a preview of the link, and they ask for text/html.
	format, err := apiutil.NegotiateAccept(c, apiutil.MIME(content.ContentType))
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	// if this is a head request, just return info + throw the reader away
	if c.Request.Method == http.MethodHead {
		c.Header("Content-Type", format)
		c.Header("Content-Length", strconv.FormatInt(content.ContentLength, 10))
		c.Status(http.StatusOK)
		return
	}

	// create a "slurp" buffer ;)
	b := make([]byte, 64)

	// Try read the first 64 bytes into memory, to try return a more useful "not found" error.
	if _, err := io.ReadFull(content.Content, b); err != nil &&
		(err != io.ErrUnexpectedEOF && err != io.EOF) {
		err = fmt.Errorf("ServeFile: error reading from content: %w", err)
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(err, err.Error()), m.processor.InstanceGet)
		return
	}

	// reconstruct the original content reader
	r := io.MultiReader(bytes.NewReader(b), content.Content)

	// Check the Range header: if this is a simple query for the whole file, we can return it now.
	if c.GetHeader("Range") == "" && c.GetHeader("If-Range") == "" {
		c.DataFromReader(http.StatusOK, content.ContentLength, format, r, nil)
		return
	}

	// Range is set, so we need a ReadSeeker to pass to the ServeContent function.
	tfs, err := iotools.TempFileSeeker(r)
	if err != nil {
		err = fmt.Errorf("ServeFile: error creating temp file seeker: %w", err)
		apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGet)
		return
	}
	defer func() {
		if err := tfs.Close(); err != nil {
			log.Errorf("ServeFile: error closing temp file seeker: %s", err)
		}
	}()

	// to avoid ServeContent wasting time seeking for the
	// mime type, set this header already since we know it
	c.Header("Content-Type", format)

	// allow ServeContent to handle the rest of the request;
	// it will handle Range as appropriate, and write correct
	// response headers, http code, etc
	http.ServeContent(c.Writer, c.Request, fileName, content.ContentUpdated, tfs)
}

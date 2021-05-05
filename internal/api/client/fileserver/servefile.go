/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// ServeFile is for serving attachments, headers, and avatars to the requester from instance storage.
//
// Note: to mitigate scraping attempts, no information should be given out on a bad request except "404 page not found".
// Don't give away account ids or media ids or anything like that; callers shouldn't be able to infer anything.
func (m *FileServer) ServeFile(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
		"func":        "ServeFile",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})
	l.Trace("received request")

	authed, err := oauth.Authed(c, false, false, false, false)
	if err != nil {
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	// We use request params to check what to pull out of the database/storage so check everything. A request URL should be formatted as follows:
	// "https://example.org/fileserver/[ACCOUNT_ID]/[MEDIA_TYPE]/[MEDIA_SIZE]/[FILE_NAME]"
	// "FILE_NAME" consists of two parts, the attachment's database id, a period, and the file extension.
	accountID := c.Param(AccountIDKey)
	if accountID == "" {
		l.Debug("missing accountID from request")
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	mediaType := c.Param(MediaTypeKey)
	if mediaType == "" {
		l.Debug("missing mediaType from request")
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	mediaSize := c.Param(MediaSizeKey)
	if mediaSize == "" {
		l.Debug("missing mediaSize from request")
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	fileName := c.Param(FileNameKey)
	if fileName == "" {
		l.Debug("missing fileName from request")
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	content, err := m.processor.MediaGet(authed, &model.GetContentRequestForm{
		AccountID: accountID,
		MediaType: mediaType,
		MediaSize: mediaSize,
		FileName:  fileName,
	})
	if err != nil {
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	c.DataFromReader(http.StatusOK, content.ContentLength, content.ContentType, bytes.NewReader(content.Content), nil)
}

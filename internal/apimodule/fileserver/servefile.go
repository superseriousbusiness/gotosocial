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
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
)

// ServeFile is for serving attachments, headers, and avatars to the requester from instance storage.
//
// Note: to mitigate scraping attempts, no information should be given out on a bad request except "404 page not found".
// Don't give away account ids or media ids or anything like that; callers shouldn't be able to infer anything.
func (m *fileServer) ServeFile(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
		"func":        "ServeFile",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})
	l.Trace("received request")

	// We use request params to check what to pull out of the database/storage so check everything. A request URL should be formatted as follows:
	// "https://example.org/fileserver/[ACCOUNT_ID]/[MEDIA_TYPE]/[MEDIA_SIZE]/[FILE_NAME]"
	// "FILE_NAME" consists of two parts, the attachment's database id, a period, and the file extension.
	accountID := c.Param(accountIDKey)
	if accountID == "" {
		l.Debug("missing accountID from request")
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	mediaType := c.Param(mediaTypeKey)
	if mediaType == "" {
		l.Debug("missing mediaType from request")
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	mediaSize := c.Param(mediaSizeKey)
	if mediaSize == "" {
		l.Debug("missing mediaSize from request")
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	fileName := c.Param(fileNameKey)
	if fileName == "" {
		l.Debug("missing fileName from request")
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	// Only serve media types that are defined in our internal media module
	switch mediaType {
	case media.MediaHeader, media.MediaAvatar, media.MediaAttachment, media.MediaEmoji:
	default:
		l.Debugf("mediatype %s not recognized", mediaType)
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	// This corresponds to original-sized image as it was uploaded, small (which is the thumbnail), or static
	switch mediaSize {
	case media.MediaOriginal, media.MediaSmall, media.MediaStatic:
	default:
		l.Debugf("mediasize %s not recognized", mediaSize)
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	// derive the media id and the file extension from the last part of the request
	spl := strings.Split(fileName, ".")
	if len(spl) != 2 {
		l.Debugf("filename %s not parseable", fileName)
		c.String(http.StatusNotFound, "404 page not found")
		return
	}
	wantedMediaID := spl[0]
	fileExtension := spl[1]
	if wantedMediaID == "" || fileExtension == "" {
		l.Debugf("filename %s not parseable", fileName)
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	// now we know the attachment ID that the caller is asking for we can use it to pull the attachment out of the db
	attachment := &gtsmodel.MediaAttachment{}
	if err := m.db.GetByID(wantedMediaID, attachment); err != nil {
		l.Debugf("attachment with id %s not retrievable: %s", wantedMediaID, err)
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	// make sure the given account id owns the requested attachment
	if accountID != attachment.AccountID {
		l.Debugf("account %s does not own attachment with id %s", accountID, wantedMediaID)
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	// now we can start preparing the response depending on whether we're serving a thumbnail or a larger attachment
	var storagePath string
	var contentType string
	var contentLength int
	switch mediaSize {
	case media.MediaOriginal:
		storagePath = attachment.File.Path
		contentType = attachment.File.ContentType
		contentLength = attachment.File.FileSize
	case media.MediaSmall:
		storagePath = attachment.Thumbnail.Path
		contentType = attachment.Thumbnail.ContentType
		contentLength = attachment.Thumbnail.FileSize
	}

	// use the path listed on the attachment we pulled out of the database to retrieve the object from storage
	attachmentBytes, err := m.storage.RetrieveFileFrom(storagePath)
	if err != nil {
		l.Debugf("error retrieving from storage: %s", err)
		c.String(http.StatusNotFound, "404 page not found")
		return
	}

	// finally we can return with all the information we derived above
	c.DataFromReader(http.StatusOK, int64(contentLength), contentType, bytes.NewReader(attachmentBytes), map[string]string{})
}

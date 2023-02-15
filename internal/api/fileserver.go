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

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/fileserver"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

type Fileserver struct {
	fileserver *fileserver.Module
}

// maxAge returns an appropriate max-age value for the
// storage method that's being used.
//
// The default max-age is very long to reflect that we
// never host different files at the same URL (since
// ULIDs are generated per piece of media), so we can
// easily prevent clients having to fetch files repeatedly.
//
// If we're using non-proxying s3, however, the max age is
// significantly shorter, to ensure that clients don't
// cache redirect responses to expired pre-signed URLs.
func maxAge() string {
	if config.GetStorageBackend() == "s3" && !config.GetStorageS3Proxy() {
		return "max-age=86400" // 24h
	}

	return "max-age=604800" // 7d
}

func (f *Fileserver) Route(r router.Router, m ...gin.HandlerFunc) {
	fileserverGroup := r.AttachGroup("fileserver")

	// attach middlewares appropriate for this group
	fileserverGroup.Use(m...)
	fileserverGroup.Use(middleware.CacheControl("private", maxAge()))

	f.fileserver.Route(fileserverGroup.Handle)
}

func NewFileserver(p *processing.Processor) *Fileserver {
	return &Fileserver{
		fileserver: fileserver.New(p),
	}
}

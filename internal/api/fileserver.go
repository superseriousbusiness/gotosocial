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

func (f *Fileserver) Route(r router.Router, m ...gin.HandlerFunc) {
	fileserverGroup := r.AttachGroup("fileserver")

	// Attach middlewares appropriate for this group.
	fileserverGroup.Use(m...)
	// If we're using local storage or proxying s3, we can set a
	// long max-age on all file requests to reflect that we
	// never host different files at the same URL (since
	// ULIDs are generated per piece of media), so we can
	// easily prevent clients having to fetch files repeatedly.
	//
	// If we *are* using non-proxying s3, however, the max age
	// must be set dynamically within the request handler,
	// based on how long the signed URL has left to live before
	// it expires. This ensures that clients won't cache expired
	// links. This is done within fileserver/servefile.go.
	if config.GetStorageBackend() == "local" || config.GetStorageS3Proxy() {
		fileserverGroup.Use(middleware.CacheControl("private", "max-age=604800")) // 7d
	}

	f.fileserver.Route(fileserverGroup.Handle)
}

func NewFileserver(p processing.Processor) *Fileserver {
	return &Fileserver{
		fileserver: fileserver.New(p),
	}
}

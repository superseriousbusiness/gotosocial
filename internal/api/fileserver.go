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
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

type Fileserver struct {
	fileserver *fileserver.Module
}

func (f *Fileserver) Route(r router.Router, m ...gin.HandlerFunc) {
	fileserverGroup := r.AttachGroup("fileserver")

	// attach middlewares appropriate for this group
	fileserverGroup.Use(m...)
	fileserverGroup.Use(
		// Since we'll never host different files at the same
		// URL (bc the ULIDs are generated per piece of media),
		// it's sensible and safe to use a long cache here, so
		// that clients don't keep fetching files over + over again.
		//
		// Nevertheless, we should use 'private' to indicate
		// that there might be media in there which are gated by ACLs.
		middleware.CacheControl("private", "max-age=604800"),
	)

	f.fileserver.Route(fileserverGroup.Handle)
}

func NewFileserver(p processing.Processor) *Fileserver {
	return &Fileserver{
		fileserver: fileserver.New(p),
	}
}

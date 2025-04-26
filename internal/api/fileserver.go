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

package api

import (
	"code.superseriousbusiness.org/gotosocial/internal/api/fileserver"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/middleware"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/router"
	"github.com/gin-gonic/gin"
)

type Fileserver struct {
	fileserver *fileserver.Module
}

// Attach cache middleware appropriate for file serving.
func useFSCacheMiddleware(grp *gin.RouterGroup) {
	// If we're using local storage or proxying s3 (ie., serving
	// from here) we can set a long max-age + immutable on file
	// requests to reflect that we never host different files at
	// the same URL (since ULIDs are generated per piece of media),
	// so we can prevent clients having to fetch files repeatedly.
	//
	// If we *are* using non-proxying s3, however (ie., not serving
	// from here) the max age must be set dynamically within the
	// request handler, based on how long the signed URL has left
	// to live before it expires. This ensures that clients won't
	// cache expired links. This is done within fileserver/servefile.go
	// so we should not set the middleware here in that case.
	//
	// See:
	//
	// - https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching#avoiding_revalidation
	// - https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#immutable
	servingFromHere := config.GetStorageBackend() == "local" || config.GetStorageS3Proxy()
	if !servingFromHere {
		return
	}

	grp.Use(middleware.CacheControl(middleware.CacheControlConfig{
		Directives: []string{"private", "max-age=604800", "immutable"},
		Vary:       []string{"Range"}, // Cache partial ranges separately.
	}))
}

// Route the "main" fileserver group
// that handles everything except emojis.
func (f *Fileserver) Route(
	r *router.Router,
	m ...gin.HandlerFunc,
) {
	const fsGroupPath = "fileserver" +
		"/:" + fileserver.AccountIDKey +
		"/:" + fileserver.MediaTypeKey
	fsGroup := r.AttachGroup(fsGroupPath)

	// Attach provided +
	// cache middlewares.
	fsGroup.Use(m...)
	useFSCacheMiddleware(fsGroup)

	f.fileserver.Route(fsGroup.Handle)
}

// Route the "emojis" fileserver
// group to handle emojis specifically.
//
// instanceAccount ID is required because
// that is the ID under which all emoji
// files are stored, and from which all
// emoji file requests are therefore served.
func (f *Fileserver) RouteEmojis(
	r *router.Router,
	instanceAcctID string,
	m ...gin.HandlerFunc,
) {
	var fsEmojiGroupPath = "fileserver" +
		"/" + instanceAcctID +
		"/" + string(media.TypeEmoji)
	fsEmojiGroup := r.AttachGroup(fsEmojiGroupPath)

	// Inject the instance account and emoji media
	// type params into the gin context manually,
	// since we know we're only going to be serving
	// emojis (stored under the instance account ID)
	// from this group. This allows us to use the
	// same handler functions for both the "main"
	// fileserver handler and the emojis handler.
	fsEmojiGroup.Use(func(c *gin.Context) {
		c.Params = append(c.Params, []gin.Param{
			{
				Key:   fileserver.AccountIDKey,
				Value: instanceAcctID,
			},
			{
				Key:   fileserver.MediaTypeKey,
				Value: string(media.TypeEmoji),
			},
		}...)
	})

	// Attach provided +
	// cache middlewares.
	fsEmojiGroup.Use(m...)
	useFSCacheMiddleware(fsEmojiGroup)

	f.fileserver.Route(fsEmojiGroup.Handle)
}

func NewFileserver(p *processing.Processor) *Fileserver {
	return &Fileserver{
		fileserver: fileserver.New(p),
	}
}

/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"codeberg.org/gruf/go-cache/v2"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

const (
	confirmEmailPath = "/" + uris.ConfirmEmailPath
	profilePath      = "/@:" + usernameKey
	statusPath       = profilePath + "/statuses/:" + statusIDKey
	adminPanelPath   = "/admin"
	userPanelpath    = "/user"
	assetsPath       = "/assets"

	tokenParam  = "token"
	usernameKey = "username"
	statusIDKey = "status"
)

// Module implements the api.ClientModule interface for web pages.
type Module struct {
	processor            processing.Processor
	webAssetsAbsFilePath string
	assetsETagCache      cache.Cache[string, eTagCacheEntry]
	defaultAvatars       []string
}

// New returns a new api.ClientModule for web pages.
func New(processor processing.Processor) (api.ClientModule, error) {
	webAssetsBaseDir := config.GetWebAssetBaseDir()
	if webAssetsBaseDir == "" {
		return nil, fmt.Errorf("%s cannot be empty and must be a relative or absolute path", config.WebAssetBaseDirFlag())
	}

	webAssetsAbsFilePath, err := filepath.Abs(webAssetsBaseDir)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path of %s: %s", webAssetsBaseDir, err)
	}

	defaultAvatarsAbsFilePath := filepath.Join(webAssetsAbsFilePath, "default_avatars")
	defaultAvatarFiles, err := ioutil.ReadDir(defaultAvatarsAbsFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading default avatars at %s: %s", defaultAvatarsAbsFilePath, err)
	}

	defaultAvatars := []string{}
	for _, f := range defaultAvatarFiles {
		// ignore directories
		if f.IsDir() {
			continue
		}

		// ignore files bigger than 50kb
		if f.Size() > 50000 {
			continue
		}

		// get the name of the file, eg avatar.jpeg
		fileName := f.Name()

		// get just the .jpeg, for example, from avatar.jpeg
		extensionWithDot := filepath.Ext(fileName)

		// remove the leading . to just get, eg, jpeg
		extension := strings.TrimPrefix(extensionWithDot, ".")

		// take only files with simple extensions
		// that we know will work OK as avatars
		switch strings.ToLower(extension) {
		case "svg", "jpeg", "jpg", "gif", "png":
			avatar := fmt.Sprintf("%s/default_avatars/%s", assetsPath, f.Name())
			defaultAvatars = append(defaultAvatars, avatar)
		default:
			continue
		}
	}

	assetsETagCache := cache.New[string, eTagCacheEntry]()
	assetsETagCache.SetTTL(time.Hour, false)
	assetsETagCache.Start(time.Minute)

	return &Module{
		processor:            processor,
		webAssetsAbsFilePath: webAssetsAbsFilePath,
		assetsETagCache:      assetsETagCache,
		defaultAvatars:       defaultAvatars,
	}, nil
}

// Route satisfies the RESTAPIModule interface
func (m *Module) Route(s router.Router) error {
	// serve static files from assets dir at /assets
	assetsGroup := s.AttachGroup(assetsPath)
	m.mountAssetsFilesystem(assetsGroup)

	s.AttachHandler(http.MethodGet, adminPanelPath, m.AdminPanelHandler)
	// redirect /admin/ to /admin
	s.AttachHandler(http.MethodGet, adminPanelPath+"/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, adminPanelPath)
	})

	s.AttachHandler(http.MethodGet, userPanelpath, m.UserPanelHandler)
	// redirect /settings/ to /settings
	s.AttachHandler(http.MethodGet, userPanelpath+"/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, userPanelpath)
	})

	// serve front-page
	s.AttachHandler(http.MethodGet, "/", m.baseHandler)

	// serve profile pages at /@username
	s.AttachHandler(http.MethodGet, profilePath, m.profileGETHandler)

	// serve statuses
	s.AttachHandler(http.MethodGet, statusPath, m.threadGETHandler)

	// serve email confirmation page at /confirm_email?token=whatever
	s.AttachHandler(http.MethodGet, confirmEmailPath, m.confirmEmailGETHandler)

	// 404 handler
	s.AttachNoRouteHandler(func(c *gin.Context) {
		api.ErrorHandler(c, gtserror.NewErrorNotFound(errors.New(http.StatusText(http.StatusNotFound))), m.processor.InstanceGet)
	})

	return nil
}

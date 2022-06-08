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

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

const (
	confirmEmailPath = "/" + uris.ConfirmEmailPath
	tokenParam       = "token"
	usernameKey      = "username"
	statusIDKey      = "status"
	profilePath      = "/@:" + usernameKey
	statusPath       = profilePath + "/statuses/:" + statusIDKey
)

// Module implements the api.ClientModule interface for web pages.
type Module struct {
	processor      processing.Processor
	assetsPath     string
	adminPath      string
	defaultAvatars []string
}

// New returns a new api.ClientModule for web pages.
func New(processor processing.Processor) (api.ClientModule, error) {
	assetsBaseDir := config.GetWebAssetBaseDir()
	if assetsBaseDir == "" {
		return nil, fmt.Errorf("%s cannot be empty and must be a relative or absolute path", config.WebAssetBaseDirFlag())
	}

	assetsPath, err := filepath.Abs(assetsBaseDir)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path of %s: %s", assetsBaseDir, err)
	}

	defaultAvatarsPath := filepath.Join(assetsPath, "default_avatars")
	defaultAvatarFiles, err := ioutil.ReadDir(defaultAvatarsPath)
	if err != nil {
		return nil, fmt.Errorf("error reading default avatars at %s: %s", defaultAvatarsPath, err)
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

		extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(f.Name())), ".")

		// take only files with simple extensions
		switch extension {
		case "svg", "jpeg", "jpg", "gif", "png":
			defaultAvatarPath := fmt.Sprintf("/assets/default_avatars/%s", f.Name())
			defaultAvatars = append(defaultAvatars, defaultAvatarPath)
		default:
			continue
		}
	}

	return &Module{
		processor:      processor,
		assetsPath:     assetsPath,
		adminPath:      filepath.Join(assetsPath, "admin"),
		defaultAvatars: defaultAvatars,
	}, nil
}

func (m *Module) baseHandler(c *gin.Context) {
	l := logrus.WithField("func", "BaseGETHandler")
	l.Trace("serving index html")

	host := config.GetHost()
	instance, err := m.processor.InstanceGet(c.Request.Context(), host)
	if err != nil {
		l.Debugf("error getting instance from processor: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"instance": instance,
	})
}

// TODO: abstract the {admin, user}panel handlers in some way
func (m *Module) AdminPanelHandler(c *gin.Context) {
	l := logrus.WithField("func", "admin-panel")
	l.Trace("serving admin panel")

	host := config.GetHost()
	instance, err := m.processor.InstanceGet(c.Request.Context(), host)
	if err != nil {
		l.Debugf("error getting instance from processor: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.HTML(http.StatusOK, "frontend.tmpl", gin.H{
		"instance": instance,
		"stylesheets": []string{
			"/assets/Fork-Awesome/css/fork-awesome.min.css",
			"/assets/dist/panels-admin-style.css",
		},
		"javascript": []string{
			"/assets/dist/bundle.js",
			"/assets/dist/admin-panel.js",
		},
	})
}

func (m *Module) UserPanelHandler(c *gin.Context) {
	l := logrus.WithField("func", "user-panel")
	l.Trace("serving user panel")

	host := config.GetHost()
	instance, err := m.processor.InstanceGet(c.Request.Context(), host)
	if err != nil {
		l.Debugf("error getting instance from processor: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.HTML(http.StatusOK, "frontend.tmpl", gin.H{
		"instance": instance,
		"stylesheets": []string{
			"/assets/Fork-Awesome/css/fork-awesome.min.css",
			"/assets/dist/_colors.css",
			"/assets/dist/base.css",
			"/assets/dist/panels-user-style.css",
		},
		"javascript": []string{
			"/assets/dist/bundle.js",
			"/assets/dist/user-panel.js",
		},
	})
}

// Route satisfies the RESTAPIModule interface
func (m *Module) Route(s router.Router) error {
	// serve static files from assets dir at /assets
	s.AttachStaticFS("/assets", fileSystem{http.Dir(m.assetsPath)})

	s.AttachHandler(http.MethodGet, "/admin", m.AdminPanelHandler)
	// redirect /admin/ to /admin
	s.AttachHandler(http.MethodGet, "/admin/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin")
	})

	s.AttachHandler(http.MethodGet, "/user", m.UserPanelHandler)
	// redirect /settings/ to /settings
	s.AttachHandler(http.MethodGet, "/user/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/user")
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

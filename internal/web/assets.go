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
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type fileSystem struct {
	fs http.FileSystem
}

// FileSystem server that only accepts directory listings when an index.html is available
// from https://gist.github.com/hauxe/f2ea1901216177ccf9550a1b8bd59178
func (fs fileSystem) Open(path string) (http.File, error) {
	f, err := fs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, _ := f.Stat()
	if s.IsDir() {
		index := strings.TrimSuffix(path, "/") + "/index.html"
		if _, err := fs.fs.Open(index); err != nil {
			return nil, err
		}
	}

	return f, nil
}

func (m *Module) mountAssetsFilesystem(group *gin.RouterGroup) {
	webAssetsAbsFilePath, err := filepath.Abs(config.GetWebAssetBaseDir())
	if err != nil {
		log.Panicf("mountAssetsFilesystem: error getting absolute path of assets dir: %s", err)
	}

	fs := fileSystem{http.Dir(webAssetsAbsFilePath)}

	// use the cache middleware on all handlers in this group
	group.Use(m.cacheControlMiddleware(fs))

	// serve static file system in the root of this group,
	// will end up being something like "/assets/"
	group.StaticFS("/", fs)
}

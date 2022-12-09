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

package testrig

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

// NewTestRouter returns a Router suitable for testing
//
// If the environment variable GTS_WEB_TEMPLATE_BASE_DIR set, it will take that
// value as the template base directory instead.
func NewTestRouter(db db.DB) router.Router {
	if alternativeTemplateBaseDir := os.Getenv("GTS_WEB_TEMPLATE_BASE_DIR"); alternativeTemplateBaseDir != "" {
		config.Config(func(cfg *config.Configuration) {
			cfg.WebTemplateBaseDir = alternativeTemplateBaseDir
		})
	}

	if alternativeBindAddress := os.Getenv("GTS_BIND_ADDRESS"); alternativeBindAddress != "" {
		config.SetBindAddress(alternativeBindAddress)
	}

	if alternativePortStr := os.Getenv("GTS_PORT"); alternativePortStr != "" {
		if alternativePort, err := strconv.Atoi(alternativePortStr); err == nil {
			config.SetPort(alternativePort)
		}
	}

	r, err := router.New(context.Background())
	if err != nil {
		panic(err)
	}
	return r
}

// ConfigureTemplatesWithGin will panic on any errors related to template loading during tests
func ConfigureTemplatesWithGin(engine *gin.Engine, templatePath string) {
	router.LoadTemplateFunctions(engine)
	engine.LoadHTMLGlob(filepath.Join(templatePath, "*"))
}

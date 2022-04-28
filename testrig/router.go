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
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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
		viper.Set(config.Keys.WebTemplateBaseDir, alternativeTemplateBaseDir)
	}

	r, err := router.New(context.Background(), db)
	if err != nil {
		panic(err)
	}
	return r
}

// ConfigureTemplatesWithGin will panic on any errors related to template loading during tests
func ConfigureTemplatesWithGin(engine *gin.Engine) {
	router.LoadTemplateFunctions(engine)

	templateBaseDir := viper.GetString(config.Keys.WebTemplateBaseDir)

	if !filepath.IsAbs(templateBaseDir) {
		// https://stackoverflow.com/questions/31873396/is-it-possible-to-get-the-current-root-of-package-structure-as-a-string-in-golan
		_, runtimeCallerLocation, _, _ := runtime.Caller(0)
		projectRoot, err := filepath.Abs(filepath.Join(filepath.Dir(runtimeCallerLocation), "../"))
		if err != nil {
			panic(err)
		}

		templateBaseDir = filepath.Join(projectRoot, templateBaseDir)
	}

	if _, err := os.Stat(filepath.Join(templateBaseDir, "index.tmpl")); err != nil {
		panic(fmt.Errorf("%s doesn't seem to contain the templates; index.tmpl is missing: %w", templateBaseDir, err))
	}

	engine.LoadHTMLGlob(filepath.Join(templateBaseDir, "*"))
}

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

package router

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/template"
)

// LoadTemplates loads html templates for use by the given engine
func LoadTemplates(engine *gin.Engine) error {
	templateBaseDir := config.GetWebTemplateBaseDir()
	if templateBaseDir == "" {
		return fmt.Errorf("%s cannot be empty and must be a relative or absolute path", config.WebTemplateBaseDirFlag())
	}

	templateBaseDir, err := filepath.Abs(templateBaseDir)
	if err != nil {
		return fmt.Errorf("error getting absolute path of %s: %s", templateBaseDir, err)
	}

	if _, err := os.Stat(filepath.Join(templateBaseDir, "index.tmpl")); err != nil {
		return fmt.Errorf("%s doesn't seem to contain the templates; index.tmpl is missing: %w", templateBaseDir, err)
	}

	engine.LoadHTMLGlob(filepath.Join(templateBaseDir, "*"))
	return nil
}

func LoadTemplateFunctions(engine *gin.Engine) {
	engine.SetFuncMap(template.FuncMap)
}

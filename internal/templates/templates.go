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

package templates

import (
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/web"
)

// ParseTemplates reads the templates from disk or embedded in the executable.
// It throws errors if an invalid path to the templates is given, or the
// templates can't be parsed.
// If templateBaseDir is empty its value is read from config instead.
func ParseTemplates(matchOn, templateBaseDir string) (*template.Template, error) {
	if templateBaseDir == "" {
		templateBaseDir = config.GetWebTemplateBaseDir()
		if templateBaseDir == "" {
			return nil, fmt.Errorf("%s cannot be empty and must be a relative or absolute path", config.WebTemplateBaseDirFlag())
		}
	}

	var err error
	templateBaseDir, err = filepath.Abs(templateBaseDir)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path of %s: %s", templateBaseDir, err)
	}

	templateSource := web.NewHybridFS(web.EmbeddedTemplates, templateBaseDir)
	return template.New("").
		Funcs(EngineTemplateFuncs).
		ParseFS(templateSource, matchOn)
}

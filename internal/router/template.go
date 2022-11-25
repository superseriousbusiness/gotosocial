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

package router

import (
	"html/template"

	"github.com/gin-gonic/gin"

	"github.com/superseriousbusiness/gotosocial/internal/templates"
)

// LoadTemplates loads html templates for use by the given engine.
// If `templateBaseDir` is empty its value is read from config instead.
func LoadTemplates(engine *gin.Engine, templateBaseDir string) (*template.Template, error) {
	templates, err := templates.ParseTemplates("*.tmpl", templateBaseDir)
	if err == nil {
		engine.SetHTMLTemplate(templates)
	}
	return templates, err
}

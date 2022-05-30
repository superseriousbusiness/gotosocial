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
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// LoadTemplates loads html templates for use by the given engine
func loadTemplates(engine *gin.Engine) error {
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

func oddOrEven(n int) string {
	if n%2 == 0 {
		return "even"
	}
	return "odd"
}

func noescape(str string) template.HTML {
	/* #nosec G203 */
	return template.HTML(str)
}

func timestamp(stamp string) string {
	t, _ := time.Parse(time.RFC3339, stamp)
	return t.Format("January 2, 2006, 15:04:05")
}

func timestampShort(stamp string) string {
	t, _ := time.Parse(time.RFC3339, stamp)
	return t.Format("January, 2006")
}

type iconWithLabel struct {
	faIcon string
	label  string
}

func visibilityIcon(visibility model.Visibility) template.HTML {
	var icon iconWithLabel

	switch visibility {
	case model.VisibilityPublic:
		icon = iconWithLabel{"globe", "public"}
	case model.VisibilityUnlisted:
		icon = iconWithLabel{"unlock", "unlisted"}
	case model.VisibilityPrivate:
		icon = iconWithLabel{"lock", "private"}
	case model.VisibilityMutualsOnly:
		icon = iconWithLabel{"handshake-o", "mutuals only"}
	case model.VisibilityDirect:
		icon = iconWithLabel{"envelope", "direct"}
	}

	/* #nosec G203 */
	return template.HTML(fmt.Sprintf(`<i aria-label="Visibility: %v" class="fa fa-%v"></i>`, icon.label, icon.faIcon))
}

func LoadTemplateFunctions(engine *gin.Engine) {
	engine.SetFuncMap(template.FuncMap{
		"noescape":       noescape,
		"oddOrEven":      oddOrEven,
		"visibilityIcon": visibilityIcon,
		"timestamp":      timestamp,
		"timestampShort": timestampShort,
	})
}

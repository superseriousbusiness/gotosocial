/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

// loadTemplates loads html templates for use by the given engine
func loadTemplates(cfg *config.Config, engine *gin.Engine) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %s", err)
	}

	tmPath := filepath.Join(cwd, fmt.Sprintf("%s*", cfg.TemplateConfig.BaseDir))

	engine.LoadHTMLGlob(tmPath)
	return nil
}

func oddOrEven(n int) string {
	if n%2 == 0 {
		return "even"
	}
	return "odd"
}

func noescape(str string) template.HTML {
	return template.HTML(str)
}

func timestamp(stamp string) string {
	t, _ := time.Parse(time.RFC3339, stamp)
	return t.Format("January 2, 2006, 15:04:05")
}

type iconWithLabel struct {
	faIcon string
	label  string
}

func visibilityIcon(visibility model.Visibility) template.HTML {
	var icon iconWithLabel

	if visibility == model.VisibilityPublic {
		icon = iconWithLabel{"globe", "public"}
	} else if visibility == model.VisibilityUnlisted {
		icon = iconWithLabel{"unlock", "unlisted"}
	} else if visibility == model.VisibilityPrivate {
		icon = iconWithLabel{"lock", "private"}
	} else if visibility == model.VisibilityMutualsOnly {
		icon = iconWithLabel{"handshake-o", "mutuals only"}
	} else if visibility == model.VisibilityDirect {
		icon = iconWithLabel{"envelope", "direct"}
	}

	return template.HTML(fmt.Sprintf(`<i aria-label="Visiblity: %v" class="fa fa-%v"></i>`, icon.label, icon.faIcon))
}

func loadTemplateFunctions(engine *gin.Engine) {
	engine.SetFuncMap(template.FuncMap{
		"noescape":       noescape,
		"oddOrEven":      oddOrEven,
		"visibilityIcon": visibilityIcon,
		"timestamp":      timestamp,
	})
}

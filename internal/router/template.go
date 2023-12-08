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
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func LoadTemplates(engine *gin.Engine) error {
	templateBaseDir := config.GetWebTemplateBaseDir()
	if templateBaseDir == "" {
		return gtserror.Newf(
			"%s cannot be empty and must be a relative or absolute path",
			config.WebTemplateBaseDirFlag(),
		)
	}

	templateDirAbs, err := filepath.Abs(templateBaseDir)
	if err != nil {
		return gtserror.Newf(
			"error getting absolute path of web-template-base-dir %s: %w",
			templateBaseDir, err,
		)
	}

	indexTmplPath := filepath.Join(templateDirAbs, "index.tmpl")
	if _, err := os.Stat(indexTmplPath); err != nil {
		return gtserror.Newf(
			"cannot find index.tmpl in web template directory %s: %w",
			templateDirAbs, err,
		)
	}

	// Bring base template into scope.
	tmpl := template.New("base")

	// Set "include" function to render provided
	// template name using the base template.
	FuncMap["include"] = func(name string, data any) (string, error) {
		var buf strings.Builder
		err := tmpl.ExecuteTemplate(&buf, name, data)
		return buf.String(), err
	}

	// Load functions into the base template, and
	// associate other templates with base template.
	templateGlob := filepath.Join(templateDirAbs, "*")
	tmpl, err = tmpl.Funcs(FuncMap).ParseGlob(templateGlob)
	if err != nil {
		return gtserror.Newf("error loading templates: %w", err)
	}

	// Almost done; teach the
	// engine how to render.
	engine.SetFuncMap(FuncMap)
	engine.HTMLRender = render.HTMLProduction{Template: tmpl}

	return nil
}

const (
	justTime     = "15:04"
	dateYear     = "Jan 02, 2006"
	dateTime     = "Jan 02, 15:04"
	dateYearTime = "Jan 02, 2006, 15:04"
	monthYear    = "Jan, 2006"
	badTimestamp = "bad timestamp"
)

var FuncMap = template.FuncMap{
	"escape":           escape,
	"noescape":         noescape,
	"noescapeAttr":     noescapeAttr,
	"oddOrEven":        oddOrEven,
	"visibilityIcon":   visibilityIcon,
	"timestamp":        timestamp,
	"timestampVague":   timestampVague,
	"timestampPrecise": timestampPrecise,
	"emojify":          emojify,
	"acctInstance":     acctInstance,
	"increment":        increment,
	"indent":           indent,
}

func oddOrEven(n int) string {
	if n%2 == 0 {
		return "even"
	}
	return "odd"
}

func escape(str string) template.HTML {
	/* #nosec G203 */
	return template.HTML(template.HTMLEscapeString(str))
}

func noescape(str string) template.HTML {
	/* #nosec G203 */
	return template.HTML(str)
}

func noescapeAttr(str string) template.HTMLAttr {
	/* #nosec G203 */
	return template.HTMLAttr(str)
}

func timestamp(stamp string) string {
	t, err := util.ParseISO8601(stamp)
	if err != nil {
		log.Errorf(nil, "error parsing timestamp %s: %s", stamp, err)
		return badTimestamp
	}

	t = t.Local()

	tYear, tMonth, tDay := t.Date()
	now := time.Now()
	currentYear, currentMonth, currentDay := now.Date()

	switch {
	case tYear == currentYear && tMonth == currentMonth && tDay == currentDay:
		return "Today, " + t.Format(justTime)
	case tYear == currentYear:
		return t.Format(dateTime)
	default:
		return t.Format(dateYear)
	}
}

func timestampPrecise(stamp string) string {
	t, err := util.ParseISO8601(stamp)
	if err != nil {
		log.Errorf(nil, "error parsing timestamp %s: %s", stamp, err)
		return badTimestamp
	}
	return t.Local().Format(dateYearTime)
}

func timestampVague(stamp string) string {
	t, err := util.ParseISO8601(stamp)
	if err != nil {
		log.Errorf(nil, "error parsing timestamp %s: %s", stamp, err)
		return badTimestamp
	}
	return t.Format(monthYear)
}

func visibilityIcon(visibility apimodel.Visibility) template.HTML {
	var (
		label string
		icon  string
	)

	switch visibility {
	case apimodel.VisibilityPublic:
		label = "public"
		icon = "globe"
	case apimodel.VisibilityUnlisted:
		label = "unlisted"
		icon = "unlock"
	case apimodel.VisibilityPrivate:
		label = "private"
		icon = "lock"
	case apimodel.VisibilityMutualsOnly:
		label = "mutuals-only"
		icon = "handshake-o"
	case apimodel.VisibilityDirect:
		label = "direct"
		icon = "envelope"
	}

	/* #nosec G203 */
	return template.HTML(fmt.Sprintf(
		`<i aria-label="Visibility: %s" class="fa fa-%s"></i>`,
		label, icon,
	))
}

// text is a template.HTML to affirm that the
// input of this function is already escaped.
func emojify(
	emojis []apimodel.Emoji,
	input template.HTML,
) template.HTML {
	out := text.Emojify(emojis, string(input))

	/* #nosec G203 */
	// (this is escaped above)
	return template.HTML(out)
}

func acctInstance(acct string) string {
	parts := strings.Split(acct, "@")
	if len(parts) > 1 {
		return "@" + parts[1]
	}

	return ""
}

func increment(i int) int {
	return i + 1
}

var (
	indentRegex  = regexp.MustCompile(`(?m)^`)
	indentStr    = "    "
	indentStrLen = len(indentStr)
	indents      = strings.Repeat(indentStr, 12)
)

func indent(n int, input string) string {
	return indentRegex.ReplaceAllString(input, indents[:n*indentStrLen])
}

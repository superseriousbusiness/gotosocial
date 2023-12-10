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
	"bytes"
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
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// LoadTemplates loads templates found at `web-template-base-dir`
// into the Gin engine, or errors if templates cannot be loaded.
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
	funcMap["include"] = func(name string, data any) (string, error) {
		var buf strings.Builder
		err := tmpl.ExecuteTemplate(&buf, name, data)
		return buf.String(), err
	}

	// Load functions into the base template, and
	// associate other templates with base template.
	templateGlob := filepath.Join(templateDirAbs, "*")
	tmpl, err = tmpl.Funcs(funcMap).ParseGlob(templateGlob)
	if err != nil {
		return gtserror.Newf("error loading templates: %w", err)
	}

	// Almost done; teach the
	// engine how to render.
	engine.SetFuncMap(funcMap)
	engine.HTMLRender = render.HTMLProduction{Template: tmpl}

	return nil
}

var funcMap = template.FuncMap{
	"acctInstance":     acctInstance,
	"demojify":         demojify,
	"emojify":          emojify,
	"escape":           escape,
	"increment":        increment,
	"indent":           indent,
	"outdentPre":       outdentPre,
	"noescapeAttr":     noescapeAttr,
	"noescape":         noescape,
	"oddOrEven":        oddOrEven,
	"timestampPrecise": timestampPrecise,
	"timestamp":        timestamp,
	"timestampVague":   timestampVague,
	"visibilityIcon":   visibilityIcon,
}

func oddOrEven(n int) string {
	if n%2 == 0 {
		return "even"
	}
	return "odd"
}

// escape HTML escapes the given string,
// returning a trusted template.
func escape(str string) template.HTML {
	/* #nosec G203 */
	return template.HTML(template.HTMLEscapeString(str))
}

// noescape marks the given string as a
// trusted template. The provided string
// MUST have already passed through a
// template or escaping function.
func noescape(str string) template.HTML {
	/* #nosec G203 */
	return template.HTML(str)
}

// noescapeAttr marks the given string as a
// trusted HTML attribute. The provided string
// MUST have already passed through a template
// or escaping function.
func noescapeAttr(str string) template.HTMLAttr {
	/* #nosec G203 */
	return template.HTMLAttr(str)
}

const (
	justTime     = "15:04"
	dateYear     = "Jan 02, 2006"
	dateTime     = "Jan 02, 15:04"
	dateYearTime = "Jan 02, 2006, 15:04"
	monthYear    = "Jan, 2006"
	badTimestamp = "bad timestamp"
)

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

// emojify replaces emojis in the given
// html fragment with suitable <img> tags.
//
// The provided input must have been
// escaped / templated already!
func emojify(
	emojis []apimodel.Emoji,
	html template.HTML,
) template.HTML {
	return text.EmojifyWeb(emojis, html)
}

// demojify replaces emoji shortcodes in
// the given fragment with empty strings.
//
// Output must then be escaped as appropriate.
func demojify(input string) string {
	return text.Demojify(input)
}

func acctInstance(acct string) string {
	parts := strings.Split(acct, "@")
	if len(parts) > 1 {
		return "@" + parts[1]
	}

	return ""
}

// increment adds 1
// to the given int.
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

var (
	pre = regexp.MustCompile(fmt.Sprintf(
		`(?mU)(?sm)^^((?:%s)+)<pre>.*</pre>`, indentStr),
	)
)

func outdentPre(html template.HTML) template.HTML {
	input := string(html)
	output := regexes.ReplaceAllStringFunc(pre, input,
		func(match string, buf *bytes.Buffer) string {
			// Reuse the regex to pull out submatches.
			matches := pre.FindAllStringSubmatch(match, -1)
			if len(matches) != 1 {
				return match
			}

			var (
				indented = matches[0][0]
				indent   = matches[0][1]
			)

			// Outdent everything in the inner match, add
			// a newline at the end to make it a bit neater.
			outdented := strings.ReplaceAll(indented, indent, "")

			// Replace original match with the outdented version.
			return strings.ReplaceAll(match, indented, outdented)
		},
	)
	return noescape(output)
}

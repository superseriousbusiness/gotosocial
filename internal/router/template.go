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
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
	"unsafe"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/regexes"
	"code.superseriousbusiness.org/gotosocial/internal/text"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

// LoadTemplates loads templates found at `web-template-base-dir`
// into the Gin engine, or errors if templates cannot be loaded.
//
// The special functions "include" and "includeAttr" will be added
// to the template funcMap for use in any template. Use these "include"
// functions when you need to pass a template through a pipeline.
// Otherwise, prefer the built-in "template" function.
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

	// Set additional "include" functions to render
	// provided template name using the base template.

	// Include renders the given template with the given data.
	// Unlike `template`, `include` can be chained with `indent`
	// to produce nicely-indented HTML.
	funcMap["include"] = func(name string, data any) (template.HTML, error) {
		var buf strings.Builder
		err := tmpl.ExecuteTemplate(&buf, name, data)

		// Template was already escaped by
		// ExecuteTemplate so we can trust it.
		return noescape(buf.String()), err
	}

	// includeIndex is like `include` but an index can be specified at
	// `.Index` and data will be nested at `.Item`. Useful when ranging.
	funcMap["includeIndex"] = func(name string, data any, index int) (template.HTML, error) {
		var buf strings.Builder
		withIndex := struct {
			Item  any
			Index int
		}{
			Item:  data,
			Index: index,
		}
		err := tmpl.ExecuteTemplate(&buf, name, withIndex)

		// Template was already escaped by
		// ExecuteTemplate so we can trust it.
		return noescape(buf.String()), err
	}

	// includeAttr is like `include` but for element attributes.
	funcMap["includeAttr"] = func(name string, data any) (template.HTMLAttr, error) {
		var buf strings.Builder
		err := tmpl.ExecuteTemplate(&buf, name, data)

		// Template was already escaped by
		// ExecuteTemplate so we can trust it.
		return noescapeAttr(buf.String()), err
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
	"add":                 add,
	"acctInstance":        acctInstance,
	"objectPosition":      objectPosition,
	"demojify":            demojify,
	"deref":               deref,
	"emojify":             emojify,
	"escape":              escape,
	"increment":           increment,
	"indent":              indent,
	"indentAttr":          indentAttr,
	"isNil":               isNil,
	"outdentPreformatted": outdentPreformatted,
	"noescapeAttr":        noescapeAttr,
	"noescape":            noescape,
	"oddOrEven":           oddOrEven,
	"subtract":            subtract,
	"timestampPrecise":    timestampPrecise,
	"timestampVague":      timestampVague,
	"visibilityIcon":      visibilityIcon,
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

// add adds n2 to n1.
func add(n1 int, n2 int) int {
	return n1 + n2
}

// subtract subtracts n2 from n1.
func subtract(n1 int, n2 int) int {
	return n1 - n2
}

var (
	// Find starts of lines to replace with indent.
	indentRegex = regexp.MustCompile(`(?m)^`)

	// One indent level.
	indentStr    = "    "
	indentStrLen = len(indentStr)

	// Preformatted slice of indents.
	indents = strings.Repeat(indentStr, 12)

	// Measure indent at the start of a line.
	indentDepthStr = fmt.Sprintf(`^((?:%s)+)`, indentStr)
	indentDepth    = regexp.MustCompile(`(?m)` + indentDepthStr)

	// Find <pre> tags and determine how indented they are.
	indentPre = regexp.MustCompile(fmt.Sprintf(`(?Ums)%s<pre>.*</pre>`, indentDepthStr))
	// Find content of alt or title attributes.
	indentAltOrTitle = regexp.MustCompile(`(?Ums)\b(?:alt|title)="(.*)"(?:\b|>|$)`)

	// Map of lazily-compiled replaceIndent
	// regexes, keyed by the indent they
	// replace, to avoid recompilation.
	//
	// At *most* 12 entries long.
	replaceIndents = sync.Map{}
)

// indent appropriately indents the given html
// by prepending each line with the indentStr.
func indent(n int, html template.HTML) template.HTML {
	out := indentRegex.ReplaceAllString(
		string(html),
		indents[:n*indentStrLen],
	)
	return noescape(out)
}

// indentAttr appropriately indents the given html
// attribute by prepending each line with the indentStr.
func indentAttr(n int, html template.HTMLAttr) template.HTMLAttr {
	out := indentRegex.ReplaceAllString(
		string(html),
		indents[:n*indentStrLen],
	)
	return noescapeAttr(out)
}

// outdentPreformatted outdents all preformatted text in
// the given HTML, ie., in `alt` and `title` attributes,
// and between `<pre>` tags, so that it renders correctly,
// even if it was indented before.
func outdentPreformatted(html template.HTML) template.HTML {
	input := string(html)
	output := regexes.ReplaceAllStringFunc(indentPre, input,
		func(match string, buf *bytes.Buffer) string {
			// Reuse the regex to pull out submatches.
			matches := indentPre.FindAllStringSubmatch(match, -1)

			// Ensure matches
			// expected length.
			if len(matches) != 1 {
				return match
			}

			// Ensure inner matches
			// expected length.
			innerMatches := matches[0]
			if len(innerMatches) != 2 {
				return match
			}

			var (
				indentedContent = innerMatches[0]
				indent          = innerMatches[1]
			)

			// Outdent everything in the inner match.
			outdented := strings.ReplaceAll(indentedContent, indent, "")

			// Replace original match with the outdented version.
			return strings.ReplaceAll(match, indentedContent, outdented)
		},
	)

	output = regexes.ReplaceAllStringFunc(indentAltOrTitle, output,
		func(match string, buf *bytes.Buffer) string {
			// Reuse the regex to pull out submatches.
			matches := indentAltOrTitle.FindAllStringSubmatch(match, -1)

			// Ensure matches
			// expected length.
			if len(matches) != 1 {
				return match
			}

			// Ensure inner matches
			// expected length.
			innerMatches := matches[0]
			if len(innerMatches) != 2 {
				return match
			}

			// The content of the alt or title
			// attr inside quotation marks.
			indentedContent := innerMatches[1]

			// Find all indents in this text.
			indents := indentDepth.FindAllString(indentedContent, -1)
			if len(indents) == 0 {
				// No indents in this text,
				// it's probably just something
				// inline like `alt="whatever"`.
				return match
			}

			// Find the shortest indent as this
			// is undoubtedly the one we added.
			//
			// By targeting the shortest one we
			// avoid removing user-inserted
			// whitespace at the start of lines
			// of alt text (eg., in poetry etc).
			slices.Sort(indents)
			indent := indents[0]

			// Load or create + store the
			// regex to replace this indent,
			// avoiding recompilation.
			var replaceIndent *regexp.Regexp
			if replaceIndentI, ok := replaceIndents.Load(indent); ok {
				// Got regex for this indent.
				replaceIndent = replaceIndentI.(*regexp.Regexp)
			} else {
				// No regex stored for
				// this indent yet, store it.
				replaceIndent = regexp.MustCompile(`(?m)^` + indent)
				replaceIndents.Store(indent, replaceIndent)
			}

			// Remove all occurrences of the indent
			// at the start of a line in the match.
			return replaceIndent.ReplaceAllString(match, "")
		},
	)

	return noescape(output)
}

// isNil will safely check if 'v' is nil without
// dealing with weird Go interface nil bullshit.
func isNil(i interface{}) bool {
	type eface struct{ _, data unsafe.Pointer }
	return (*eface)(unsafe.Pointer(&i)).data == nil
}

// deref returns the dereferenced value of
// its input. To ensure you don't pass nil
// pointers into this func, use isNil first.
func deref(i any) any {
	vOf := reflect.ValueOf(i)
	if vOf.Kind() != reflect.Pointer {
		// Not a pointer.
		return i
	}

	return vOf.Elem()
}

// objectPosition formats the given focus coordinates to a
// string suitable for use as a css object-position value.
func objectPosition(focusX float32, focusY float32) string {
	const fmts = "%.2f"
	xPos := ((focusX / 2) + .5) * 100
	yPos := ((focusY / -2) + .5) * 100
	return fmt.Sprintf(fmts, xPos) + "%" + " " + fmt.Sprintf(fmts, yPos) + "%"
}

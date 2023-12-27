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

package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

// WebPage encapsulates variables for
// rendering an HTML template within
// a standard GtS "page" template.
type WebPage struct {
	// Name of the template for rendering
	// the page. Eg., "example.tmpl".
	Template string

	// Instance model for rendering header,
	// footer, and "about" information.
	Instance *apimodel.InstanceV1

	// OGMeta for rendering page
	// "meta:og*" tags. Can be nil.
	OGMeta *OGMeta

	// Paths to CSS files to add to
	// the page as "stylesheet" entries.
	// Can be nil.
	Stylesheets []string

	// Paths to JS files to add to
	// the page as "script" entries.
	// Can be nil.
	Javascript []string

	// Extra parameters to pass to
	// the template for rendering,
	// eg., "account": *Account etc.
	// Can be nil.
	Extra map[string]any
}

// TemplateWebPage renders the given HTML template and
// page params within the standard GtS "page" template.
//
// ogMeta, stylesheets, javascript, and any extra
// properties will be provided to the template if
// set, but can all be nil.
func TemplateWebPage(
	c *gin.Context,
	page WebPage,
) {
	obj := map[string]any{
		"instance":    page.Instance,
		"ogMeta":      page.OGMeta,
		"stylesheets": page.Stylesheets,
		"javascript":  page.Javascript,
	}

	for k, v := range page.Extra {
		obj[k] = v
	}

	templatePage(c, page.Template, http.StatusOK, obj)
}

// templateErrorPage renders the given
// HTTP code, error, and request ID
// within the standard error template.
func templateErrorPage(
	c *gin.Context,
	instance *apimodel.InstanceV1,
	code int,
	err string,
	requestID string,
) {
	const errorTmpl = "error.tmpl"

	obj := map[string]any{
		"instance":  instance,
		"code":      code,
		"error":     err,
		"requestID": requestID,
	}

	templatePage(c, errorTmpl, code, obj)
}

// template404Page renders
// a standard 404 page.
func template404Page(
	c *gin.Context,
	instance *apimodel.InstanceV1,
	requestID string,
) {
	const notFoundTmpl = "404.tmpl"

	obj := map[string]any{
		"instance":  instance,
		"requestID": requestID,
	}

	templatePage(c, notFoundTmpl, http.StatusNotFound, obj)
}

// render the given template inside
// "page.tmpl" with the provided
// code and template object.
func templatePage(
	c *gin.Context,
	template string,
	code int,
	obj map[string]any,
) {
	const pageTmpl = "page.tmpl"
	obj["pageContent"] = template
	c.HTML(code, pageTmpl, obj)
}

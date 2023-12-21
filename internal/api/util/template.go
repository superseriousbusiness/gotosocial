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

// TemplatePage renders the given HTML template
// inside of the standard GtS "page" template.
//
// ogMeta, stylesheets, javascript, and any extra
// properties will be provided to the template if
// set, but can all be nil.
func TemplatePage(
	c *gin.Context,
	template string,
	instance *apimodel.InstanceV1,
	ogMeta *OGMeta,
	stylesheets []string,
	javascript []string,
	extra map[string]any,
) {
	obj := map[string]any{
		"instance":    instance,
		"ogMeta":      ogMeta,
		"stylesheets": stylesheets,
		"javascript":  javascript,
	}

	for k, v := range extra {
		obj[k] = v
	}

	templatePage(c, template, http.StatusOK, obj)
}

// templatePageError renders the given
// HTTP code, error, and request ID
// within the standard error template.
func templatePageError(
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

// templatePageNotFound renders
// a standard 404 page.
func templatePageNotFound(
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

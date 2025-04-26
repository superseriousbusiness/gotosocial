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

package web

import (
	"context"
	"net/http"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

func (m *Module) indexHandler(c *gin.Context) {
	instance, errWithCode := m.processor.InstanceGetV1(c.Request.Context())
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// Return instance we already got from the db,
	// don't try to fetch it again when erroring.
	instanceGet := func(ctx context.Context) (*apimodel.InstanceV1, gtserror.WithCode) {
		return instance, nil
	}

	// We only serve text/html at this endpoint.
	if _, err := apiutil.NegotiateAccept(c, apiutil.TextHTML); err != nil {
		apiutil.WebErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), instanceGet)
		return
	}

	// If a landingPageUser is set in the config, redirect to
	// that user's profile instead of rendering landing/index page.
	if landingPageUser := config.GetLandingPageUser(); landingPageUser != "" {
		c.Redirect(http.StatusFound, "/@"+strings.ToLower(landingPageUser))
		return
	}

	page := apiutil.WebPage{
		Template:    "index.tmpl",
		Instance:    instance,
		OGMeta:      apiutil.OGBase(instance),
		Stylesheets: []string{cssAbout, cssIndex},
		Extra: map[string]any{
			// Render "home to x
			// users [etc]" strap.
			"showStrap": true,
			// Show "log in" button
			// in top-right corner.
			"showLoginButton": true,
		},
	}

	apiutil.TemplateWebPage(c, page)
}

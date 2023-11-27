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

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

func (m *Module) tagGETHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// We'll need the instance later, and we can also use it
	// before then to make it easier to return a web error.
	instance, errWithCode := m.processor.InstanceGetV1(ctx)
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

	tagName, errWithCode := apiutil.ParseTagName(c.Param(apiutil.TagNameKey))
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return
	}

	stylesheets := []string{
		assetsPathPrefix + "/Fork-Awesome/css/fork-awesome.min.css",
		distPathPrefix + "/status.css",
		distPathPrefix + "/tag.css",
	}

	c.HTML(http.StatusOK, "tag.tmpl", gin.H{
		"instance":    instance,
		"ogMeta":      ogBase(instance),
		"tagName":     tagName,
		"stylesheets": stylesheets,
	})
}

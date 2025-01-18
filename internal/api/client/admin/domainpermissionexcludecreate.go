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

package admin

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// DomainPermissionExcludesPOSTHandler swagger:operation POST /api/v1/admin/domain_permission_excludes domainPermissionExcludeCreate
//
// Create a domain permission exclude with the given parameters.
//
// Excluded domains (and their subdomains) will not be automatically blocked or allowed when a list of domain permissions is imported or subscribed to.
//
// You can still manually create domain blocks or domain allows for excluded domains, and any new or existing domain blocks or domain allows for an excluded domain will still be enforced.
//
//	---
//	tags:
//	- admin
//
//	consumes:
//	- multipart/form-data
//	- application/json
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: domain
//		in: formData
//		description: Domain to create the permission exclude for.
//		type: string
//	-
//		name: private_comment
//		in: formData
//		description: >-
//			Private comment about this domain exclude.
//		type: string
//
//	security:
//	- OAuth2 Bearer:
//		- admin
//
//	responses:
//		'200':
//			description: The newly created domain permission exclude.
//			schema:
//				"$ref": "#/definitions/domainPermission"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'406':
//			description: not acceptable
//		'409':
//			description: conflict
//		'500':
//			description: internal server error
func (m *Module) DomainPermissionExcludesPOSTHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	// Parse + validate form.
	type ExcludeForm struct {
		Domain         string `form:"domain" json:"domain"`
		PrivateComment string `form:"private_comment" json:"private_comment"`
	}

	form := new(ExcludeForm)
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if form.Domain == "" {
		const errText = "domain must be set"
		errWithCode := gtserror.NewErrorBadRequest(errors.New(errText), errText)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	permExclude, errWithCode := m.processor.Admin().DomainPermissionExcludeCreate(
		c.Request.Context(),
		authed.Account,
		form.Domain,
		form.PrivateComment,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, permExclude)
}

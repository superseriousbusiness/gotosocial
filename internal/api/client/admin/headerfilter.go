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
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (m *Module) HeaderFilterAllowGET(c *gin.Context) {
	m.getHeaderFilter(c, m.processor.Admin().GetAllowHeaderFilter)
}

func (m *Module) HeaderFilterBlockGET(c *gin.Context) {
	m.getHeaderFilter(c, m.processor.Admin().GetBlockHeaderFilter)
}

func (m *Module) HeaderFilterAllowsGET(c *gin.Context) {
	m.getHeaderFilters(c, m.processor.Admin().GetAllowHeaderFilters)
}

func (m *Module) HeaderFilterBlocksGET(c *gin.Context) {
	m.getHeaderFilters(c, m.processor.Admin().GetBlockHeaderFilters)
}

func (m *Module) HeaderFilterAllowPOST(c *gin.Context) {
	m.createHeaderFilter(c, m.processor.Admin().CreateAllowHeaderFilter)
}

func (m *Module) HeaderFilterBlockPOST(c *gin.Context) {
	m.createHeaderFilter(c, m.processor.Admin().CreateBlockHeaderFilter)
}

func (m *Module) HeaderFilterAllowDELETE(c *gin.Context) {
	m.deleteHeaderFilter(c, m.processor.Admin().DeleteAllowHeaderFilter)
}

func (m *Module) HeaderFilterBlockDELETE(c *gin.Context) {
	m.deleteHeaderFilter(c, m.processor.Admin().DeleteAllowHeaderFilter)
}

func (m *Module) getHeaderFilter(c *gin.Context, get func(context.Context, string) (*apimodel.HeaderFilter, gtserror.WithCode)) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		errWithCode := gtserror.NewErrorUnauthorized(err, err.Error())
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		const text = "user not an admin"
		errWithCode := gtserror.NewErrorForbidden(errors.New(text), text)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		errWithCode := gtserror.NewErrorNotAcceptable(err, err.Error())
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	filterID := c.Param("id")
	if filterID == "" {
		const text = "no filter id specified"
		errWithCode := gtserror.NewErrorBadRequest(errors.New(text), text)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	filter, errWithCode := get(c.Request.Context(), filterID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, filter)
}

func (m *Module) getHeaderFilters(c *gin.Context, get func(context.Context) ([]*apimodel.HeaderFilter, gtserror.WithCode)) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		errWithCode := gtserror.NewErrorUnauthorized(err, err.Error())
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		const text = "user not an admin"
		errWithCode := gtserror.NewErrorForbidden(errors.New(text), text)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		errWithCode := gtserror.NewErrorNotAcceptable(err, err.Error())
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	filters, errWithCode := get(c.Request.Context())
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, filters)
}

func (m *Module) createHeaderFilter(c *gin.Context, create func(context.Context, *gtsmodel.Account, *apimodel.HeaderFilterRequest) (*apimodel.HeaderFilter, gtserror.WithCode)) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		errWithCode := gtserror.NewErrorUnauthorized(err, err.Error())
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		const text = "user not an admin"
		errWithCode := gtserror.NewErrorForbidden(errors.New(text), text)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		errWithCode := gtserror.NewErrorNotAcceptable(err, err.Error())
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	var form apimodel.HeaderFilterRequest

	if err := c.ShouldBind(&form); err != nil {
		errWithCode := gtserror.NewErrorBadRequest(err, err.Error())
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	filter, errWithCode := create(
		c.Request.Context(),
		authed.Account,
		&form,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, filter)
}

func (m *Module) deleteHeaderFilter(c *gin.Context, delete func(context.Context, string) gtserror.WithCode) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		errWithCode := gtserror.NewErrorUnauthorized(err, err.Error())
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		const text = "user not an admin"
		errWithCode := gtserror.NewErrorForbidden(errors.New(text), text)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		errWithCode := gtserror.NewErrorNotAcceptable(err, err.Error())
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	filterID := c.Param("id")
	if filterID == "" {
		const text = "no filter id specified"
		errWithCode := gtserror.NewErrorBadRequest(errors.New(text), text)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	errWithCode := delete(c.Request.Context(), filterID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}
}

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
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

func (m *Module) confirmEmailGETHandler(c *gin.Context) {
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

	// If there's no token in the query,
	// just serve the 404 web handler.
	token := c.Query("token")
	if token == "" {
		errWithCode := gtserror.NewErrorNotFound(errors.New(http.StatusText(http.StatusNotFound)))
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return
	}

	// Get user but don't confirm yet.
	user, errWithCode := m.processor.User().EmailGetUserForConfirmToken(c.Request.Context(), token)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return
	}

	// They may have already confirmed before
	// and are visiting the link again for
	// whatever reason. This is fine, just make
	// sure we have an email address to show them.
	email := user.UnconfirmedEmail
	if email == "" {
		// Already confirmed, take
		// that address instead.
		email = user.Email
	}

	// Serve page where user can click button
	// to POST confirmation to same endpoint.
	page := apiutil.WebPage{
		Template: "confirm-email.tmpl",
		Instance: instance,
		Extra: map[string]any{
			"email":    email,
			"username": user.Account.Username,
			"token":    token,
		},
	}

	apiutil.TemplateWebPage(c, page)
}

func (m *Module) confirmEmailPOSTHandler(c *gin.Context) {
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

	// If there's no token in the query,
	// just serve the 404 web handler.
	token := c.Query("token")
	if token == "" {
		errWithCode := gtserror.NewErrorNotFound(errors.New(http.StatusText(http.StatusNotFound)))
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return
	}

	// Confirm email address for real this time.
	user, errWithCode := m.processor.User().EmailConfirm(c.Request.Context(), token)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return
	}

	// Serve page informing user that their
	// email address is now confirmed.
	page := apiutil.WebPage{
		Template: "confirmed-email.tmpl",
		Instance: instance,
		Extra: map[string]any{
			"email":    user.Email,
			"username": user.Account.Username,
			"token":    token,
			"approved": *user.Approved,
		},
	}

	apiutil.TemplateWebPage(c, page)
}

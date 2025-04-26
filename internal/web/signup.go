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
	"net"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/validate"
	"github.com/gin-gonic/gin"
)

func (m *Module) signupGETHandler(c *gin.Context) {
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

	page := apiutil.WebPage{
		Template: "sign-up.tmpl",
		Instance: instance,
		OGMeta:   apiutil.OGBase(instance),
		Extra: map[string]any{
			"reasonRequired":   config.GetAccountsReasonRequired(),
			"registrationOpen": config.GetAccountsRegistrationOpen(),
		},
	}

	apiutil.TemplateWebPage(c, page)
}

func (m *Module) signupPOSTHandler(c *gin.Context) {
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

	form := &apimodel.AccountCreateRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.WebErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), instanceGet)
		return
	}

	if err := validate.CreateAccount(form); err != nil {
		apiutil.WebErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), instanceGet)
		return
	}

	clientIP := c.ClientIP()
	signUpIP := net.ParseIP(clientIP)
	if signUpIP == nil {
		err := errors.New("ip address could not be parsed from request")
		apiutil.WebErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), instanceGet)
		return
	}
	form.IP = signUpIP

	// We have all the info we need, call user+account create
	// (this will also trigger side effects like sending emails etc).
	user, errWithCode := m.processor.User().Create(
		c.Request.Context(),
		// nil to use
		// instance app.
		nil,
		form,
	)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return
	}

	// Serve a page informing the
	// user that they've signed up.
	page := apiutil.WebPage{
		Template: "signed-up.tmpl",
		Instance: instance,
		OGMeta:   apiutil.OGBase(instance),
		Extra: map[string]any{
			"email":    user.UnconfirmedEmail,
			"username": user.Account.Username,
		},
	}

	apiutil.TemplateWebPage(c, page)
}

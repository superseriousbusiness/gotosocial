/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (m *Module) threadGETHandler(c *gin.Context) {
	ctx := c.Request.Context()

	authed, err := oauth.Authed(c, false, false, false, false)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	// usernames on our instance will always be lowercase
	username := strings.ToLower(c.Param(usernameKey))
	if username == "" {
		err := errors.New("no account username specified")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	// status ids will always be uppercase
	statusID := strings.ToUpper(c.Param(statusIDKey))
	if statusID == "" {
		err := errors.New("no status id specified")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	host := config.GetHost()
	instance, err := m.processor.InstanceGet(ctx, host)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGet)
		return
	}

	instanceGet := func(ctx context.Context, domain string) (*apimodel.Instance, gtserror.WithCode) {
		return instance, nil
	}

	// do this check to make sure the status is actually from a local account,
	// we shouldn't render threads from statuses that don't belong to us!
	if _, errWithCode := m.processor.AccountGetLocalByUsername(ctx, authed, username); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, instanceGet)
		return
	}

	status, errWithCode := m.processor.StatusGet(ctx, authed, statusID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, instanceGet)
		return
	}

	if !strings.EqualFold(username, status.Account.Username) {
		err := gtserror.NewErrorNotFound(errors.New("path username not equal to status author username"))
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(err), instanceGet)
		return
	}

	// if we're getting an AP request on this endpoint we
	// should render the status's AP representation instead
	accept := c.NegotiateFormat(string(apiutil.TextHTML), string(apiutil.AppActivityJSON), string(apiutil.AppActivityLDJSON))
	if accept == string(apiutil.AppActivityJSON) || accept == string(apiutil.AppActivityLDJSON) {
		m.returnAPStatus(ctx, c, username, statusID, accept)
		return
	}

	context, errWithCode := m.processor.StatusGetContext(ctx, authed, statusID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, instanceGet)
		return
	}

	stylesheets := []string{
		assetsPathPrefix + "/Fork-Awesome/css/fork-awesome.min.css",
		distPathPrefix + "/status.css",
	}
	if config.GetAccountsAllowCustomCSS() {
		stylesheets = append(stylesheets, "/@"+username+"/custom.css")
	}

	c.HTML(http.StatusOK, "thread.tmpl", gin.H{
		"instance":    instance,
		"status":      status,
		"context":     context,
		"ogMeta":      ogBase(instance).withStatus(status),
		"stylesheets": stylesheets,
		"javascript":  []string{distPathPrefix + "/frontend.js"},
	})
}

func (m *Module) returnAPStatus(ctx context.Context, c *gin.Context, username string, statusID string, accept string) {
	verifier, signed := c.Get(string(ap.ContextRequestingPublicKeyVerifier))
	if signed {
		ctx = context.WithValue(ctx, ap.ContextRequestingPublicKeyVerifier, verifier)
	}

	signature, signed := c.Get(string(ap.ContextRequestingPublicKeySignature))
	if signed {
		ctx = context.WithValue(ctx, ap.ContextRequestingPublicKeySignature, signature)
	}

	status, errWithCode := m.processor.GetFediStatus(ctx, username, statusID, c.Request.URL)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet) //nolint:contextcheck
		return
	}

	b, mErr := json.Marshal(status)
	if mErr != nil {
		err := fmt.Errorf("could not marshal json: %s", mErr)
		apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGet) //nolint:contextcheck
		return
	}

	c.Data(http.StatusOK, accept, b)
}

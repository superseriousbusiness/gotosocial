/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

const (
	// MaxStatusIDKey is for specifying the maximum ID of the status to retrieve.
	MaxStatusIDKey = "max_id"
)

func (m *Module) profileGETHandler(c *gin.Context) {
	ctx := c.Request.Context()

	authed, err := oauth.Authed(c, false, false, false, false)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	username := strings.ToLower(c.Param(usernameKey))
	if username == "" {
		err := errors.New("no account username specified")
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

	account, errWithCode := m.processor.AccountGetLocalByUsername(ctx, authed, username)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, instanceGet)
		return
	}

	// if we're getting an AP request on this endpoint we
	// should render the account's AP representation instead
	accept := c.NegotiateFormat(string(apiutil.TextHTML), string(apiutil.AppActivityJSON), string(apiutil.AppActivityLDJSON))
	if accept == string(apiutil.AppActivityJSON) || accept == string(apiutil.AppActivityLDJSON) {
		m.returnAPProfile(ctx, c, username, accept)
		return
	}

	var rssFeed string
	if account.EnableRSS {
		rssFeed = "/@" + account.Username + "/feed.rss"
	}

	// only allow search engines / robots to view this page if account is discoverable
	var robotsMeta string
	if account.Discoverable {
		robotsMeta = robotsMetaAllowSome
	}

	// we should only show the 'back to top' button if the
	// profile visitor is paging through statuses
	showBackToTop := false

	maxStatusID := ""
	maxStatusIDString := c.Query(MaxStatusIDKey)
	if maxStatusIDString != "" {
		maxStatusID = maxStatusIDString
		showBackToTop = true
	}

	statusResp, errWithCode := m.processor.AccountWebStatusesGet(ctx, account.ID, maxStatusID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, instanceGet)
		return
	}

	stylesheets := []string{
		assetsPathPrefix + "/Fork-Awesome/css/fork-awesome.min.css",
		distPathPrefix + "/status.css",
		distPathPrefix + "/profile.css",
	}
	if config.GetAccountsAllowCustomCSS() {
		stylesheets = append(stylesheets, "/@"+account.Username+"/custom.css")
	}

	c.HTML(http.StatusOK, "profile.tmpl", gin.H{
		"instance":         instance,
		"account":          account,
		"ogMeta":           ogBase(instance).withAccount(account),
		"rssFeed":          rssFeed,
		"robotsMeta":       robotsMeta,
		"statuses":         statusResp.Items,
		"statuses_next":    statusResp.NextLink,
		"show_back_to_top": showBackToTop,
		"stylesheets":      stylesheets,
		"javascript":       []string{distPathPrefix + "/frontend.js"},
	})
}

func (m *Module) returnAPProfile(ctx context.Context, c *gin.Context, username string, accept string) {
	verifier, signed := c.Get(string(ap.ContextRequestingPublicKeyVerifier))
	if signed {
		ctx = context.WithValue(ctx, ap.ContextRequestingPublicKeyVerifier, verifier)
	}

	signature, signed := c.Get(string(ap.ContextRequestingPublicKeySignature))
	if signed {
		ctx = context.WithValue(ctx, ap.ContextRequestingPublicKeySignature, signature)
	}

	user, errWithCode := m.processor.GetFediUser(ctx, username, c.Request.URL)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet) //nolint:contextcheck
		return
	}

	b, mErr := json.Marshal(user)
	if mErr != nil {
		err := fmt.Errorf("could not marshal json: %s", mErr)
		apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGet) //nolint:contextcheck
		return
	}

	c.Data(http.StatusOK, accept, b)
}

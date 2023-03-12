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
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	username := strings.ToLower(c.Param(usernameKey))
	if username == "" {
		err := errors.New("no account username specified")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	instance, err := m.processor.InstanceGetV1(ctx)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGetV1)
		return
	}

	instanceGet := func(ctx context.Context) (*apimodel.InstanceV1, gtserror.WithCode) {
		return instance, nil
	}

	account, errWithCode := m.processor.Account().GetLocalByUsername(ctx, authed.Account, username)
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

	// We need to change our response slightly if the
	// profile visitor is paging through statuses.
	var (
		paging      bool
		pinnedResp  = &apimodel.PageableResponse{}
		maxStatusID string
	)

	if maxStatusIDString := c.Query(MaxStatusIDKey); maxStatusIDString != "" {
		maxStatusID = maxStatusIDString
		paging = true
	}

	statusResp, errWithCode := m.processor.Account().WebStatusesGet(ctx, account.ID, maxStatusID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, instanceGet)
		return
	}

	// If we're not paging, then the profile visitor
	// is currently just opening the bare profile, so
	// load pinned statuses so we can show them at the
	// top of the profile.
	if !paging {
		pinnedResp, errWithCode = m.processor.Account().StatusesGet(ctx, authed.Account, account.ID, 0, false, false, "", "", true, false, false)
		if errWithCode != nil {
			apiutil.ErrorHandler(c, errWithCode, instanceGet)
			return
		}
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
		"pinned_statuses":  pinnedResp.Items,
		"show_back_to_top": paging,
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

	user, errWithCode := m.processor.Fedi().UserGet(ctx, username, c.Request.URL)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1) //nolint:contextcheck
		return
	}

	b, mErr := json.Marshal(user)
	if mErr != nil {
		err := fmt.Errorf("could not marshal json: %s", mErr)
		apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGetV1) //nolint:contextcheck
		return
	}

	c.Data(http.StatusOK, accept, b)
}

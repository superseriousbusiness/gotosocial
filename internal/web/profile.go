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
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (m *Module) profileGETHandler(c *gin.Context) {
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

	// Parse account targetUsername from the URL.
	targetUsername, errWithCode := apiutil.ParseWebUsername(c.Param(apiutil.WebUsernameKey))
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return
	}

	// Normalize requested username:
	//
	//   - Usernames on our instance are (currently) always lowercase.
	//
	// todo: Update this logic when different username patterns
	// are allowed, and/or when status slugs are introduced.
	targetUsername = strings.ToLower(targetUsername)

	// Check what type of content is being requested. If we're getting an AP
	// request on this endpoint we should render the AP representation instead.
	accept, err := apiutil.NegotiateAccept(c, apiutil.HTMLOrActivityPubHeaders...)
	if err != nil {
		apiutil.WebErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), instanceGet)
		return
	}

	if accept == string(apiutil.AppActivityJSON) || accept == string(apiutil.AppActivityLDJSON) {
		// AP account representation has been requested.
		m.returnAPAccount(c, targetUsername, accept, instanceGet)
		return
	}

	// text/html has been requested. Proceed with getting the web view of the account.

	// Don't require auth for web endpoints, but do take it if it was provided.
	// authed.Account might end up nil here, but that's fine in case of public pages.
	authed, err := oauth.Authed(c, false, false, false, false)
	if err != nil {
		apiutil.WebErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	// Fetch the target account so we can do some checks on it.
	targetAccount, errWithCode := m.processor.Account().GetLocalByUsername(ctx, authed.Account, targetUsername)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return
	}

	// If target account is suspended, this page should not be visible.
	// TODO: change this to 410?
	if targetAccount.Suspended {
		err := fmt.Errorf("target account %s is suspended", targetUsername)
		apiutil.WebErrorHandler(c, gtserror.NewErrorNotFound(err), instanceGet)
		return
	}

	// Only generate RSS link if account has RSS enabled.
	var rssFeed string
	if targetAccount.EnableRSS {
		rssFeed = "/@" + targetAccount.Username + "/feed.rss"
	}

	// Only allow search engines / robots to
	// index if account is discoverable.
	var robotsMeta string
	if targetAccount.Discoverable {
		robotsMeta = robotsMetaAllowSome
	}

	// We need to change our response slightly if the
	// profile visitor is paging through statuses.
	var (
		maxStatusID    = apiutil.ParseMaxID(c.Query(apiutil.MaxIDKey), "")
		paging         = maxStatusID != ""
		pinnedStatuses []*apimodel.Status
	)

	if !paging {
		// Client opened bare profile (from the top)
		// so load + display pinned statuses.
		pinnedStatuses, errWithCode = m.processor.Account().WebStatusesGetPinned(ctx, targetAccount.ID)
		if errWithCode != nil {
			apiutil.WebErrorHandler(c, errWithCode, instanceGet)
			return
		}
	}

	// Get statuses from maxStatusID onwards (or from top if empty string).
	statusResp, errWithCode := m.processor.Account().WebStatusesGet(ctx, targetAccount.ID, maxStatusID)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return
	}

	stylesheets := []string{
		assetsPathPrefix + "/Fork-Awesome/css/fork-awesome.min.css",
		distPathPrefix + "/status.css",
		distPathPrefix + "/profile.css",
	}
	if config.GetAccountsAllowCustomCSS() {
		stylesheets = append(stylesheets, "/@"+targetAccount.Username+"/custom.css")
	}

	c.HTML(http.StatusOK, "profile.tmpl", gin.H{
		"instance":         instance,
		"account":          targetAccount,
		"ogMeta":           ogBase(instance).withAccount(targetAccount),
		"rssFeed":          rssFeed,
		"robotsMeta":       robotsMeta,
		"statuses":         statusResp.Items,
		"statuses_next":    statusResp.NextLink,
		"pinned_statuses":  pinnedStatuses,
		"show_back_to_top": paging,
		"stylesheets":      stylesheets,
		"javascript":       []string{distPathPrefix + "/frontend.js"},
	})
}

// returnAPAccount returns an ActivityPub representation of
// target account. It will do http signature authentication.
func (m *Module) returnAPAccount(
	c *gin.Context,
	targetUsername string,
	accept string,
	instanceGet func(ctx context.Context) (*apimodel.InstanceV1, gtserror.WithCode),
) {
	user, errWithCode := m.processor.Fedi().UserGet(c.Request.Context(), targetUsername, c.Request.URL)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	b, err := json.Marshal(user)
	if err != nil {
		err := gtserror.Newf("could not marshal json: %w", err)
		apiutil.WebErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGetV1)
		return
	}

	c.Data(http.StatusOK, accept, b)
}

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
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

func (m *Module) threadGETHandler(c *gin.Context) {
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

	// Parse account targetUsername and status ID from the URL.
	targetUsername, errWithCode := apiutil.ParseUsername(c.Param(apiutil.UsernameKey))
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return
	}

	targetStatusID, errWithCode := apiutil.ParseWebStatusID(c.Param(apiutil.WebStatusIDKey))
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return
	}

	// Normalize requested username + status ID:
	//
	//   - Usernames on our instance are (currently) always lowercase.
	//   - StatusIDs on our instance are (currently) always ULIDs.
	//
	// todo: Update this logic when different username patterns
	// are allowed, and/or when status slugs are introduced.
	targetUsername = strings.ToLower(targetUsername)
	targetStatusID = strings.ToUpper(targetStatusID)

	// Check what type of content is being requested. If we're getting an AP
	// request on this endpoint we should render the AP representation instead.
	accept, err := apiutil.NegotiateAccept(c, apiutil.HTMLOrActivityPubHeaders...)
	if err != nil {
		apiutil.WebErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), instanceGet)
		return
	}

	if accept == string(apiutil.AppActivityJSON) || accept == string(apiutil.AppActivityLDJSON) {
		// AP status representation has been requested.
		m.returnAPStatus(c, targetUsername, targetStatusID, accept, instanceGet)
		return
	}

	// text/html has been requested. Proceed with getting the web view of the status.

	// Fetch the target account so we can do some checks on it.
	targetAccount, errWithCode := m.processor.Account().GetWeb(ctx, targetUsername)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return
	}

	// If target account is suspended, this page should not be visible.
	if targetAccount.Suspended {
		err := fmt.Errorf("target account %s is suspended", targetUsername)
		apiutil.WebErrorHandler(c, gtserror.NewErrorNotFound(err), instanceGet)
		return
	}

	// Get the thread context. This will fetch the target status as well.
	context, errWithCode := m.processor.Status().WebContextGet(ctx, targetStatusID)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return
	}

	// Ensure status actually belongs to target account.
	if context.Status.Account.ID != targetAccount.ID {
		err := fmt.Errorf("target account %s does not own status %s", targetUsername, targetStatusID)
		apiutil.WebErrorHandler(c, gtserror.NewErrorNotFound(err), instanceGet)
		return
	}

	// Prepare stylesheets for thread.
	stylesheets := make([]string, 0, 6)

	// Basic thread stylesheets.
	stylesheets = append(
		stylesheets,
		[]string{
			cssFA,
			cssStatus,
			cssThread,
		}...,
	)

	// User-selected theme if set.
	if theme := targetAccount.Theme; theme != "" {
		stylesheets = append(
			stylesheets,
			themesPathPrefix+"/"+theme,
		)
	}

	// Custom CSS for this user last in cascade.
	stylesheets = append(
		stylesheets,
		"/@"+targetAccount.Username+"/custom.css",
	)

	page := apiutil.WebPage{
		Template:    "thread.tmpl",
		Instance:    instance,
		OGMeta:      apiutil.OGBase(instance).WithStatus(context.Status),
		Stylesheets: stylesheets,
		Javascript:  []string{jsFrontend},
		Extra: map[string]any{
			"context": context,
		},
	}

	apiutil.TemplateWebPage(c, page)
}

// returnAPStatus returns an ActivityPub representation of target status,
// created by targetUsername. It will do http signature authentication.
func (m *Module) returnAPStatus(
	c *gin.Context,
	targetUsername string,
	targetStatusID string,
	accept string,
	instanceGet func(ctx context.Context) (*apimodel.InstanceV1, gtserror.WithCode),
) {
	status, errWithCode := m.processor.Fedi().StatusGet(c.Request.Context(), targetUsername, targetStatusID)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, errWithCode, instanceGet)
		return
	}

	b, err := json.Marshal(status)
	if err != nil {
		err := gtserror.Newf("could not marshal json: %w", err)
		apiutil.WebErrorHandler(c, gtserror.NewErrorInternalError(err), instanceGet)
		return
	}

	c.Data(http.StatusOK, accept, b)
}

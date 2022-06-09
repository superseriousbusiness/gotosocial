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
	"math/rand"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (m *Module) profileGETHandler(c *gin.Context) {
	ctx := c.Request.Context()

	authed, err := oauth.Authed(c, false, false, false, false)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	// usernames on our instance will always be lowercase
	username := strings.ToLower(c.Param(usernameKey))
	if username == "" {
		err := errors.New("no account username specified")
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	host := config.GetHost()
	instance, err := m.processor.InstanceGet(ctx, host)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGet)
		return
	}

	instanceGet := func(ctx context.Context, domain string) (*apimodel.Instance, gtserror.WithCode) {
		return instance, nil
	}

	account, errWithCode := m.processor.AccountGetLocalByUsername(ctx, authed, username)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, instanceGet)
		return
	}

	// if we're getting an AP request on this endpoint we
	// should render the account's AP representation instead
	accept := c.NegotiateFormat(string(api.TextHTML), string(api.AppActivityJSON), string(api.AppActivityLDJSON))
	if accept == string(api.AppActivityJSON) || accept == string(api.AppActivityLDJSON) {
		m.returnAPProfile(ctx, c, username, accept)
		return
	}

	// get latest 10 top-level public statuses;
	// ie., exclude replies and boosts, public only,
	// with or without media
	statusResp, errWithCode := m.processor.AccountStatusesGet(ctx, authed, account.ID, 10, true, true, "", "", false, false, true)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, instanceGet)
		return
	}

	// pick a random dummy avatar if this account avatar isn't set yet
	if account.Avatar == "" && len(m.defaultAvatars) > 0 {
		//nolint:gosec
		randomIndex := rand.Intn(len(m.defaultAvatars))
		dummyAvatar := m.defaultAvatars[randomIndex]
		account.Avatar = dummyAvatar
		for _, i := range statusResp.Items {
			s, ok := i.(*apimodel.Status)
			if !ok {
				panic("timelineable was not *apimodel.Status")
			}
			s.Account.Avatar = dummyAvatar
		}
	}

	c.HTML(http.StatusOK, "profile.tmpl", gin.H{
		"instance": instance,
		"account":  account,
		"statuses": statusResp.Items,
		"stylesheets": []string{
			"/assets/Fork-Awesome/css/fork-awesome.min.css",
			"/assets/dist/status.css",
			"/assets/dist/profile.css",
		},
		"javascript": []string{
			"/assets/dist/frontend.js",
		},
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
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	b, mErr := json.Marshal(user)
	if mErr != nil {
		err := fmt.Errorf("could not marshal json: %s", mErr)
		api.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGet)
		return
	}

	c.Data(http.StatusOK, accept, b)
}

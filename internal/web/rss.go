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
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (m *Module) rssFeedGETHandler(c *gin.Context) {
	ctx := c.Request.Context()

	host := config.GetHost()
	instance, errWithCode := m.processor.InstanceGet(ctx, host)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	instanceGet := func(ctx context.Context, domain string) (*apimodel.Instance, gtserror.WithCode) {
		return instance, nil
	}

	authed, err := oauth.Authed(c, false, false, false, false)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), instanceGet)
		return
	}

	if _, err := api.NegotiateAccept(c, api.AppRSSXML); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), instanceGet)
		return
	}

	// usernames on our instance will always be lowercase
	username := strings.ToLower(c.Param(usernameKey))
	if username == "" {
		err := errors.New("no account username specified")
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), instanceGet)
		return
	}

	rssFeed, errWithCode := m.processor.AccountGetRSSFeedForUsername(ctx, authed, username)
	if err != nil {
		api.ErrorHandler(c, errWithCode, instanceGet)
		return
	}

	c.Data(http.StatusOK, string(api.AppRSSXML+"; charset=utf-8"), []byte(rssFeed))
}

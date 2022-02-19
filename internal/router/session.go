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

package router

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

// SessionOptions returns the standard set of options to use for each session.
func SessionOptions() sessions.Options {
	return sessions.Options{
		Path:     "/",
		Domain:   viper.GetString(config.Keys.Host),
		MaxAge:   120,                                              // 2 minutes
		Secure:   viper.GetString(config.Keys.Protocol) == "https", // only use cookie over https
		HttpOnly: true,                                             // exclude javascript from inspecting cookie
		SameSite: http.SameSiteDefaultMode,                         // https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-cookie-same-site-00#section-4.1.1
	}
}

// SessionName is a utility function that derives an appropriate session name from the hostname.
func SessionName() (string, error) {
	// parse the protocol + host
	protocol := viper.GetString(config.Keys.Protocol)
	host := viper.GetString(config.Keys.Host)
	u, err := url.Parse(fmt.Sprintf("%s://%s", protocol, host))
	if err != nil {
		return "", err
	}

	// take the hostname without any port attached
	strippedHostname := u.Hostname()
	if strippedHostname == "" {
		return "", fmt.Errorf("could not derive hostname without port from %s://%s", protocol, host)
	}

	return fmt.Sprintf("gotosocial-%s", strippedHostname), nil
}

func useSession(ctx context.Context, sessionDB db.Session, engine *gin.Engine) error {
	// check if we have a saved router session already
	rs, err := sessionDB.GetSession(ctx)
	if err != nil {
		return fmt.Errorf("error using session: %s", err)
	}
	if rs == nil || rs.Auth == nil || rs.Crypt == nil {
		return errors.New("router session was nil")
	}

	store := memstore.NewStore(rs.Auth, rs.Crypt)
	store.Options(SessionOptions())

	sessionName, err := SessionName()
	if err != nil {
		return err
	}

	engine.Use(sessions.Sessions(sessionName, store))
	return nil
}

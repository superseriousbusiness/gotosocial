/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

// SessionOptions returns the standard set of options to use for each session.
func SessionOptions(cfg *config.Config) sessions.Options {
	return sessions.Options{
		Path:     "/",
		Domain:   cfg.Host,
		MaxAge:   120,                      // 2 minutes
		Secure:   true,                     // only use cookie over https
		HttpOnly: true,                     // exclude javascript from inspecting cookie
		SameSite: http.SameSiteDefaultMode, // https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-cookie-same-site-00#section-4.1.1
	}
}

func useSession(ctx context.Context, cfg *config.Config, sessionDB db.Session, engine *gin.Engine) error {
	// check if we have a saved router session already
	rs, err := sessionDB.GetSession(ctx)
	if err != nil {
		return fmt.Errorf("error using session: %s", err)
	}
	if rs == nil || rs.Auth == nil || rs.Crypt == nil {
		return errors.New("router session was nil")
	}

	store := memstore.NewStore(rs.Auth, rs.Crypt)
	store.Options(SessionOptions(cfg))
	sessionName := fmt.Sprintf("gotosocial-%s", cfg.Host)
	engine.Use(sessions.Sessions(sessionName, store))
	return nil
}

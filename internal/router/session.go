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
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

func useSession(cfg *config.Config, dbService db.DB, engine *gin.Engine) error {
	// check if we have a saved router session already
	routerSessions := []*gtsmodel.RouterSession{}
	if err := dbService.GetAll(&routerSessions); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			// proper error occurred
			return err
		}
	}

	var rs *gtsmodel.RouterSession
	if len(routerSessions) == 1 {
		// we have a router session stored
		rs = routerSessions[0]
	} else if len(routerSessions) == 0 {
		// we have no router sessions so we need to create a new one
		var err error
		rs, err = routerSession(dbService)
		if err != nil {
			return fmt.Errorf("error creating new router session: %s", err)
		}
	} else {
		// we should only have one router session stored ever
		return errors.New("we had more than one router session in the db")
	}

	if rs == nil {
		return errors.New("error getting or creating router session: router session was nil")
	}

	store := memstore.NewStore(rs.Auth, rs.Crypt)
	store.Options(sessions.Options{
		Path:     "/",
		Domain:   cfg.Host,
		MaxAge:   120,                     // 2 minutes
		Secure:   true,                    // only use cookie over https
		HttpOnly: true,                    // exclude javascript from inspecting cookie
		SameSite: http.SameSiteStrictMode, // https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-cookie-same-site-00#section-4.1.1
	})
	sessionName := fmt.Sprintf("gotosocial-%s", cfg.Host)
	engine.Use(sessions.Sessions(sessionName, store))
	return nil
}

// routerSession generates a new router session with random auth and crypt bytes,
// puts it in the database for persistence, and returns it for use.
func routerSession(dbService db.DB) (*gtsmodel.RouterSession, error) {
	auth := make([]byte, 32)
	crypt := make([]byte, 32)

	if _, err := rand.Read(auth); err != nil {
		return nil, err
	}
	if _, err := rand.Read(crypt); err != nil {
		return nil, err
	}

	rid, err := id.NewULID()
	if err != nil {
		return nil, err
	}

	rs := &gtsmodel.RouterSession{
		ID:    rid,
		Auth:  auth,
		Crypt: crypt,
	}

	if err := dbService.Put(rs); err != nil {
		return nil, err
	}

	return rs, nil
}

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

package app

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/module"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/pkg/mastotypes"
)

const appsPath = "/api/v1/apps"

type appModule struct {
	server oauth.Server
	db     db.DB
	log    *logrus.Logger
}

// New returns a new auth module
func New(srv oauth.Server, db db.DB, log *logrus.Logger) module.ClientAPIModule {
	return &appModule{
		server: srv,
		db:     db,
		log:    log,
	}
}

// Route satisfies the RESTAPIModule interface
func (m *appModule) Route(s router.Router) error {
	s.AttachHandler(http.MethodPost, appsPath, m.appsPOSTHandler)
	return nil
}

// appsPOSTHandler should be served at https://example.org/api/v1/apps
// It is equivalent to: https://docs.joinmastodon.org/methods/apps/
func (m *appModule) appsPOSTHandler(c *gin.Context) {
	l := m.log.WithField("func", "AppsPOSTHandler")
	l.Trace("entering AppsPOSTHandler")

	form := &mastotypes.ApplicationPOSTRequest{}
	if err := c.ShouldBind(form); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// permitted length for most fields
	permittedLength := 64
	// redirect can be a bit bigger because we probably need to encode data in the redirect uri
	permittedRedirect := 256

	// check lengths of fields before proceeding so the user can't spam huge entries into the database
	if len(form.ClientName) > permittedLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("client_name must be less than %d bytes", permittedLength)})
		return
	}
	if len(form.Website) > permittedLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("website must be less than %d bytes", permittedLength)})
		return
	}
	if len(form.RedirectURIs) > permittedRedirect {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("redirect_uris must be less than %d bytes", permittedRedirect)})
		return
	}
	if len(form.Scopes) > permittedLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("scopes must be less than %d bytes", permittedLength)})
		return
	}

	// set default 'read' for scopes if it's not set, this follows the default of the mastodon api https://docs.joinmastodon.org/methods/apps/
	var scopes string
	if form.Scopes == "" {
		scopes = "read"
	} else {
		scopes = form.Scopes
	}

	// generate new IDs for this application and its associated client
	clientID := uuid.NewString()
	clientSecret := uuid.NewString()
	vapidKey := uuid.NewString()

	// generate the application to put in the database
	app := &model.Application{
		Name:         form.ClientName,
		Website:      form.Website,
		RedirectURI:  form.RedirectURIs,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		VapidKey:     vapidKey,
	}

	// chuck it in the db
	if err := m.db.Put(app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// now we need to model an oauth client from the application that the oauth library can use
	oc := &oauth.Client{
		ID:     clientID,
		Secret: clientSecret,
		Domain: form.RedirectURIs,
		UserID: "", // This client isn't yet associated with a specific user,  it's just an app client right now
	}

	// chuck it in the db
	if err := m.db.Put(oc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// done, return the new app information per the spec here: https://docs.joinmastodon.org/methods/apps/
	c.JSON(http.StatusOK, app.ToMasto())
}

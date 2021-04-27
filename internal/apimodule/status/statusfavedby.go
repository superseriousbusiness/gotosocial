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

package status

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/mastotypes"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// StatusFavedByGETHandler is for serving a list of accounts that have faved a given status
func (m *Module) StatusFavedByGETHandler(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
		"func":        "statusGETHandler",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})
	l.Debugf("entering function")

	var requestingAccount *gtsmodel.Account
	authed, err := oauth.MustAuth(c, true, false, true, true) // we don't really need an app here but we want everything else
	if err != nil {
		l.Debug("not authed but will continue to serve anyway if public status")
		requestingAccount = nil
	} else {
		requestingAccount = authed.Account
	}

	targetStatusID := c.Param(IDKey)
	if targetStatusID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no status id provided"})
		return
	}

	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := m.db.GetByID(targetStatusID, targetStatus); err != nil {
		l.Errorf("error fetching status %s: %s", targetStatusID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("status %s not found", targetStatusID)})
		return
	}

	l.Tracef("going to search for target account %s", targetStatus.AccountID)
	targetAccount := &gtsmodel.Account{}
	if err := m.db.GetByID(targetStatus.AccountID, targetAccount); err != nil {
		l.Errorf("error fetching target account %s: %s", targetStatus.AccountID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("status %s not found", targetStatusID)})
		return
	}

	l.Trace("going to get relevant accounts")
	relevantAccounts, err := m.db.PullRelevantAccountsFromStatus(targetStatus)
	if err != nil {
		l.Errorf("error fetching related accounts for status %s: %s", targetStatusID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("status %s not found", targetStatusID)})
		return
	}

	l.Trace("going to see if status is visible")
	visible, err := m.db.StatusVisible(targetStatus, targetAccount, requestingAccount, relevantAccounts) // requestingAccount might well be nil here, but StatusVisible knows how to take care of that
	if err != nil {
		l.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("status %s not found", targetStatusID)})
		return
	}

	if !visible {
		l.Trace("status is not visible")
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("status %s not found", targetStatusID)})
		return
	}

	// get ALL accounts that faved a status -- doesn't take account of blocks and mutes and stuff
	favingAccounts, err := m.db.WhoFavedStatus(targetStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// filter the list so the user doesn't see accounts they blocked or which blocked them
	filteredAccounts := []*gtsmodel.Account{}
	for _, acc := range favingAccounts {
		blocked, err := m.db.Blocked(authed.Account.ID, acc.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !blocked {
			filteredAccounts = append(filteredAccounts, acc)
		}
	}

	// TODO: filter other things here? suspended? muted? silenced?

	// now we can return the masto representation of those accounts
	mastoAccounts := []*mastotypes.Account{}
	for _, acc := range filteredAccounts {
		mastoAccount, err := m.tc.AccountToMastoPublic(acc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		mastoAccounts = append(mastoAccounts, mastoAccount)
	}

	c.JSON(http.StatusOK, mastoAccounts)
}

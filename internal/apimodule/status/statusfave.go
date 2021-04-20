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
	"github.com/superseriousbusiness/gotosocial/internal/distributor"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// StatusFavePOSTHandler handles fave requests against a given status ID
func (m *Module) StatusFavePOSTHandler(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
		"func":        "StatusFavePOSTHandler",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})
	l.Debugf("entering function")

	authed, err := oauth.MustAuth(c, true, false, true, true) // we don't really need an app here but we want everything else
	if err != nil {
		l.Debug("not authed so can't fave status")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
		return
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
	visible, err := m.db.StatusVisible(targetStatus, targetAccount, authed.Account, relevantAccounts) // requestingAccount might well be nil here, but StatusVisible knows how to take care of that
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

	// is the status faveable?
	if !targetStatus.VisibilityAdvanced.Likeable {
		l.Debug("status is not faveable")
		c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("status %s not faveable", targetStatusID)})
		return
	}

	// it's visible! it's faveable! so let's fave the FUCK out of it
	fave, err := m.db.FaveStatus(targetStatus, authed.Account.ID)
	if err != nil {
		l.Debugf("error faveing status: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var boostOfStatus *gtsmodel.Status
	if targetStatus.BoostOfID != "" {
		boostOfStatus = &gtsmodel.Status{}
		if err := m.db.GetByID(targetStatus.BoostOfID, boostOfStatus); err != nil {
			l.Errorf("error fetching boosted status %s: %s", targetStatus.BoostOfID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("status %s not found", targetStatusID)})
			return
		}
	}

	mastoStatus, err := m.mastoConverter.StatusToMasto(targetStatus, targetAccount, authed.Account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, boostOfStatus)
	if err != nil {
		l.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("status %s not found", targetStatusID)})
		return
	}

	// if the targeted status was already faved, faved will be nil
	// only put the fave in the distributor if something actually changed
	if fave != nil {
		fave.FavedStatus = targetStatus // attach the status pointer to the fave for easy retrieval in the distributor
		m.distributor.FromClientAPI() <- distributor.FromClientAPI{
			APObjectType:   gtsmodel.ActivityStreamsNote, // status is a note
			APActivityType: gtsmodel.ActivityStreamsLike, // we're creating a like/fave on the note
			Activity:       fave,                         // pass the fave along for processing
		}
	}

	c.JSON(http.StatusOK, mastoStatus)
}

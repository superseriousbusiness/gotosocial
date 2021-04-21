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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/distributor"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// StatusReblogPOSTHandler handles boost/reblog requests against a given status ID
func (m *Module) StatusReblogPOSTHandler(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
		"func":        "StatusReblogPOSTHandler",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})
	l.Debugf("entering function")

	authed, err := oauth.MustAuth(c, true, false, true, true) // we don't really need an app here but we want everything else
	if err != nil {
		l.Debug("not authed so can't boost status")
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
		l.Trace("status is not visible so cannot be boosted")
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("status %s not found", targetStatusID)})
		return
	}

	// is the status boostable?
	if !targetStatus.VisibilityAdvanced.Boostable {
		l.Debug("status is not boostable")
		c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("status %s not boostable", targetStatusID)})
		return
	}

	/*
		FROM THIS POINT ONWARDS WE ARE HAPPY WITH THE BOOST -- it is valid and we will try to create it
	*/

	// it's visible! it's boostable! so let's boost the FUCK out of it
	// first we create a new status and add some basic info to it -- this will be the wrapper for the boosted status

	// the wrapper won't use the same ID as the boosted status so we generate some new UUIDs
	uris := util.GenerateURIs(authed.Account.Username, m.config.Protocol, m.config.Host)
	boostWrapperStatusID := uuid.NewString()
	boostWrapperStatusURI := fmt.Sprintf("%s/%s", uris.StatusesURI, boostWrapperStatusID)
	boostWrapperStatusURL := fmt.Sprintf("%s/%s", uris.StatusesURL, boostWrapperStatusID)

	boostWrapperStatus := &gtsmodel.Status{
		ID:  boostWrapperStatusID,
		URI: boostWrapperStatusURI,
		URL: boostWrapperStatusURL,

		// the boosted status is not created now, but the boost certainly is
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
		Local:                    true, // always local since this is being done through the client API
		AccountID:                authed.Account.ID,
		CreatedWithApplicationID: authed.Application.ID,

		// replies can be boosted, but boosts are never replies
		InReplyToID:        "",
		InReplyToAccountID: "",

		// these will all be wrapped in the boosted status so set them empty here
		Attachments: []string{},
		Tags:        []string{},
		Mentions:    []string{},
		Emojis:      []string{},

		// the below fields will be taken from the target status
		Content:             util.HTMLFormat(targetStatus.Content), // take content from target status
		ContentWarning:      targetStatus.ContentWarning,           // same warning as the target status
		ActivityStreamsType: targetStatus.ActivityStreamsType,      // same activitystreams type as target status
		Sensitive:           targetStatus.Sensitive,
		Language:            targetStatus.Language,
		Text:                targetStatus.Text,
		BoostOfID:           targetStatus.ID,
		Visibility:          targetStatus.Visibility,
		VisibilityAdvanced:  targetStatus.VisibilityAdvanced,

		// attach these here for convenience -- the boosted status/account won't go in the DB
		// but they're needed in the distributor and for the frontend. Since we have them, we can
		// attach them so we don't need to fetch them again later (save some DB calls)
		GTSBoostedStatus:  targetStatus,
		GTSBoostedAccount: targetAccount,
	}

	// put the boost in the database
	if err := m.db.Put(boostWrapperStatus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// pass to the distributor to take care of side effects asynchronously -- federation, mentions, updating metadata, etc, etc
	m.distributor.FromClientAPI() <- distributor.FromClientAPI{
		APObjectType:   gtsmodel.ActivityStreamsNote,
		APActivityType: gtsmodel.ActivityStreamsAnnounce, // boost/reblog is an 'announce' activity
		Activity:       boostWrapperStatus,
	}

	// return the frontend representation of the new status to the submitter
	mastoStatus, err := m.mastoConverter.StatusToMasto(boostWrapperStatus, authed.Account, authed.Account, targetAccount, nil, targetStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mastoStatus)
}

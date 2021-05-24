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

package message

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) StatusCreate(auth *oauth.Auth, form *apimodel.AdvancedStatusCreateForm) (*apimodel.Status, error) {
	uris := util.GenerateURIsForAccount(auth.Account.Username, p.config.Protocol, p.config.Host)
	thisStatusID := uuid.NewString()
	thisStatusURI := fmt.Sprintf("%s/%s", uris.StatusesURI, thisStatusID)
	thisStatusURL := fmt.Sprintf("%s/%s", uris.StatusesURL, thisStatusID)
	newStatus := &gtsmodel.Status{
		ID:                       thisStatusID,
		URI:                      thisStatusURI,
		URL:                      thisStatusURL,
		Content:                  util.HTMLFormat(form.Status),
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
		Local:                    true,
		AccountID:                auth.Account.ID,
		ContentWarning:           form.SpoilerText,
		ActivityStreamsType:      gtsmodel.ActivityStreamsNote,
		Sensitive:                form.Sensitive,
		Language:                 form.Language,
		CreatedWithApplicationID: auth.Application.ID,
		Text:                     form.Status,
	}

	// check if replyToID is ok
	if err := p.processReplyToID(form, auth.Account.ID, newStatus); err != nil {
		return nil, err
	}

	// check if mediaIDs are ok
	if err := p.processMediaIDs(form, auth.Account.ID, newStatus); err != nil {
		return nil, err
	}

	// check if visibility settings are ok
	if err := p.processVisibility(form, auth.Account.Privacy, newStatus); err != nil {
		return nil, err
	}

	// handle language settings
	if err := p.processLanguage(form, auth.Account.Language, newStatus); err != nil {
		return nil, err
	}

	// handle mentions
	if err := p.processMentions(form, auth.Account.ID, newStatus); err != nil {
		return nil, err
	}

	if err := p.processTags(form, auth.Account.ID, newStatus); err != nil {
		return nil, err
	}

	if err := p.processEmojis(form, auth.Account.ID, newStatus); err != nil {
		return nil, err
	}

	// put the new status in the database, generating an ID for it in the process
	if err := p.db.Put(newStatus); err != nil {
		return nil, err
	}

	// change the status ID of the media attachments to the new status
	for _, a := range newStatus.GTSMediaAttachments {
		a.StatusID = newStatus.ID
		a.UpdatedAt = time.Now()
		if err := p.db.UpdateByID(a.ID, a); err != nil {
			return nil, err
		}
	}

	// put the new status in the appropriate channel for async processing
	p.fromClientAPI <- gtsmodel.FromClientAPI{
		APObjectType:   newStatus.ActivityStreamsType,
		APActivityType: gtsmodel.ActivityStreamsCreate,
		GTSModel:       newStatus,
	}

	// return the frontend representation of the new status to the submitter
	return p.tc.StatusToMasto(newStatus, auth.Account, auth.Account, nil, newStatus.GTSReplyToAccount, nil)
}

func (p *processor) StatusDelete(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error) {
	l := p.log.WithField("func", "StatusDelete")
	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		return nil, fmt.Errorf("error fetching status %s: %s", targetStatusID, err)
	}

	if targetStatus.AccountID != authed.Account.ID {
		return nil, errors.New("status doesn't belong to requesting account")
	}

	l.Trace("going to get relevant accounts")
	relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(targetStatus)
	if err != nil {
		return nil, fmt.Errorf("error fetching related accounts for status %s: %s", targetStatusID, err)
	}

	var boostOfStatus *gtsmodel.Status
	if targetStatus.BoostOfID != "" {
		boostOfStatus = &gtsmodel.Status{}
		if err := p.db.GetByID(targetStatus.BoostOfID, boostOfStatus); err != nil {
			return nil, fmt.Errorf("error fetching boosted status %s: %s", targetStatus.BoostOfID, err)
		}
	}

	mastoStatus, err := p.tc.StatusToMasto(targetStatus, authed.Account, authed.Account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, boostOfStatus)
	if err != nil {
		return nil, fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err)
	}

	if err := p.db.DeleteByID(targetStatus.ID, targetStatus); err != nil {
		return nil, fmt.Errorf("error deleting status from the database: %s", err)
	}

	return mastoStatus, nil
}

func (p *processor) StatusFave(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error) {
	l := p.log.WithField("func", "StatusFave")
	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		return nil, fmt.Errorf("error fetching status %s: %s", targetStatusID, err)
	}

	l.Tracef("going to search for target account %s", targetStatus.AccountID)
	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(targetStatus.AccountID, targetAccount); err != nil {
		return nil, fmt.Errorf("error fetching target account %s: %s", targetStatus.AccountID, err)
	}

	l.Trace("going to get relevant accounts")
	relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(targetStatus)
	if err != nil {
		return nil, fmt.Errorf("error fetching related accounts for status %s: %s", targetStatusID, err)
	}

	var boostOfStatus *gtsmodel.Status
	if targetStatus.BoostOfID != "" {
		boostOfStatus = &gtsmodel.Status{}
		if err := p.db.GetByID(targetStatus.BoostOfID, boostOfStatus); err != nil {
			return nil, fmt.Errorf("error fetching boosted status %s: %s", targetStatus.BoostOfID, err)
		}
	}

	l.Trace("going to see if status is visible")
	visible, err := p.db.StatusVisible(targetStatus, targetAccount, authed.Account, relevantAccounts) // requestingAccount might well be nil here, but StatusVisible knows how to take care of that
	if err != nil {
		return nil, fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err)
	}

	if !visible {
		return nil, errors.New("status is not visible")
	}

	// is the status faveable?
	if targetStatus.VisibilityAdvanced != nil {
		if !targetStatus.VisibilityAdvanced.Likeable {
			return nil, errors.New("status is not faveable")
		}
	}

	// first check if the status is already faved, if so we don't need to do anything
	newFave := true
	gtsFave := &gtsmodel.Status{}
	if err := p.db.GetWhere([]db.Where{{Key: "status_id", Value: targetStatus.ID}, {Key: "account_id", Value: authed.Account.ID}}, gtsFave); err == nil {
		// we already have a fave for this status
		newFave = false
	}

	if newFave {
		thisFaveID := uuid.NewString()

		// we need to create a new fave in the database
		gtsFave := &gtsmodel.StatusFave{
			ID:               thisFaveID,
			AccountID:        authed.Account.ID,
			TargetAccountID:  targetAccount.ID,
			StatusID:         targetStatus.ID,
			URI:              util.GenerateURIForLike(authed.Account.Username, p.config.Protocol, p.config.Host, thisFaveID),
			GTSStatus:        targetStatus,
			GTSTargetAccount: targetAccount,
			GTSFavingAccount: authed.Account,
		}

		if err := p.db.Put(gtsFave); err != nil {
			return nil, err
		}

		// send the new fave through the processor channel for federation etc
		p.fromClientAPI <- gtsmodel.FromClientAPI{
			APObjectType:   gtsmodel.ActivityStreamsLike,
			APActivityType: gtsmodel.ActivityStreamsCreate,
			GTSModel:       gtsFave,
			OriginAccount:  authed.Account,
			TargetAccount:  targetAccount,
		}
	}

	// return the mastodon representation of the target status
	mastoStatus, err := p.tc.StatusToMasto(targetStatus, targetAccount, authed.Account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, boostOfStatus)
	if err != nil {
		return nil, fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err)
	}

	return mastoStatus, nil
}

func (p *processor) StatusBoost(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, ErrorWithCode) {
	l := p.log.WithField("func", "StatusBoost")

	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("error fetching status %s: %s", targetStatusID, err))
	}

	l.Tracef("going to search for target account %s", targetStatus.AccountID)
	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(targetStatus.AccountID, targetAccount); err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("error fetching target account %s: %s", targetStatus.AccountID, err))
	}

	l.Trace("going to get relevant accounts")
	relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(targetStatus)
	if err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("error fetching related accounts for status %s: %s", targetStatusID, err))
	}

	l.Trace("going to see if status is visible")
	visible, err := p.db.StatusVisible(targetStatus, targetAccount, authed.Account, relevantAccounts) // requestingAccount might well be nil here, but StatusVisible knows how to take care of that
	if err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err))
	}

	if !visible {
		return nil, NewErrorNotFound(errors.New("status is not visible"))
	}

	if targetStatus.VisibilityAdvanced != nil {
		if !targetStatus.VisibilityAdvanced.Boostable {
			return nil, NewErrorForbidden(errors.New("status is not boostable"))
		}
	}

	// it's visible! it's boostable! so let's boost the FUCK out of it
	// first we create a new status and add some basic info to it -- this will be the wrapper for the boosted status

	// the wrapper won't use the same ID as the boosted status so we generate some new UUIDs
	uris := util.GenerateURIsForAccount(authed.Account.Username, p.config.Protocol, p.config.Host)
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
		Content:             util.HTMLFormat(targetStatus.Content),
		ContentWarning:      targetStatus.ContentWarning,
		ActivityStreamsType: targetStatus.ActivityStreamsType,
		Sensitive:           targetStatus.Sensitive,
		Language:            targetStatus.Language,
		Text:                targetStatus.Text,
		BoostOfID:           targetStatus.ID,
		Visibility:          targetStatus.Visibility,
		VisibilityAdvanced:  targetStatus.VisibilityAdvanced,

		// attach these here for convenience -- the boosted status/account won't go in the DB
		// but they're needed in the processor and for the frontend. Since we have them, we can
		// attach them so we don't need to fetch them again later (save some DB calls)
		GTSBoostedStatus:  targetStatus,
		GTSBoostedAccount: targetAccount,
	}

	// put the boost in the database
	if err := p.db.Put(boostWrapperStatus); err != nil {
		return nil, NewErrorInternalError(err)
	}

	// return the frontend representation of the new status to the submitter
	mastoStatus, err := p.tc.StatusToMasto(boostWrapperStatus, authed.Account, authed.Account, targetAccount, nil, targetStatus)
	if err != nil {
		return nil, NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err))
	}

	return mastoStatus, nil
}

func (p *processor) StatusFavedBy(authed *oauth.Auth, targetStatusID string) ([]*apimodel.Account, error) {
	l := p.log.WithField("func", "StatusFavedBy")

	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		return nil, fmt.Errorf("error fetching status %s: %s", targetStatusID, err)
	}

	l.Tracef("going to search for target account %s", targetStatus.AccountID)
	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(targetStatus.AccountID, targetAccount); err != nil {
		return nil, fmt.Errorf("error fetching target account %s: %s", targetStatus.AccountID, err)
	}

	l.Trace("going to get relevant accounts")
	relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(targetStatus)
	if err != nil {
		return nil, fmt.Errorf("error fetching related accounts for status %s: %s", targetStatusID, err)
	}

	l.Trace("going to see if status is visible")
	visible, err := p.db.StatusVisible(targetStatus, targetAccount, authed.Account, relevantAccounts) // requestingAccount might well be nil here, but StatusVisible knows how to take care of that
	if err != nil {
		return nil, fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err)
	}

	if !visible {
		return nil, errors.New("status is not visible")
	}

	// get ALL accounts that faved a status -- doesn't take account of blocks and mutes and stuff
	favingAccounts, err := p.db.WhoFavedStatus(targetStatus)
	if err != nil {
		return nil, fmt.Errorf("error seeing who faved status: %s", err)
	}

	// filter the list so the user doesn't see accounts they blocked or which blocked them
	filteredAccounts := []*gtsmodel.Account{}
	for _, acc := range favingAccounts {
		blocked, err := p.db.Blocked(authed.Account.ID, acc.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking blocks: %s", err)
		}
		if !blocked {
			filteredAccounts = append(filteredAccounts, acc)
		}
	}

	// TODO: filter other things here? suspended? muted? silenced?

	// now we can return the masto representation of those accounts
	mastoAccounts := []*apimodel.Account{}
	for _, acc := range filteredAccounts {
		mastoAccount, err := p.tc.AccountToMastoPublic(acc)
		if err != nil {
			return nil, fmt.Errorf("error converting account to api model: %s", err)
		}
		mastoAccounts = append(mastoAccounts, mastoAccount)
	}

	return mastoAccounts, nil
}

func (p *processor) StatusGet(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error) {
	l := p.log.WithField("func", "StatusGet")

	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		return nil, fmt.Errorf("error fetching status %s: %s", targetStatusID, err)
	}

	l.Tracef("going to search for target account %s", targetStatus.AccountID)
	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(targetStatus.AccountID, targetAccount); err != nil {
		return nil, fmt.Errorf("error fetching target account %s: %s", targetStatus.AccountID, err)
	}

	l.Trace("going to get relevant accounts")
	relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(targetStatus)
	if err != nil {
		return nil, fmt.Errorf("error fetching related accounts for status %s: %s", targetStatusID, err)
	}

	l.Trace("going to see if status is visible")
	visible, err := p.db.StatusVisible(targetStatus, targetAccount, authed.Account, relevantAccounts) // requestingAccount might well be nil here, but StatusVisible knows how to take care of that
	if err != nil {
		return nil, fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err)
	}

	if !visible {
		return nil, errors.New("status is not visible")
	}

	var boostOfStatus *gtsmodel.Status
	if targetStatus.BoostOfID != "" {
		boostOfStatus = &gtsmodel.Status{}
		if err := p.db.GetByID(targetStatus.BoostOfID, boostOfStatus); err != nil {
			return nil, fmt.Errorf("error fetching boosted status %s: %s", targetStatus.BoostOfID, err)
		}
	}

	mastoStatus, err := p.tc.StatusToMasto(targetStatus, targetAccount, authed.Account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, boostOfStatus)
	if err != nil {
		return nil, fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err)
	}

	return mastoStatus, nil

}

func (p *processor) StatusUnfave(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error) {
	l := p.log.WithField("func", "StatusUnfave")
	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		return nil, fmt.Errorf("error fetching status %s: %s", targetStatusID, err)
	}

	l.Tracef("going to search for target account %s", targetStatus.AccountID)
	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(targetStatus.AccountID, targetAccount); err != nil {
		return nil, fmt.Errorf("error fetching target account %s: %s", targetStatus.AccountID, err)
	}

	l.Trace("going to get relevant accounts")
	relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(targetStatus)
	if err != nil {
		return nil, fmt.Errorf("error fetching related accounts for status %s: %s", targetStatusID, err)
	}

	l.Trace("going to see if status is visible")
	visible, err := p.db.StatusVisible(targetStatus, targetAccount, authed.Account, relevantAccounts) // requestingAccount might well be nil here, but StatusVisible knows how to take care of that
	if err != nil {
		return nil, fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err)
	}

	if !visible {
		return nil, errors.New("status is not visible")
	}

	// is the status faveable?
	if targetStatus.VisibilityAdvanced != nil {
		if !targetStatus.VisibilityAdvanced.Likeable {
			return nil, errors.New("status is not faveable")
		}
	}

	// it's visible! it's faveable! so let's unfave the FUCK out of it
	_, err = p.db.UnfaveStatus(targetStatus, authed.Account.ID)
	if err != nil {
		return nil, fmt.Errorf("error unfaveing status: %s", err)
	}

	var boostOfStatus *gtsmodel.Status
	if targetStatus.BoostOfID != "" {
		boostOfStatus = &gtsmodel.Status{}
		if err := p.db.GetByID(targetStatus.BoostOfID, boostOfStatus); err != nil {
			return nil, fmt.Errorf("error fetching boosted status %s: %s", targetStatus.BoostOfID, err)
		}
	}

	mastoStatus, err := p.tc.StatusToMasto(targetStatus, targetAccount, authed.Account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, boostOfStatus)
	if err != nil {
		return nil, fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err)
	}

	return mastoStatus, nil
}

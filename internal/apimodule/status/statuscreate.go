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
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/distributor"
	mastotypes "github.com/superseriousbusiness/gotosocial/internal/mastotypes/mastomodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type advancedStatusCreateForm struct {
	mastotypes.StatusCreateRequest
	advancedVisibilityFlagsForm
}

type advancedVisibilityFlagsForm struct {
	// The gotosocial visibility model
	VisibilityAdvanced *gtsmodel.Visibility `form:"visibility_advanced"`
	// This status will be federated beyond the local timeline(s)
	Federated *bool `form:"federated"`
	// This status can be boosted/reblogged
	Boostable *bool `form:"boostable"`
	// This status can be replied to
	Replyable *bool `form:"replyable"`
	// This status can be liked/faved
	Likeable *bool `form:"likeable"`
}

func (m *statusModule) statusCreatePOSTHandler(c *gin.Context) {
	l := m.log.WithField("func", "statusCreatePOSTHandler")
	authed, err := oauth.MustAuth(c, true, true, true, true) // posting a status is serious business so we want *everything*
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// First check this user/account is permitted to post new statuses.
	// There's no point continuing otherwise.
	if authed.User.Disabled || !authed.User.Approved || !authed.Account.SuspendedAt.IsZero() {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled, not yet approved, or suspended"})
		return
	}

	// extract the status create form from the request context
	l.Tracef("parsing request form: %s", c.Request.Form)
	form := &advancedStatusCreateForm{}
	if err := c.ShouldBind(form); err != nil || form == nil {
		l.Debugf("could not parse form from request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing one or more required form values"})
		return
	}

	// Give the fields on the request form a first pass to make sure the request is superficially valid.
	l.Tracef("validating form %+v", form)
	if err := validateCreateStatus(form, m.config.StatusesConfig); err != nil {
		l.Debugf("error validating form: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// At this point we know the account is permitted to post, and we know the request form
	// is valid (at least according to the API specifications and the instance configuration).
	// So now we can start digging a bit deeper into the form and building up the new status from it.

	// first we create a new status and add some basic info to it
	uris := util.GenerateURIs(authed.Account.Username, m.config.Protocol, m.config.Host)
	thisStatusID := uuid.NewString()
	thisStatusURI := fmt.Sprintf("%s/%s", uris.StatusesURI, thisStatusID)
	thisStatusURL := fmt.Sprintf("%s/%s", uris.StatusesURL, thisStatusID)
	newStatus := &gtsmodel.Status{
		ID:                  thisStatusID,
		URI:                 thisStatusURI,
		URL:                 thisStatusURL,
		Content:             util.HTMLFormat(form.Status),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		Local:               true,
		AccountID:           authed.Account.ID,
		ContentWarning:      form.SpoilerText,
		ActivityStreamsType: gtsmodel.ActivityStreamsNote,
		Sensitive:           form.Sensitive,
		Language:            form.Language,
	}

	// check if replyToID is ok
	if err := m.parseReplyToID(form, authed.Account.ID, newStatus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if mediaIDs are ok
	if err := m.parseMediaIDs(form, authed.Account.ID, newStatus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if visibility settings are ok
	if err := parseVisibility(form, authed.Account.Privacy, newStatus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// handle language settings
	if err := parseLanguage(form, authed.Account.Language, newStatus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// convert mentions to *gtsmodel.Mention
	menchies, err := m.db.MentionStringsToMentions(util.DeriveMentions(form.Status), authed.Account.ID, thisStatusID)
	if err != nil {
		l.Debugf("error generating mentions from status: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error generating mentions from status"})
		return
	}
	for _, menchie := range menchies {
		if err := m.db.Put(menchie); err != nil {
			l.Debugf("error putting mentions in db: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error while generating mentions from status"})
			return
		}
	}
	newStatus.Mentions = menchies

	// convert tags to *gtsmodel.Tag
	tags, err := m.db.TagStringsToTags(util.DeriveHashtags(form.Status), authed.Account.ID, thisStatusID)
	if err != nil {
		l.Debugf("error generating hashtags from status: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error generating hashtags from status"})
		return
	}
	newStatus.Tags = tags

	// convert emojis to *gtsmodel.Emoji
	emojis, err := m.db.EmojiStringsToEmojis(util.DeriveEmojis(form.Status), authed.Account.ID, thisStatusID)
	if err != nil {
		l.Debugf("error generating emojis from status: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error generating emojis from status"})
		return
	}
	newStatus.Emojis = emojis

	// put the new status in the database
	if err := m.db.Put(newStatus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// pass to the distributor to take care of side effects -- federation, mentions, updating metadata, etc, etc
	m.distributor.FromClientAPI() <- distributor.FromClientAPI{
		APObjectType:   gtsmodel.ActivityStreamsNote,
		APActivityType: gtsmodel.ActivityStreamsCreate,
		Activity:       newStatus,
	}

	// now we need to build up the mastodon-style status object to return to the submitter

	mastoVis := util.ParseMastoVisFromGTSVis(newStatus.Visibility)

	mastoAccount, err := m.mastoConverter.AccountToMastoPublic(authed.Account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mastoAttachments := []mastotypes.Attachment{}
	for _, a := range newStatus.Attachments {
		ma, err := m.mastoConverter.AttachmentToMasto(a)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		mastoAttachments = append(mastoAttachments, ma)
	}

	mastoMentions := []mastotypes.Mention{}
	for _, gtsm := range newStatus.Mentions {
		mm, err := m.mastoConverter.MentionToMasto(gtsm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		mastoMentions = append(mastoMentions, mm)
	}

	mastoApplication, err := m.mastoConverter.AppToMastoPublic(authed.Application)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mastoStatus := &mastotypes.Status{
		ID:                 newStatus.ID,
		CreatedAt:          newStatus.CreatedAt.Format(time.RFC3339),
		InReplyToID:        newStatus.InReplyToID,
		InReplyToAccountID: newStatus.InReplyToAccountID,
		Sensitive:          newStatus.Sensitive,
		SpoilerText:        newStatus.ContentWarning,
		Visibility:         mastoVis,
		Language:           newStatus.Language,
		URI:                newStatus.URI,
		URL:                newStatus.URL,
		Content:            newStatus.Content,
		Application:        mastoApplication,
		Account:            mastoAccount,
		MediaAttachments:   mastoAttachments,
		Mentions:           mastoMentions,
		Text:               form.Status,
	}
	c.JSON(http.StatusOK, mastoStatus)
}

func validateCreateStatus(form *advancedStatusCreateForm, config *config.StatusesConfig) error {
	// validate that, structurally, we have a valid status/post
	if form.Status == "" && form.MediaIDs == nil && form.Poll == nil {
		return errors.New("no status, media, or poll provided")
	}

	if form.MediaIDs != nil && form.Poll != nil {
		return errors.New("can't post media + poll in same status")
	}

	// validate status
	if form.Status != "" {
		if len(form.Status) > config.MaxChars {
			return fmt.Errorf("status too long, %d characters provided but limit is %d", len(form.Status), config.MaxChars)
		}
	}

	// validate media attachments
	if len(form.MediaIDs) > config.MaxMediaFiles {
		return fmt.Errorf("too many media files attached to status, %d attached but limit is %d", len(form.MediaIDs), config.MaxMediaFiles)
	}

	// validate poll
	if form.Poll != nil {
		if form.Poll.Options == nil {
			return errors.New("poll with no options")
		}
		if len(form.Poll.Options) > config.PollMaxOptions {
			return fmt.Errorf("too many poll options provided, %d provided but limit is %d", len(form.Poll.Options), config.PollMaxOptions)
		}
		for _, p := range form.Poll.Options {
			if len(p) > config.PollOptionMaxChars {
				return fmt.Errorf("poll option too long, %d characters provided but limit is %d", len(p), config.PollOptionMaxChars)
			}
		}
	}

	// validate spoiler text/cw
	if form.SpoilerText != "" {
		if len(form.SpoilerText) > config.CWMaxChars {
			return fmt.Errorf("content-warning/spoilertext too long, %d characters provided but limit is %d", len(form.SpoilerText), config.CWMaxChars)
		}
	}

	// validate post language
	if form.Language != "" {
		if err := util.ValidateLanguage(form.Language); err != nil {
			return err
		}
	}

	return nil
}

func parseVisibility(form *advancedStatusCreateForm, accountDefaultVis gtsmodel.Visibility, status *gtsmodel.Status) error {
	// by default all flags are set to true
	gtsAdvancedVis := &gtsmodel.VisibilityAdvanced{
		Federated: true,
		Boostable: true,
		Replyable: true,
		Likeable:  true,
	}

	var gtsBasicVis gtsmodel.Visibility
	// Advanced takes priority if it's set.
	// If it's not set, take whatever masto visibility is set.
	// If *that's* not set either, then just take the account default.
	if form.VisibilityAdvanced != nil {
		gtsBasicVis = *form.VisibilityAdvanced
	} else if form.Visibility != "" {
		gtsBasicVis = util.ParseGTSVisFromMastoVis(form.Visibility)
	} else {
		gtsBasicVis = accountDefaultVis
	}

	switch gtsBasicVis {
	case gtsmodel.VisibilityPublic:
		// for public, there's no need to change any of the advanced flags from true regardless of what the user filled out
		break
	case gtsmodel.VisibilityUnlocked:
		// for unlocked the user can set any combination of flags they like so look at them all to see if they're set and then apply them
		if form.Federated != nil {
			gtsAdvancedVis.Federated = *form.Federated
		}

		if form.Boostable != nil {
			gtsAdvancedVis.Boostable = *form.Boostable
		}

		if form.Replyable != nil {
			gtsAdvancedVis.Replyable = *form.Replyable
		}

		if form.Likeable != nil {
			gtsAdvancedVis.Likeable = *form.Likeable
		}

	case gtsmodel.VisibilityFollowersOnly, gtsmodel.VisibilityMutualsOnly:
		// for followers or mutuals only, boostable will *always* be false, but the other fields can be set so check and apply them
		gtsAdvancedVis.Boostable = false

		if form.Federated != nil {
			gtsAdvancedVis.Federated = *form.Federated
		}

		if form.Replyable != nil {
			gtsAdvancedVis.Replyable = *form.Replyable
		}

		if form.Likeable != nil {
			gtsAdvancedVis.Likeable = *form.Likeable
		}

	case gtsmodel.VisibilityDirect:
		// direct is pretty easy: there's only one possible setting so return it
		gtsAdvancedVis.Federated = true
		gtsAdvancedVis.Boostable = false
		gtsAdvancedVis.Federated = true
		gtsAdvancedVis.Likeable = true
	}

	status.Visibility = gtsBasicVis
	status.VisibilityAdvanced = gtsAdvancedVis
	return nil
}

func (m *statusModule) parseReplyToID(form *advancedStatusCreateForm, thisAccountID string, status *gtsmodel.Status) error {
	if form.InReplyToID == "" {
		return nil
	}

	// If this status is a reply to another status, we need to do a bit of work to establish whether or not this status can be posted:
	//
	// 1. Does the replied status exist in the database?
	// 2. Is the replied status marked as replyable?
	// 3. Does a block exist between either the current account or the account that posted the status it's replying to?
	//
	// If this is all OK, then we fetch the repliedStatus and the repliedAccount for later processing.
	repliedStatus := &gtsmodel.Status{}
	repliedAccount := &gtsmodel.Account{}
	// check replied status exists + is replyable
	if err := m.db.GetByID(form.InReplyToID, repliedStatus); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return fmt.Errorf("status with id %s not replyable because it doesn't exist", form.InReplyToID)
		} else {
			return fmt.Errorf("status with id %s not replyable: %s", form.InReplyToID, err)
		}
	}

	if !repliedStatus.VisibilityAdvanced.Replyable {
		return fmt.Errorf("status with id %s is marked as not replyable", form.InReplyToID)
	}

	// check replied account is known to us
	if err := m.db.GetByID(repliedStatus.AccountID, repliedAccount); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return fmt.Errorf("status with id %s not replyable because account id %s is not known", form.InReplyToID, repliedStatus.AccountID)
		} else {
			return fmt.Errorf("status with id %s not replyable: %s", form.InReplyToID, err)
		}
	}
	// check if a block exists
	if blocked, err := m.db.Blocked(thisAccountID, repliedAccount.ID); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			return fmt.Errorf("status with id %s not replyable: %s", form.InReplyToID, err)
		}
	} else if blocked {
		return fmt.Errorf("status with id %s not replyable", form.InReplyToID)
	}
	status.InReplyToID = repliedStatus.ID
	status.InReplyToAccountID = repliedAccount.ID

	return nil
}

func (m *statusModule) parseMediaIDs(form *advancedStatusCreateForm, thisAccountID string, status *gtsmodel.Status) error {
	if form.MediaIDs == nil {
		return nil
	}

	attachments := []*gtsmodel.MediaAttachment{}
	for _, mediaID := range form.MediaIDs {
		// check these attachments exist
		a := &gtsmodel.MediaAttachment{}
		if err := m.db.GetByID(mediaID, a); err != nil {
			return fmt.Errorf("invalid media type or media not found for media id %s", mediaID)
		}
		// check they belong to the requesting account id
		if a.AccountID != thisAccountID {
			return fmt.Errorf("media with id %s does not belong to account %s", mediaID, thisAccountID)
		}
		attachments = append(attachments, a)
	}
	status.Attachments = attachments
	return nil
}

func parseLanguage(form *advancedStatusCreateForm, accountDefaultLanguage string, status *gtsmodel.Status) error {
	if form.Language != "" {
		status.Language = form.Language
	} else {
		status.Language = accountDefaultLanguage
	}
	if status.Language == "" {
		return errors.New("no language given either in status create form or account default")
	}
	return nil
}

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
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/distributor"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/pkg/mastotypes"
)

type advancedStatusCreateForm struct {
	mastotypes.StatusCreateRequest
	AdvancedVisibility *advancedVisibilityFlagsForm `form:"visibility_advanced"`
}

type advancedVisibilityFlagsForm struct {
	// The gotosocial visibility model
	Visibility *model.Visibility
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
	l.Trace("parsing request form")
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
	newStatus := &model.Status{
		ID:                  thisStatusID,
		URI:                 thisStatusURI,
		URL:                 thisStatusURL,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		Local:               true,
		AccountID:           authed.Account.ID,
		ContentWarning:      form.SpoilerText,
		ActivityStreamsType: model.ActivityStreamsNote,
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

	// convert mentions to *model.Mention
	menchies, err := m.db.MentionStringsToMentions(util.DeriveMentions(form.Status), authed.Account.ID, thisStatusID)
	if err != nil {
		l.Debugf("error generating mentions from status: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error generating mentions from status"})
		return
	}
	newStatus.Mentions = menchies

	// convert tags to *model.Tag
	tags, err := m.db.TagStringsToTags(util.DeriveHashtags(form.Status), authed.Account.ID, thisStatusID)
	if err != nil {
		l.Debugf("error generating hashtags from status: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error generating hashtags from status"})
		return
	}
	newStatus.Tags = tags

	// convert emojis to *model.Emoji
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
		APObjectType:   model.ActivityStreamsNote,
		APActivityType: model.ActivityStreamsCreate,
		Activity:       newStatus,
	}

	// return populated status to submitter
	// mastoStatus := &mastotypes.Status{
	// 	ID:                 newStatus.ID,
	// 	CreatedAt:          time.Now().Format(time.RFC3339),
	// 	InReplyToID:        newStatus.InReplyToID,
	// 	// InReplyToAccountID: newStatus.,
	// }
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

func parseVisibility(form *advancedStatusCreateForm, accountDefaultVis model.Visibility, status *model.Status) error {
	// by default all flags are set to true
	gtsAdvancedVis := &model.VisibilityAdvanced{
		Federated: true,
		Boostable: true,
		Replyable: true,
		Likeable:  true,
	}

	var gtsBasicVis model.Visibility
	// Advanced takes priority if it's set.
	// If it's not set, take whatever masto visibility is set.
	// If *that's* not set either, then just take the account default.
	if form.AdvancedVisibility != nil && form.AdvancedVisibility.Visibility != nil {
		gtsBasicVis = *form.AdvancedVisibility.Visibility
	} else if form.Visibility != "" {
		gtsBasicVis = util.ParseGTSVisFromMastoVis(form.Visibility)
	} else {
		gtsBasicVis = accountDefaultVis
	}

	switch gtsBasicVis {
	case model.VisibilityPublic:
		// for public, there's no need to change any of the advanced flags from true regardless of what the user filled out
		break
	case model.VisibilityUnlocked:
		// for unlocked the user can set any combination of flags they like so look at them all to see if they're set and then apply them
		if form.AdvancedVisibility != nil {
			if form.AdvancedVisibility.Federated != nil {
				gtsAdvancedVis.Federated = *form.AdvancedVisibility.Federated
			}

			if form.AdvancedVisibility.Boostable != nil {
				gtsAdvancedVis.Boostable = *form.AdvancedVisibility.Boostable
			}

			if form.AdvancedVisibility.Replyable != nil {
				gtsAdvancedVis.Replyable = *form.AdvancedVisibility.Replyable
			}

			if form.AdvancedVisibility.Likeable != nil {
				gtsAdvancedVis.Likeable = *form.AdvancedVisibility.Likeable
			}
		}
	case model.VisibilityFollowersOnly, model.VisibilityMutualsOnly:
		// for followers or mutuals only, boostable will *always* be false, but the other fields can be set so check and apply them
		gtsAdvancedVis.Boostable = false

		if form.AdvancedVisibility != nil {
			if form.AdvancedVisibility.Federated != nil {
				gtsAdvancedVis.Federated = *form.AdvancedVisibility.Federated
			}

			if form.AdvancedVisibility.Replyable != nil {
				gtsAdvancedVis.Replyable = *form.AdvancedVisibility.Replyable
			}

			if form.AdvancedVisibility.Likeable != nil {
				gtsAdvancedVis.Likeable = *form.AdvancedVisibility.Likeable
			}
		}
	case model.VisibilityDirect:
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

func (m *statusModule) parseReplyToID(form *advancedStatusCreateForm, thisAccountID string, status *model.Status) error {
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
	repliedStatus := &model.Status{}
	repliedAccount := &model.Account{}
	// check replied status exists + is replyable
	if err := m.db.GetByID(form.InReplyToID, repliedStatus); err != nil || !repliedStatus.VisibilityAdvanced.Replyable {
		return fmt.Errorf("status with id %s not replyable: %s", form.InReplyToID, err)
	}
	// check replied account is known to us
	if err := m.db.GetByID(repliedStatus.AccountID, repliedAccount); err != nil {
		return fmt.Errorf("status with id %s not replyable: %s", form.InReplyToID, err)
	}
	// check if a block exists
	if blocked, err := m.db.Blocked(thisAccountID, repliedAccount.ID); err != nil || blocked {
		return fmt.Errorf("status with id %s not replyable: %s", form.InReplyToID, err)
	}
	status.InReplyToID = repliedStatus.ID

	return nil
}

func (m *statusModule) parseMediaIDs(form *advancedStatusCreateForm, thisAccountID string, status *model.Status) error {
	if form.MediaIDs == nil {
		return nil
	}

	attachments := []*model.MediaAttachment{}
	for _, mediaID := range form.MediaIDs {
		// check these attachments exist
		a := &model.MediaAttachment{}
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

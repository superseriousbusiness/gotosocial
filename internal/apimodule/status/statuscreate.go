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
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
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

	// check this user/account is permitted to post new statuses
	if authed.User.Disabled || !authed.User.Approved || !authed.Account.SuspendedAt.IsZero() {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled, not yet approved, or suspended"})
		return
	}

	l.Trace("parsing request form")
	form := &advancedStatusCreateForm{}
	if err := c.ShouldBind(form); err != nil || form == nil {
		l.Debugf("could not parse form from request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing one or more required form values"})
		return
	}

	l.Tracef("validating form %+v", form)
	if err := validateCreateStatus(form, m.config.StatusesConfig, authed.Account.ID, m.db); err != nil {
		l.Debugf("error validating form: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// here we check if any advanced visibility flags have been set and fiddle with them if so
	l.Trace("deriving visibility")
	basicVis, advancedVis, err := deriveTotalVisibility(form.Visibility, form.AdvancedVisibility, authed.Account.Privacy)

	clientIP := c.ClientIP()
	l.Tracef("attempting to parse client ip address %s", clientIP)
	signUpIP := net.ParseIP(clientIP)
	if signUpIP == nil {
		l.Debugf("error validating client ip address %s", clientIP)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ip address could not be parsed from request"})
		return
	}

	uris := util.GenerateURIs(authed.Account.Username, m.config.Protocol, m.config.Host)
	thisStatusID := uuid.NewString()
	thisStatusURI := fmt.Sprintf("%s/%s", uris.StatusesURI, thisStatusID)
	thisStatusURL := fmt.Sprintf("%s/%s", uris.StatusesURL, thisStatusID)
	newStatus := &model.Status{
		ID:                  thisStatusID,
		URI:                 thisStatusURI,
		URL:                 thisStatusURL,
		Content:             util.HTMLFormat(form.Status),
		Local:               true, // will always be true if this status is being created through the client API, since only local users can do that
		AccountID:           authed.Account.ID,
		InReplyToID:         form.InReplyToID,
		ContentWarning:      form.SpoilerText,
		Visibility:          basicVis,
		VisibilityAdvanced:  *advancedVis,
		ActivityStreamsType: model.ActivityStreamsNote,
	}

	menchies, err := m.db.MentionStringsToMentions(util.DeriveMentions(form.Status), authed.Account.ID, thisStatusID)
	if err != nil {
		l.Debugf("error generating mentions from status: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error generating mentions from status"})
		return
	}

	tags, err := m.db.TagStringsToTags(util.DeriveHashtags(form.Status), authed.Account.ID, thisStatusID)
	if err != nil {
		l.Debugf("error generating hashtags from status: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error generating hashtags from status"})
		return
	}

	emojis, err := m.db.EmojiStringsToEmojis(util.DeriveEmojis(form.Status), authed.Account.ID, thisStatusID)
	if err != nil {
		l.Debugf("error generating emojis from status: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error generating emojis from status"})
		return
	}

	newStatus.Mentions = menchies
	newStatus.Tags = tags
	newStatus.Emojis = emojis

	// take care of side effects -- federation, mentions, updating metadata, etc, etc
	m.distributor.FromClientAPI() <- distributor.FromClientAPI{
		APObjectType: model.ActivityStreamsNote,
		APActivityType: model.ActivityStreamsCreate,
		Activity: newStatus,
	}

	// return populated status to submitter
	

}

func validateCreateStatus(form *advancedStatusCreateForm, config *config.StatusesConfig, accountID string, db db.DB) error {
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

	for _, m := range form.MediaIDs {
		// check these attachments exist
		a := &model.MediaAttachment{}
		if err := db.GetByID(m, a); err != nil {
			return fmt.Errorf("invalid media type or media not found for media id %s: %s", m, err)
		}
		// check they belong to the requesting account id
		if a.AccountID != accountID {
			return fmt.Errorf("media attachment %s does not belong to account id %s", m, accountID)
		}
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

	// validate reply-to status exists and is reply-able
	if form.InReplyToID != "" {
		s := &model.Status{}
		if err := db.GetByID(form.InReplyToID, s); err != nil {
			return fmt.Errorf("status id %s cannot be retrieved from the db: %s", form.InReplyToID, err)
		}
		if !s.VisibilityAdvanced.Replyable {
			return fmt.Errorf("status with id %s is not replyable", form.InReplyToID)
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

func deriveTotalVisibility(basicVisForm mastotypes.Visibility, advancedVisForm *advancedVisibilityFlagsForm, accountDefaultVis model.Visibility) (model.Visibility, *model.VisibilityAdvanced, error) {
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
	if advancedVisForm != nil && advancedVisForm.Visibility != nil {
		gtsBasicVis = *advancedVisForm.Visibility
	} else if basicVisForm != "" {
		gtsBasicVis = util.ParseGTSVisFromMastoVis(basicVisForm)
	} else {
		gtsBasicVis = accountDefaultVis
	}

	switch gtsBasicVis {
	case model.VisibilityPublic:
		// for public, there's no need to change any of the advanced flags from true regardless of what the user filled out
		return gtsBasicVis, gtsAdvancedVis, nil
	case model.VisibilityUnlocked:
		// for unlocked the user can set any combination of flags they like so look at them all to see if they're set and then apply them
		if advancedVisForm != nil {
			if advancedVisForm.Federated != nil {
				gtsAdvancedVis.Federated = *advancedVisForm.Federated
			}

			if advancedVisForm.Boostable != nil {
				gtsAdvancedVis.Boostable = *advancedVisForm.Boostable
			}

			if advancedVisForm.Replyable != nil {
				gtsAdvancedVis.Replyable = *advancedVisForm.Replyable
			}

			if advancedVisForm.Likeable != nil {
				gtsAdvancedVis.Likeable = *advancedVisForm.Likeable
			}
		}
		return gtsBasicVis, gtsAdvancedVis, nil
	case model.VisibilityFollowersOnly, model.VisibilityMutualsOnly:
		// for followers or mutuals only, boostable will *always* be false, but the other fields can be set so check and apply them
		gtsAdvancedVis.Boostable = false

		if advancedVisForm != nil {
			if advancedVisForm.Federated != nil {
				gtsAdvancedVis.Federated = *advancedVisForm.Federated
			}

			if advancedVisForm.Replyable != nil {
				gtsAdvancedVis.Replyable = *advancedVisForm.Replyable
			}

			if advancedVisForm.Likeable != nil {
				gtsAdvancedVis.Likeable = *advancedVisForm.Likeable
			}
		}

		return gtsBasicVis, gtsAdvancedVis, nil
	case model.VisibilityDirect:
		// direct is pretty easy: there's only one possible setting so return it
		gtsAdvancedVis.Federated = true
		gtsAdvancedVis.Boostable = false
		gtsAdvancedVis.Federated = true
		gtsAdvancedVis.Likeable = true
		return gtsBasicVis, gtsAdvancedVis, nil
	}

	// this should never happen but just in case...
	return "", nil, errors.New("could not parse visibility")
}

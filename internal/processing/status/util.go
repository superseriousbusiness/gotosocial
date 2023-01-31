/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"context"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

func (p *processor) ProcessVisibility(ctx context.Context, form *apimodel.AdvancedStatusCreateForm, accountDefaultVis gtsmodel.Visibility, status *gtsmodel.Status) error {
	// by default all flags are set to true
	federated := true
	boostable := true
	replyable := true
	likeable := true

	// If visibility isn't set on the form, then just take the account default.
	// If that's also not set, take the default for the whole instance.
	var vis gtsmodel.Visibility
	switch {
	case form.Visibility != "":
		vis = p.tc.APIVisToVis(form.Visibility)
	case accountDefaultVis != "":
		vis = accountDefaultVis
	default:
		vis = gtsmodel.VisibilityDefault
	}

	switch vis {
	case gtsmodel.VisibilityPublic:
		// for public, there's no need to change any of the advanced flags from true regardless of what the user filled out
		break
	case gtsmodel.VisibilityUnlocked:
		// for unlocked the user can set any combination of flags they like so look at them all to see if they're set and then apply them
		if form.Federated != nil {
			federated = *form.Federated
		}

		if form.Boostable != nil {
			boostable = *form.Boostable
		}

		if form.Replyable != nil {
			replyable = *form.Replyable
		}

		if form.Likeable != nil {
			likeable = *form.Likeable
		}

	case gtsmodel.VisibilityFollowersOnly, gtsmodel.VisibilityMutualsOnly:
		// for followers or mutuals only, boostable will *always* be false, but the other fields can be set so check and apply them
		boostable = false

		if form.Federated != nil {
			federated = *form.Federated
		}

		if form.Replyable != nil {
			replyable = *form.Replyable
		}

		if form.Likeable != nil {
			likeable = *form.Likeable
		}

	case gtsmodel.VisibilityDirect:
		// direct is pretty easy: there's only one possible setting so return it
		federated = true
		boostable = false
		replyable = true
		likeable = true
	}

	status.Visibility = vis
	status.Federated = &federated
	status.Boostable = &boostable
	status.Replyable = &replyable
	status.Likeable = &likeable
	return nil
}

func (p *processor) ProcessReplyToID(ctx context.Context, form *apimodel.AdvancedStatusCreateForm, thisAccountID string, status *gtsmodel.Status) gtserror.WithCode {
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

	if err := p.db.GetByID(ctx, form.InReplyToID, repliedStatus); err != nil {
		if err == db.ErrNoEntries {
			err := fmt.Errorf("status with id %s not replyable because it doesn't exist", form.InReplyToID)
			return gtserror.NewErrorBadRequest(err, err.Error())
		}
		err := fmt.Errorf("db error fetching status with id %s: %s", form.InReplyToID, err)
		return gtserror.NewErrorInternalError(err)
	}
	if !*repliedStatus.Replyable {
		err := fmt.Errorf("status with id %s is marked as not replyable", form.InReplyToID)
		return gtserror.NewErrorForbidden(err, err.Error())
	}

	if err := p.db.GetByID(ctx, repliedStatus.AccountID, repliedAccount); err != nil {
		if err == db.ErrNoEntries {
			err := fmt.Errorf("status with id %s not replyable because account id %s is not known", form.InReplyToID, repliedStatus.AccountID)
			return gtserror.NewErrorBadRequest(err, err.Error())
		}
		err := fmt.Errorf("db error fetching account with id %s: %s", repliedStatus.AccountID, err)
		return gtserror.NewErrorInternalError(err)
	}

	if blocked, err := p.db.IsBlocked(ctx, thisAccountID, repliedAccount.ID, true); err != nil {
		err := fmt.Errorf("db error checking block: %s", err)
		return gtserror.NewErrorInternalError(err)
	} else if blocked {
		err := fmt.Errorf("status with id %s not replyable", form.InReplyToID)
		return gtserror.NewErrorNotFound(err)
	}

	status.InReplyToID = repliedStatus.ID
	status.InReplyToURI = repliedStatus.URI
	status.InReplyToAccountID = repliedAccount.ID

	return nil
}

func (p *processor) ProcessMediaIDs(ctx context.Context, form *apimodel.AdvancedStatusCreateForm, thisAccountID string, status *gtsmodel.Status) gtserror.WithCode {
	if form.MediaIDs == nil {
		return nil
	}

	attachments := []*gtsmodel.MediaAttachment{}
	attachmentIDs := []string{}
	for _, mediaID := range form.MediaIDs {
		attachment, err := p.db.GetAttachmentByID(ctx, mediaID)
		if err != nil {
			if errors.Is(err, db.ErrNoEntries) {
				err = fmt.Errorf("ProcessMediaIDs: media not found for media id %s", mediaID)
				return gtserror.NewErrorBadRequest(err, err.Error())
			}
			err = fmt.Errorf("ProcessMediaIDs: db error for media id %s", mediaID)
			return gtserror.NewErrorInternalError(err)
		}

		if attachment.AccountID != thisAccountID {
			err = fmt.Errorf("ProcessMediaIDs: media with id %s does not belong to account %s", mediaID, thisAccountID)
			return gtserror.NewErrorBadRequest(err, err.Error())
		}

		if attachment.StatusID != "" || attachment.ScheduledStatusID != "" {
			err = fmt.Errorf("ProcessMediaIDs: media with id %s is already attached to a status", mediaID)
			return gtserror.NewErrorBadRequest(err, err.Error())
		}

		minDescriptionChars := config.GetMediaDescriptionMinChars()
		if descriptionLength := len([]rune(attachment.Description)); descriptionLength < minDescriptionChars {
			err = fmt.Errorf("ProcessMediaIDs: description too short! media description of at least %d chararacters is required but %d was provided for media with id %s", minDescriptionChars, descriptionLength, mediaID)
			return gtserror.NewErrorBadRequest(err, err.Error())
		}

		attachments = append(attachments, attachment)
		attachmentIDs = append(attachmentIDs, attachment.ID)
	}

	status.Attachments = attachments
	status.AttachmentIDs = attachmentIDs
	return nil
}

func (p *processor) ProcessLanguage(ctx context.Context, form *apimodel.AdvancedStatusCreateForm, accountDefaultLanguage string, status *gtsmodel.Status) error {
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

func (p *processor) ProcessContent(ctx context.Context, form *apimodel.AdvancedStatusCreateForm, accountID string, status *gtsmodel.Status) error {
	// if there's nothing in the status at all we can just return early
	if form.Status == "" {
		status.Content = ""
		return nil
	}

	// if format wasn't specified we should try to figure out what format this user prefers
	if form.Format == "" {
		acct, err := p.db.GetAccountByID(ctx, accountID)
		if err != nil {
			return fmt.Errorf("error processing new content: couldn't retrieve account from db to check post format: %s", err)
		}

		switch acct.StatusFormat {
		case "plain":
			form.Format = apimodel.StatusFormatPlain
		case "markdown":
			form.Format = apimodel.StatusFormatMarkdown
		default:
			form.Format = apimodel.StatusFormatDefault
		}
	}

	// parse content out of the status depending on what format has been submitted
	var f text.FormatFunc
	switch form.Format {
	case apimodel.StatusFormatPlain:
		f = p.formatter.FromPlain
	case apimodel.StatusFormatMarkdown:
		f = p.formatter.FromMarkdown
	default:
		return fmt.Errorf("format %s not recognised as a valid status format", form.Format)
	}
	formatted := f(ctx, p.parseMention, accountID, status.ID, form.Status)

	// add full populated gts {mentions, tags, emojis} to the status for passing them around conveniently
	// add just their ids to the status for putting in the db
	status.Mentions = formatted.Mentions
	status.MentionIDs = make([]string, 0, len(formatted.Mentions))
	for _, gtsmention := range formatted.Mentions {
		status.MentionIDs = append(status.MentionIDs, gtsmention.ID)
	}

	status.Tags = formatted.Tags
	status.TagIDs = make([]string, 0, len(formatted.Tags))
	for _, gtstag := range formatted.Tags {
		status.TagIDs = append(status.TagIDs, gtstag.ID)
	}

	status.Emojis = formatted.Emojis
	status.EmojiIDs = make([]string, 0, len(formatted.Emojis))
	for _, gtsemoji := range formatted.Emojis {
		status.EmojiIDs = append(status.EmojiIDs, gtsemoji.ID)
	}

	spoilerformatted := p.formatter.FromPlainEmojiOnly(ctx, p.parseMention, accountID, status.ID, form.SpoilerText)
	for _, gtsemoji := range spoilerformatted.Emojis {
		status.Emojis = append(status.Emojis, gtsemoji)
		status.EmojiIDs = append(status.EmojiIDs, gtsemoji.ID)
	}

	status.Content = formatted.HTML
	return nil
}
